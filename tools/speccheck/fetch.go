package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/terrakube-io/terrakube-go/tools/speccheck/internal/speccheck"
)

const (
	defaultRepo      = "terrakube-io/terrakube"
	defaultGitHubAPI = "https://api.github.com"
	upstreamSpecDir  = "openapi-spec"
	maxSpecBytes     = 64 << 20
	tokenEnv         = "TERRAKUBE_TOKEN" //nolint:gosec // G101: name of an environment variable, not a credential
	githubTokenEnv   = "GITHUB_TOKEN"    //nolint:gosec // G101: name of an environment variable, not a credential
)

// versionPattern is the shape of a Terrakube release version. It keeps the
// version, which becomes part of the output file name, free of path elements.
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?$`)

// upstreamSpecFile matches files such as v2_27_0.yml in the upstream repo.
var upstreamSpecFile = regexp.MustCompile(`^v(\d+)_(\d+)_(\d+)\.(ya?ml|json)$`)

type fetchOptions struct {
	version   string
	output    string
	repo      string
	ref       string
	server    string
	githubAPI string
	force     bool
	root      string
}

func runFetch(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("fetch", stderr)
	var o fetchOptions
	fs.StringVar(&o.version, "version", "latest", `API version to fetch, e.g. 2.34.1; "latest" picks the newest file upstream (GitHub only)`)
	fs.StringVar(&o.output, "output", "", "output file (default: openapi/v{VERSION}.json in the repository)")
	fs.StringVar(&o.repo, "repo", defaultRepo, "GitHub repository that publishes "+upstreamSpecDir+"/v{X_Y_Z}.yml")
	fs.StringVar(&o.ref, "ref", "", "git ref to read the upstream file from (default: the repository's default branch)")
	fs.StringVar(&o.server, "server", "", "fetch from a running Terrakube API (GET {server}/doc) instead of GitHub; the token is read from $"+tokenEnv)
	fs.StringVar(&o.githubAPI, "github-api", defaultGitHubAPI, "GitHub API base URL")
	fs.BoolVar(&o.force, "force", false, "overwrite an existing output file")
	fs.StringVar(&o.root, "root", "", "repository root (default: found from the working directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return fetch(ctx, o, &http.Client{CheckRedirect: sameOriginRedirect}, stdout, stderr)
}

// sameOriginRedirect refuses redirects that change scheme or host, so a
// bearer token is never replayed to another origin or over plain http.
func sameOriginRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if req.URL.Scheme != via[0].URL.Scheme || req.URL.Host != via[0].URL.Host {
		return fmt.Errorf("refusing redirect from %s to another origin %s", via[0].URL.Redacted(), req.URL.Redacted())
	}
	return nil
}

func fetch(ctx context.Context, o fetchOptions, hc *http.Client, stdout, stderr io.Writer) error {
	var (
		raw     []byte
		version = strings.TrimPrefix(o.version, "v")
		source  string
		err     error
	)
	if o.server != "" {
		if version == "" || version == "latest" {
			return errors.New("--version is required with --server: the document served by Terrakube does not state its version")
		}
		raw, source, err = fetchFromServer(ctx, hc, o.server, os.Getenv(tokenEnv), stderr)
	} else {
		raw, version, source, err = fetchFromGitHub(ctx, hc, o, version, os.Getenv(githubTokenEnv), stderr)
	}
	if err != nil {
		return err
	}

	if !versionPattern.MatchString(version) {
		return fmt.Errorf("version %q is not of the form X.Y.Z", version)
	}
	normalized, err := normalizeSpec(raw, version)
	if err != nil {
		return fmt.Errorf("document from %s: %w", source, err)
	}

	out := o.output
	if out == "" {
		root, err := resolveRoot(o.root)
		if err != nil {
			return err
		}
		out = specFile(root, version)
	}
	if !o.force {
		if _, err := os.Stat(out); err == nil {
			return fmt.Errorf("%s exists; use --force to overwrite", out)
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return err
	}
	//nolint:gosec // G703: out is the --output the user chose or openapi/v{version}.json with a validated version
	if err := os.WriteFile(out, normalized, 0o600); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "wrote %s (version %s, from %s)\n", out, version, source)
	return nil
}

// fetchFromServer downloads the OpenAPI document a Terrakube API serves at /doc.
func fetchFromServer(ctx context.Context, hc *http.Client, server, token string, stderr io.Writer) ([]byte, string, error) {
	base, err := url.Parse(server)
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Scheme != "http") {
		return nil, "", fmt.Errorf("--server must be an http(s) URL, got %q", server)
	}
	if token == "" {
		return nil, "", fmt.Errorf("$%s is not set; the /doc endpoint requires a bearer token", tokenEnv)
	}
	if base.Scheme == "http" {
		_, _ = fmt.Fprintf(stderr, "warning: %s uses plain http; the token is sent unencrypted\n", base.Redacted())
	}
	u := base.JoinPath("doc")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	body, err := doGet(hc, req)
	if err != nil {
		return nil, "", err
	}
	return body, u.Redacted(), nil
}

// fetchFromGitHub downloads openapi-spec/v{X_Y_Z}.yml from the upstream
// repository. version "latest" selects the highest version published there.
func fetchFromGitHub(ctx context.Context, hc *http.Client, o fetchOptions, version, token string, stderr io.Writer) (body []byte, resolved, source string, err error) {
	if token == "" {
		_, _ = fmt.Fprintf(stderr, "warning: $%s is not set; GitHub allows 60 unauthenticated API requests per hour\n", githubTokenEnv)
	}
	api, err := url.Parse(o.githubAPI)
	if err != nil {
		return nil, "", "", fmt.Errorf("--github-api: %w", err)
	}
	if api.Scheme == "http" && token != "" {
		_, _ = fmt.Fprintf(stderr, "warning: %s uses plain http; $%s is sent unencrypted\n", api.Redacted(), githubTokenEnv)
	}
	files, err := listUpstreamSpecs(ctx, hc, api, o.repo, o.ref, token)
	if err != nil {
		return nil, "", "", err
	}
	if len(files) == 0 {
		return nil, "", "", fmt.Errorf("%s has no %s/v*_*_*.yml files", o.repo, upstreamSpecDir)
	}
	var name string
	if version == "" || version == "latest" {
		name = files[len(files)-1].name
		version = files[len(files)-1].version
	} else {
		i := slices.IndexFunc(files, func(f upstreamFile) bool { return f.version == version })
		if i < 0 {
			var have []string
			for _, f := range files {
				have = append(have, f.version)
			}
			return nil, "", "", fmt.Errorf("%s publishes no document for %s (available: %s); fetch it from a running server with --server",
				o.repo, version, strings.Join(have, ", "))
		}
		name = files[i].name
	}

	u := api.JoinPath("repos", o.repo, "contents", upstreamSpecDir, name)
	if o.ref != "" {
		u.RawQuery = url.Values{"ref": {o.ref}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	setGitHubAuth(req, token)
	body, err = doGet(hc, req)
	if err != nil {
		return nil, "", "", err
	}
	return body, version, o.repo + "/" + upstreamSpecDir + "/" + name, nil
}

type upstreamFile struct {
	name    string
	version string
	key     [3]int
}

func listUpstreamSpecs(ctx context.Context, hc *http.Client, api *url.URL, repo, ref, token string) ([]upstreamFile, error) {
	u := api.JoinPath("repos", repo, "contents", upstreamSpecDir)
	if ref != "" {
		u.RawQuery = url.Values{"ref": {ref}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	setGitHubAuth(req, token)
	body, err := doGet(hc, req)
	if err != nil {
		return nil, err
	}
	var entries []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("listing %s/%s: %w", repo, upstreamSpecDir, err)
	}
	var files []upstreamFile
	for _, e := range entries {
		m := upstreamSpecFile.FindStringSubmatch(e.Name)
		if e.Type != "file" || m == nil {
			continue
		}
		var key [3]int
		for i := range key {
			key[i], _ = strconv.Atoi(m[i+1])
		}
		files = append(files, upstreamFile{name: e.Name, version: m[1] + "." + m[2] + "." + m[3], key: key})
	}
	slices.SortFunc(files, func(a, b upstreamFile) int { return slices.Compare(a.key[:], b.key[:]) })
	return files, nil
}

func setGitHubAuth(req *http.Request, token string) {
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// doGet sends req and returns the body of a 200 response. Error messages
// carry the URL without credentials and never the request headers.
func doGet(hc *http.Client, req *http.Request) ([]byte, error) {
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", req.URL.Redacted(), err)
	}
	defer resp.Body.Close() //nolint:errcheck // nothing to do about a failed close of a read body
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSpecBytes+1))
	if err != nil {
		return nil, fmt.Errorf("GET %s: reading body: %w", req.URL.Redacted(), err)
	}
	if len(body) > maxSpecBytes {
		return nil, fmt.Errorf("GET %s: response larger than %d MiB", req.URL.Redacted(), maxSpecBytes>>20)
	}
	if resp.StatusCode != http.StatusOK {
		hint := ""
		if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			hint = fmt.Sprintf(" (GitHub rate limit reached; set $%s)", githubTokenEnv)
		}
		return nil, fmt.Errorf("GET %s: %s%s", req.URL.Redacted(), resp.Status, hint)
	}
	return body, nil
}

// normalizeSpec validates that data is an OpenAPI document and re-encodes it
// as indented JSON with sorted keys, so that vendored files diff cleanly.
// Elide leaves info.version empty; it is set to version.
func normalizeSpec(data []byte, version string) ([]byte, error) {
	spec, err := speccheck.LoadData(data)
	if err != nil {
		return nil, err
	}
	doc := spec.Doc()
	if doc.OpenAPI == "" || doc.Paths == nil || doc.Paths.Len() == 0 {
		return nil, errors.New("not an OpenAPI 3 document with paths")
	}
	if doc.Info != nil {
		doc.Info.Version = version
	}
	compact, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var generic any
	if err := json.Unmarshal(compact, &generic); err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(generic, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
