//go:build smoke

package terrakube_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	terrakube "github.com/terrakube-io/terrakube-go"
)

// TestSmoke runs read/write round trips against a real Terrakube server. The
// unit tests answer from mocks written from this library's own assumptions;
// this test is the one that notices when the server disagrees with them.
//
// It needs a server of the version in APIVersion and a token of an instance
// administrator:
//
//	TERRAKUBE_SMOKE_ENDPOINT=https://terrakube-api.localhost \
//	TERRAKUBE_SMOKE_TOKEN=... go test -tags=smoke -run TestSmoke -count=1 .
//
// Optional: TERRAKUBE_SMOKE_INSECURE=1 skips TLS verification, and
// TERRAKUBE_SMOKE_GROUP names the group of the token (default TERRAKUBE_ADMIN).
//
// Without a token, the test signs one itself from TERRAKUBE_SMOKE_PAT_SECRET
// (the server's PatSecret) and TERRAKUBE_SMOKE_PAT_ID (the id of a row in the
// server's pat table). CI uses this: nobody is there to log in.
// Everything is created inside a new organization, which is disabled at the end.
func TestSmoke(t *testing.T) {
	group := os.Getenv("TERRAKUBE_SMOKE_GROUP")
	if group == "" {
		group = "TERRAKUBE_ADMIN"
	}
	endpoint, token := os.Getenv("TERRAKUBE_SMOKE_ENDPOINT"), os.Getenv("TERRAKUBE_SMOKE_TOKEN")
	if secret, id := os.Getenv("TERRAKUBE_SMOKE_PAT_SECRET"), os.Getenv("TERRAKUBE_SMOKE_PAT_ID"); token == "" && secret != "" && id != "" {
		var err error
		if token, err = signPAT(secret, id, group); err != nil {
			t.Fatalf("signing a token: %v", err)
		}
	}
	if endpoint == "" || token == "" {
		t.Fatal("set TERRAKUBE_SMOKE_ENDPOINT and either TERRAKUBE_SMOKE_TOKEN or TERRAKUBE_SMOKE_PAT_SECRET with TERRAKUBE_SMOKE_PAT_ID")
	}
	opts := []terrakube.Option{terrakube.WithEndpoint(endpoint), terrakube.WithToken(token), terrakube.WithUserAgent("terrakube-go-smoke")}
	if os.Getenv("TERRAKUBE_SMOKE_INSECURE") == "1" {
		opts = append(opts, terrakube.WithInsecureTLS())
	}
	c, err := terrakube.NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// must stops the test: every later step depends on the earlier ones.
	must := func(step string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", step, err)
		}
	}
	started := time.Now().Add(-time.Minute) // tolerate clock difference with the server

	// --- organization: create, get, update, list with a filter ---
	name := fmt.Sprintf("smoke-%d", time.Now().UnixNano())
	org, err := c.Organizations.Create(ctx, &terrakube.Organization{
		Name: name, Description: new("terrakube-go smoke test"), ExecutionMode: "remote",
	})
	must("Organizations.Create", err)
	orgID := org.ID // org is reassigned below; the cleanup must not depend on it
	t.Cleanup(func() {
		// A fresh context: the test's own may have expired.
		cctx, ccancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ccancel()
		// The server refuses to delete an organization that still has
		// resources (409), so it is disabled, which is how the UI removes one.
		o, err := c.Organizations.Get(cctx, orgID)
		if err == nil {
			o.Disabled = true
			_, err = c.Organizations.Update(cctx, o)
		}
		if err != nil {
			t.Errorf("cleanup: disabling organization %s: %v", orgID, err)
		}
	})

	// The token's group needs a team in the new organization to manage its resources.
	_, err = c.Teams.Create(ctx, org.ID, &terrakube.Team{
		Name: group, ManageState: true, ManageWorkspace: true, ManageModule: true, ManageProvider: true,
		ManageVcs: true, ManageTemplate: true, ManageJob: true, ManageCollection: true,
		PlanJob: true, ApproveJob: true, DestroyJob: true, ManagePolicies: true,
	})
	must("Teams.Create", err)

	org, err = c.Organizations.Get(ctx, org.ID)
	must("Organizations.Get", err)
	org.Description = new("terrakube-go smoke test, updated")
	org, err = c.Organizations.Update(ctx, org)
	must("Organizations.Update", err)
	if org.Name != name || org.Description == nil || *org.Description != "terrakube-go smoke test, updated" {
		t.Errorf("organization after update = %+v: name must survive a Get-change-Update, description must change", org)
	}
	orgs, err := c.Organizations.List(ctx, &terrakube.ListOptions{Filter: "name==" + name})
	must("Organizations.List(filter)", err)
	if len(orgs) != 1 || orgs[0].ID != org.ID {
		t.Errorf("Organizations.List(filter) returned %d organizations, want only the new one", len(orgs))
	}
	if _, err := c.Organizations.List(ctx, &terrakube.ListOptions{PageSize: 1}); !isValidationError(err) {
		t.Errorf("Organizations.List(PageSize) err = %v, want *ValidationError (the server cannot page organizations)", err)
	}

	// --- workspace: a paged type, with dates ---
	ws, err := c.Workspaces.Create(ctx, org.ID, &terrakube.Workspace{
		Name: "smoke", Source: "https://github.com/terrakube-io/terrakube-go.git", Branch: "main", Folder: "/",
		IaCType: "terraform", IaCVersion: "1.9.8", ExecutionMode: "remote",
	})
	must("Workspaces.Create", err)
	second, err := c.Workspaces.Create(ctx, org.ID, &terrakube.Workspace{
		Name: "smoke-2", Source: "https://github.com/terrakube-io/terrakube-go.git", Branch: "main", Folder: "/",
		IaCType: "terraform", IaCVersion: "1.9.8", ExecutionMode: "remote",
	})
	must("Workspaces.Create (second)", err)

	ws, err = c.Workspaces.Get(ctx, org.ID, ws.ID)
	must("Workspaces.Get", err)
	if ws.CreatedDate == nil || ws.CreatedDate.Before(started) {
		t.Errorf("workspace CreatedDate = %v, want a time after %v", ws.CreatedDate, started)
	}
	ws.Description = new("updated")
	ws, err = c.Workspaces.Update(ctx, org.ID, ws) // sends the dates back to the server
	must("Workspaces.Update", err)
	if ws.Name != "smoke" || ws.Branch != "main" || ws.Description == nil || *ws.Description != "updated" {
		t.Errorf("workspace after update = name %q branch %q description %v", ws.Name, ws.Branch, ws.Description)
	}

	all, err := c.Workspaces.List(ctx, org.ID, nil)
	must("Workspaces.List", err)
	wantIDs := []string{ws.ID, second.ID}
	slices.Sort(wantIDs)
	if got := workspaceIDs(all); !slices.Equal(got, wantIDs) {
		t.Errorf("Workspaces.List ids = %v, want %v in id order", got, wantIDs)
	}
	var paged []string
	for page := 1; page <= 3; page++ { // the third page must be empty
		items, err := c.Workspaces.List(ctx, org.ID, &terrakube.ListOptions{Page: page, PageSize: 1})
		must(fmt.Sprintf("Workspaces.List(page %d)", page), err)
		if page <= 2 && len(items) != 1 {
			t.Errorf("Workspaces.List(page %d, size 1) returned %d items, want 1", page, len(items))
		}
		paged = append(paged, workspaceIDs(items)...)
	}
	slices.Sort(paged)
	if !slices.Equal(paged, wantIDs) {
		t.Errorf("pages of one = %v, want %v", paged, wantIDs)
	}

	// --- variable: create, list, update, delete ---
	v, err := c.Variables.Create(ctx, org.ID, ws.ID, &terrakube.Variable{Key: "smoke_key", Value: "one", Category: "TERRAFORM"})
	must("Variables.Create", err)
	v.Value = "two"
	v, err = c.Variables.Update(ctx, org.ID, ws.ID, v)
	must("Variables.Update", err)
	vars, err := c.Variables.List(ctx, org.ID, ws.ID, nil)
	must("Variables.List", err)
	if len(vars) != 1 || vars[0].Key != "smoke_key" || vars[0].Value != "two" || vars[0].Sensitive {
		t.Errorf("Variables.List = %+v, want one variable smoke_key=two, not sensitive", vars)
	}
	must("Variables.Delete", c.Variables.Delete(ctx, org.ID, ws.ID, v.ID))
	if _, err := c.Variables.Get(ctx, org.ID, ws.ID, v.ID); !terrakube.IsNotFound(err) {
		t.Errorf("Variables.Get after delete err = %v, want a 404", err)
	}

	// --- webhook: a to-one relationship; this workspace has none ---
	hooks, err := c.Webhooks.List(ctx, org.ID, ws.ID, nil)
	must("Webhooks.List", err)
	if len(hooks) != 0 {
		t.Errorf("Webhooks.List returned %d webhooks, want 0", len(hooks))
	}

	// --- atomic operations: add a tag, then find it ---
	res, err := c.Operations.Submit(ctx, &terrakube.AtomicRequest{Operations: []terrakube.Operation{{
		Op:   terrakube.OperationAdd,
		Href: "/organization/" + org.ID + "/tag",
		// The server wants a local id on a resource added through a relationship path.
		Data: map[string]any{"type": "tag", "lid": "smoke-tag", "attributes": map[string]any{"name": "smoke-tag"}},
	}}})
	must("Operations.Submit", err)
	if len(res.Results) != 1 {
		t.Errorf("Operations.Submit returned %d results, want 1", len(res.Results))
	}
	tags, err := c.Tags.List(ctx, org.ID, nil)
	must("Tags.List", err)
	if len(tags) != 1 || tags[0].Name != "smoke-tag" {
		t.Errorf("Tags.List = %d tags, want the one added through Operations.Submit", len(tags))
	}

	// --- job: creating one is what failed with unknown attributes ---
	tpl, err := c.Templates.Create(ctx, org.ID, &terrakube.Template{
		Name:    "smoke-plan",
		Content: "flow:\n  - type: \"terraformPlan\"\n    name: \"Plan\"\n    step: 100\n",
	})
	must("Templates.Create", err)
	job, err := c.Jobs.Create(ctx, org.ID, &terrakube.Job{TemplateReference: &tpl.ID, Workspace: ws})
	must("Jobs.Create", err)
	job, err = c.Jobs.Get(ctx, org.ID, job.ID)
	must("Jobs.Get", err)
	if job.Status == "" || job.CreatedDate == nil {
		t.Errorf("job after create: status %q, created %v; want both set", job.Status, job.CreatedDate)
	}
	jobs, err := c.Jobs.List(ctx, org.ID, nil)
	must("Jobs.List", err)
	if len(jobs) != 1 {
		t.Errorf("Jobs.List returned %d jobs, want 1", len(jobs))
	}
}

func workspaceIDs(ws []*terrakube.Workspace) []string {
	ids := make([]string, len(ws))
	for i, w := range ws {
		ids[i] = w.ID
	}
	return ids
}

func isValidationError(err error) bool {
	_, ok := errors.AsType[*terrakube.ValidationError](err)
	return ok
}

// signPAT builds the token a Terrakube server issues as a personal access
// token: an HS256 JWT signed with the server's PatSecret, whose id names a row
// in the pat table.
func signPAT(secret, id, group string) (string, error) {
	key, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(secret, "="))
	if err != nil {
		return "", fmt.Errorf("decoding the secret: %w", err)
	}
	now := time.Now()
	claims, err := json.Marshal(map[string]any{
		"iss": "Terrakube", "aud": []string{"Terrakube"}, "sub": "smoke (Token)", "jti": id,
		"email": "smoke@example.com", "email_verified": true, "name": "smoke (Token)",
		"groups": []string{group}, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding.EncodeToString
	unsigned := enc([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + enc(claims)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc(mac.Sum(nil)), nil
}
