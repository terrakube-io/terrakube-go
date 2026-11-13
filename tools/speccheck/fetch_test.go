package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const fixtureYAML = `openapi: 3.0.1
info:
  title: Terrakube
  version: ""
paths:
  /organization:
    get:
      responses:
        "200":
          description: ok
`

const fixtureJSON = `{"openapi":"3.0.1","info":{"title":"Terrakube","version":""},"paths":{"/organization":{"get":{"responses":{"200":{"description":"ok"}}}}}}`

type seenRequest struct {
	path, query, accept, auth string
}

// fakeGitHub serves a contents API listing for openapi-spec/ and the raw files.
type fakeGitHub struct {
	*httptest.Server
	mu   sync.Mutex
	seen []seenRequest
}

func newFakeGitHub(t *testing.T, files ...string) *fakeGitHub {
	t.Helper()
	g := &fakeGitHub{}
	g.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		g.seen = append(g.seen, seenRequest{r.URL.Path, r.URL.RawQuery, r.Header.Get("Accept"), r.Header.Get("Authorization")})
		g.mu.Unlock()
		const dir = "/repos/terrakube-io/terrakube/contents/openapi-spec"
		switch {
		case r.URL.Path == dir:
			entries := make([]map[string]string, 0, len(files)+1)
			for _, f := range files {
				entries = append(entries, map[string]string{"name": f, "type": "file"})
			}
			entries = append(entries, map[string]string{"name": "v9_9_9.yml", "type": "dir"}) // directories are ignored
			_ = json.NewEncoder(w).Encode(entries)
		case strings.HasPrefix(r.URL.Path, dir+"/"):
			if !strings.HasSuffix(r.URL.Path, ".yml") {
				http.NotFound(w, r)
				return
			}
			_, _ = io.WriteString(w, fixtureYAML)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(g.Close)
	return g
}

func (g *fakeGitHub) requests() []seenRequest {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]seenRequest(nil), g.seen...)
}

func githubOptions(g *fakeGitHub, out string) fetchOptions {
	return fetchOptions{version: "latest", output: out, repo: defaultRepo, githubAPI: g.URL}
}

func doFetch(t *testing.T, o fetchOptions, hc *http.Client) (stdout, stderr string, err error) {
	t.Helper()
	var out, errb strings.Builder
	err = fetch(context.Background(), o, hc, &out, &errb)
	return out.String(), errb.String(), err
}

func readInfoVersion(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	return doc.Info.Version
}

func TestFetch_GitHubLatestPicksHighestVersion(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_27_0.yml", "v2_30_0.yml", "README.md")
	out := filepath.Join(t.TempDir(), "spec.json")

	stdout, _, err := doFetch(t, githubOptions(g, out), g.Client())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if got := readInfoVersion(t, out); got != "2.30.0" {
		t.Errorf("info.version = %q, want 2.30.0", got)
	}
	if !strings.Contains(stdout, "version 2.30.0") || !strings.Contains(stdout, "v2_30_0.yml") {
		t.Errorf("stdout = %q", stdout)
	}
	data, _ := os.ReadFile(out)
	if !strings.HasPrefix(string(data), "{\n  \"") || !strings.HasSuffix(string(data), "}\n") {
		t.Errorf("output is not indented JSON with a trailing newline:\n%s", data)
	}
	reqs := g.requests()
	if len(reqs) != 2 || reqs[1].accept != "application/vnd.github.raw+json" || reqs[0].accept != "application/vnd.github+json" {
		t.Errorf("requests = %+v", reqs)
	}
}

func TestFetch_GitHubExplicitVersion(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_27_0.yml", "v2_30_0.yml")

	for _, version := range []string{"2.27.0", "v2.27.0"} {
		out := filepath.Join(t.TempDir(), "spec.json")
		o := githubOptions(g, out)
		o.version = version
		if _, _, err := doFetch(t, o, g.Client()); err != nil {
			t.Fatalf("fetch %s: %v", version, err)
		}
		if got := readInfoVersion(t, out); got != "2.27.0" {
			t.Errorf("version %s: info.version = %q, want 2.27.0", version, got)
		}
	}
}

func TestFetch_GitHubMissingVersionListsAvailable(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_27_0.yml", "v2_30_0.yml")
	o := githubOptions(g, filepath.Join(t.TempDir(), "spec.json"))
	o.version = "9.9.9"

	_, _, err := doFetch(t, o, g.Client())
	if err == nil || !strings.Contains(err.Error(), "available: 2.27.0, 2.30.0") {
		t.Fatalf("err = %v, want one listing available versions", err)
	}
}

func TestFetch_GitHubNoSpecFiles(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "README.md")

	_, _, err := doFetch(t, githubOptions(g, filepath.Join(t.TempDir(), "spec.json")), g.Client())
	if err == nil || !strings.Contains(err.Error(), "has no openapi-spec/v*_*_*.yml files") {
		t.Fatalf("err = %v", err)
	}
}

func TestFetch_GitHubPassesRefAsQueryParameter(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_30_0.yml")
	o := githubOptions(g, filepath.Join(t.TempDir(), "spec.json"))
	o.ref = "release/2.30"

	if _, _, err := doFetch(t, o, g.Client()); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	for _, r := range g.requests() {
		if r.query != "ref=release%2F2.30" {
			t.Errorf("request %s query = %q, want ref=release%%2F2.30", r.path, r.query)
		}
	}
}

func TestFetch_GitHubWarnsWithoutToken(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_30_0.yml")

	_, stderr, err := doFetch(t, githubOptions(g, filepath.Join(t.TempDir(), "spec.json")), g.Client())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !strings.Contains(stderr, "GITHUB_TOKEN is not set") {
		t.Errorf("stderr = %q, want rate-limit warning", stderr)
	}
	for _, r := range g.requests() {
		if r.auth != "" {
			t.Errorf("Authorization = %q, want none", r.auth)
		}
	}
}

func TestFetch_GitHubSendsTokenWhenSet(t *testing.T) {
	t.Setenv(githubTokenEnv, "ghp-secret")
	g := newFakeGitHub(t, "v2_30_0.yml")

	_, stderr, err := doFetch(t, githubOptions(g, filepath.Join(t.TempDir(), "spec.json")), g.Client())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if strings.Contains(stderr, "is not set") {
		t.Errorf("stderr = %q, want no rate-limit warning when the token is set", stderr)
	}
	if !strings.Contains(stderr, "plain http") {
		t.Errorf("stderr = %q, want a plain-http warning for the http test server", stderr)
	}
	for _, r := range g.requests() {
		if r.auth != "Bearer ghp-secret" {
			t.Errorf("Authorization = %q, want bearer token", r.auth)
		}
	}
}

func TestFetch_GitHubRateLimitHint(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	o := fetchOptions{version: "latest", output: filepath.Join(t.TempDir(), "s.json"), repo: defaultRepo, githubAPI: srv.URL}

	_, _, err := doFetch(t, o, srv.Client())
	if err == nil || !strings.Contains(err.Error(), "rate limit reached; set $GITHUB_TOKEN") {
		t.Fatalf("err = %v, want rate limit hint", err)
	}
}

func TestFetch_GitHubForbiddenWithoutRateLimitHasNoHint(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	o := fetchOptions{version: "latest", output: filepath.Join(t.TempDir(), "s.json"), repo: defaultRepo, githubAPI: srv.URL}

	_, _, err := doFetch(t, o, srv.Client())
	if err == nil || strings.Contains(err.Error(), "rate limit") || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v, want a plain 403", err)
	}
}

func TestFetch_GitHubListingNotJSON(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "nope") }))
	t.Cleanup(srv.Close)
	o := fetchOptions{version: "latest", output: filepath.Join(t.TempDir(), "s.json"), repo: defaultRepo, githubAPI: srv.URL}

	if _, _, err := doFetch(t, o, srv.Client()); err == nil {
		t.Fatal("fetch returned nil error for a non-JSON listing")
	}
}

func TestFetch_GitHubUnreachableServerReportsURL(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	o := fetchOptions{version: "latest", output: filepath.Join(t.TempDir(), "s.json"), repo: defaultRepo, githubAPI: url}

	_, _, err := doFetch(t, o, http.DefaultClient)
	if err == nil || !strings.Contains(err.Error(), "GET "+url) {
		t.Fatalf("err = %v, want one naming the URL", err)
	}
}

func TestFetch_RefusesToOverwriteWithoutForce(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_30_0.yml")
	out := filepath.Join(t.TempDir(), "spec.json")
	writeFile(t, out, "old")

	_, _, err := doFetch(t, githubOptions(g, out), g.Client())
	if err == nil || !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("err = %v, want overwrite refusal", err)
	}
	if data, _ := os.ReadFile(out); string(data) != "old" {
		t.Errorf("file was modified: %q", data)
	}

	o := githubOptions(g, out)
	o.force = true
	if _, _, err := doFetch(t, o, g.Client()); err != nil {
		t.Fatalf("fetch with force: %v", err)
	}
	if got := readInfoVersion(t, out); got != "2.30.0" {
		t.Errorf("info.version = %q after force", got)
	}
}

func TestFetch_DefaultOutputIsVendoredFileInRepo(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_30_0.yml")
	root := writeRepo(t)
	o := githubOptions(g, "")
	o.root = root

	if _, _, err := doFetch(t, o, g.Client()); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if got := readInfoVersion(t, filepath.Join(root, "openapi", "v2.30.0.json")); got != "2.30.0" {
		t.Errorf("info.version = %q", got)
	}
}

func TestFetch_ExplicitRootNeedNotExist(t *testing.T) {
	t.Setenv(githubTokenEnv, "")
	g := newFakeGitHub(t, "v2_30_0.yml")
	o := githubOptions(g, "")
	// An explicit root is not validated; the vendored file is created below it.
	o.root = filepath.Join(t.TempDir(), "missing")
	if _, _, err := doFetch(t, o, g.Client()); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if _, err := os.Stat(filepath.Join(o.root, "openapi", "v2.30.0.json")); err != nil {
		t.Errorf("vendored file not created: %v", err)
	}
}

type serverSeen struct {
	mu   sync.Mutex
	auth string
	path string
}

func newFakeTerrakube(t *testing.T, body string, seen *serverSeen) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.mu.Lock()
		seen.auth, seen.path = r.Header.Get("Authorization"), r.URL.Path
		seen.mu.Unlock()
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetch_ServerWritesIndentedJSONWithVersion(t *testing.T) {
	t.Setenv(tokenEnv, "tok-123")
	var seen serverSeen
	srv := newFakeTerrakube(t, fixtureJSON, &seen)
	out := filepath.Join(t.TempDir(), "spec.json")

	stdout, stderr, err := doFetch(t, fetchOptions{version: "v2.99.0", server: srv.URL, output: out}, srv.Client())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if seen.auth != "Bearer tok-123" || seen.path != "/doc" {
		t.Errorf("server saw auth=%q path=%q", seen.auth, seen.path)
	}
	if got := readInfoVersion(t, out); got != "2.99.0" {
		t.Errorf("info.version = %q, want 2.99.0", got)
	}
	if !strings.Contains(stdout, "version 2.99.0") {
		t.Errorf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "plain http") {
		t.Errorf("stderr = %q, want http warning", stderr)
	}
	if strings.Contains(stderr, "tok-123") || strings.Contains(stdout, "tok-123") {
		t.Error("token leaked into output")
	}
}

func TestFetch_ServerRequiresToken(t *testing.T) {
	t.Setenv(tokenEnv, "")
	var seen serverSeen
	srv := newFakeTerrakube(t, fixtureJSON, &seen)

	_, _, err := doFetch(t, fetchOptions{version: "2.99.0", server: srv.URL, output: filepath.Join(t.TempDir(), "s.json")}, srv.Client())
	if err == nil || !strings.Contains(err.Error(), "TERRAKUBE_TOKEN is not set") {
		t.Fatalf("err = %v", err)
	}
	if seen.path != "" {
		t.Error("server was contacted without a token")
	}
}

func TestFetch_ServerRequiresVersion(t *testing.T) {
	t.Setenv(tokenEnv, "tok")
	for _, version := range []string{"latest", "", "v"} {
		_, _, err := doFetch(t, fetchOptions{version: version, server: "http://127.0.0.1:1"}, http.DefaultClient)
		if err == nil || !strings.Contains(err.Error(), "--version is required with --server") {
			t.Errorf("version %q: err = %v", version, err)
		}
	}
}

func TestFetch_ServerRejectsInvalidURL(t *testing.T) {
	t.Setenv(tokenEnv, "tok")
	for _, server := range []string{"ftp://example.com", "example.com", "://bad"} {
		_, _, err := doFetch(t, fetchOptions{version: "1.0.0", server: server}, http.DefaultClient)
		if err == nil || !strings.Contains(err.Error(), "--server must be an http(s) URL") {
			t.Errorf("server %q: err = %v", server, err)
		}
	}
}

func TestFetch_ServerDocumentIsNotOpenAPI(t *testing.T) {
	t.Setenv(tokenEnv, "tok")
	var seen serverSeen
	srv := newFakeTerrakube(t, `{"hello":"world"}`, &seen)

	_, _, err := doFetch(t, fetchOptions{version: "1.0.0", server: srv.URL, output: filepath.Join(t.TempDir(), "s.json")}, srv.Client())
	if err == nil || !strings.Contains(err.Error(), "document from") {
		t.Fatalf("err = %v, want a rejected document", err)
	}
}

func TestFetch_ServerNon200(t *testing.T) {
	t.Setenv(tokenEnv, "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	t.Cleanup(srv.Close)

	_, _, err := doFetch(t, fetchOptions{version: "1.0.0", server: srv.URL}, srv.Client())
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v, want a 401", err)
	}
}

func TestRunFetch_ServerEndToEnd(t *testing.T) {
	t.Setenv(tokenEnv, "tok")
	var seen serverSeen
	srv := newFakeTerrakube(t, fixtureJSON, &seen)
	out := filepath.Join(t.TempDir(), "spec.json")

	var stdout, stderr strings.Builder
	if err := runFetch([]string{"--server", srv.URL, "--version", "3.0.0", "--output", out}, &stdout, &stderr); err != nil {
		t.Fatalf("runFetch: %v", err)
	}
	if got := readInfoVersion(t, out); got != "3.0.0" {
		t.Errorf("info.version = %q", got)
	}
}

func TestRunFetch_FlagErrors(t *testing.T) {
	t.Parallel()

	var stdout, stderr strings.Builder
	if err := runFetch([]string{"-h"}, &stdout, &stderr); !errors.Is(err, flag.ErrHelp) {
		t.Errorf("-h err = %v, want flag.ErrHelp", err)
	}
	if err := runFetch([]string{"--nope"}, &stdout, &stderr); err == nil {
		t.Error("unknown flag returned nil error")
	}
}

func TestNormalizeSpec(t *testing.T) {
	t.Parallel()

	t.Run("sets info.version and sorts keys", func(t *testing.T) {
		t.Parallel()
		out, err := normalizeSpec([]byte(fixtureYAML), "1.2.3")
		if err != nil {
			t.Fatalf("normalizeSpec: %v", err)
		}
		if !strings.Contains(string(out), `"version": "1.2.3"`) {
			t.Errorf("version not set:\n%s", out)
		}
		if strings.Index(string(out), `"info"`) > strings.Index(string(out), `"openapi"`) {
			t.Errorf("keys are not sorted:\n%s", out)
		}
	})

	for name, in := range map[string]string{
		"not a document":        `{"hello":"world"}`,
		"openapi without paths": `{"openapi":"3.0.1","info":{"title":"t","version":""},"paths":{}}`,
		"garbage":               "{not json",
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			if _, err := normalizeSpec([]byte(in), "1.0.0"); err == nil {
				t.Error("normalizeSpec returned nil error")
			}
		})
	}
}
