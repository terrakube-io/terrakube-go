package terrakube_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	terrakube "github.com/terrakube-io/terrakube-go"
)

// jsonapiOne builds a single-resource JSON:API response body.
func jsonapiOne(typeName, id string, attrs map[string]interface{}) []byte {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       typeName,
			"id":         id,
			"attributes": attrs,
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

// jsonapiList builds a multi-resource JSON:API response body.
func jsonapiList(items []map[string]interface{}) []byte {
	resp := map[string]interface{}{"data": items}
	b, _ := json.Marshal(resp)
	return b
}

// jsonapiItem is a shorthand for building one item in a list payload.
func jsonapiItem(typeName, id string, attrs map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":       typeName,
		"id":         id,
		"attributes": attrs,
	}
}

func ExampleNewClient() {
	// In production, point to your real Terrakube server:
	//
	//   client, err := terrakube.NewClient(
	//       terrakube.WithEndpoint("https://terrakube.example.com"),
	//       terrakube.WithToken("your-api-token"),
	//   )

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("organization", "org-1", map[string]interface{}{"name": "my-org"}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := client.Organizations.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(orgs[0].Name)
	srv.Close()
	// Output: my-org
}

func ExampleOrganizationService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("organization", "org-1", map[string]interface{}{
				"name":          "Alpha",
				"executionMode": "remote",
			}),
			jsonapiItem("organization", "org-2", map[string]interface{}{
				"name":          "Beta",
				"executionMode": "local",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := client.Organizations.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, org := range orgs {
		fmt.Printf("%s (%s)\n", org.Name, org.ExecutionMode)
	}
	srv.Close()
	// Output:
	// Alpha (remote)
	// Beta (local)
}

func ExampleOrganizationService_Get() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiOne("organization", "org-1", map[string]interface{}{
			"name":          "Alpha",
			"description":   "Primary organization",
			"executionMode": "remote",
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	org, err := client.Organizations.Get(context.Background(), "org-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %s\n", org.Name, *org.Description)
	srv.Close()
	// Output: Alpha: Primary organization
}

func ExampleOrganizationService_Create() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(jsonapiOne("organization", "org-new", map[string]interface{}{
			"name":          "NewOrg",
			"executionMode": "remote",
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	created, err := client.Organizations.Create(context.Background(), &terrakube.Organization{
		Name:          "NewOrg",
		ExecutionMode: "remote",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (id=%s)\n", created.Name, created.ID)
	srv.Close()
	// Output: NewOrg (id=org-new)
}

func ExampleOrganizationService_List_filtered() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, r *http.Request) {
		// Verify that the filter query parameter is forwarded.
		if f := r.URL.Query().Get("filter"); f != "name==Alpha" {
			http.Error(w, "unexpected filter: "+f, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("organization", "org-1", map[string]interface{}{
				"name":          "Alpha",
				"executionMode": "remote",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := client.Organizations.List(context.Background(), &terrakube.ListOptions{
		Filter: "name==Alpha",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(orgs[0].Name)
	srv.Close()
	// Output: Alpha
}

func ExampleWorkspaceService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("workspace", "ws-1", map[string]interface{}{
				"name":             "production",
				"source":           "https://github.com/example/infra",
				"branch":           "main",
				"folder":           "/",
				"iacType":          "terraform",
				"terraformVersion": "1.5.0",
				"executionMode":    "remote",
			}),
			jsonapiItem("workspace", "ws-2", map[string]interface{}{
				"name":             "staging",
				"source":           "https://github.com/example/infra",
				"branch":           "develop",
				"folder":           "/",
				"iacType":          "terraform",
				"terraformVersion": "1.5.0",
				"executionMode":    "local",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	workspaces, err := client.Workspaces.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, ws := range workspaces {
		fmt.Printf("%s (branch=%s, mode=%s)\n", ws.Name, ws.Branch, ws.ExecutionMode)
	}
	srv.Close()
	// Output:
	// production (branch=main, mode=remote)
	// staging (branch=develop, mode=local)
}

func ExampleWorkspaceService_Get() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiOne("workspace", "ws-1", map[string]interface{}{
			"name":             "production",
			"source":           "https://github.com/example/infra",
			"branch":           "main",
			"folder":           "/envs/prod",
			"iacType":          "terraform",
			"terraformVersion": "1.6.0",
			"executionMode":    "remote",
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := client.Workspaces.Get(context.Background(), "org-1", "ws-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s branch=%s folder=%s\n", ws.Name, ws.Branch, ws.Folder)
	srv.Close()
	// Output: production branch=main folder=/envs/prod
}

func ExampleWorkspaceService_Create() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(jsonapiOne("workspace", "ws-new", map[string]interface{}{
			"name":             "dev",
			"source":           "https://github.com/example/infra",
			"branch":           "feature-x",
			"folder":           "/",
			"iacType":          "terraform",
			"terraformVersion": "1.6.0",
			"executionMode":    "local",
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := client.Workspaces.Create(context.Background(), "org-1", &terrakube.Workspace{
		Name:          "dev",
		Source:        "https://github.com/example/infra",
		Branch:        "feature-x",
		Folder:        "/",
		IaCType:       "terraform",
		IaCVersion:    "1.6.0",
		ExecutionMode: "local",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (id=%s)\n", ws.Name, ws.ID)
	srv.Close()
	// Output: dev (id=ws-new)
}

func ExampleWorkspaceService_Update() {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiOne("workspace", "ws-1", map[string]interface{}{
			"name":             "production",
			"source":           "https://github.com/example/infra",
			"branch":           "main",
			"folder":           "/",
			"iacType":          "terraform",
			"terraformVersion": "1.7.0",
			"executionMode":    "remote",
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := client.Workspaces.Update(context.Background(), "org-1", &terrakube.Workspace{
		ID:         "ws-1",
		IaCVersion: "1.7.0",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s version=%s\n", ws.Name, ws.IaCVersion)
	srv.Close()
	// Output: production version=1.7.0
}

func ExampleWorkspaceService_Delete() {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/organization/org-1/workspace/ws-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Workspaces.Delete(context.Background(), "org-1", "ws-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("workspace deleted")
	srv.Close()
	// Output: workspace deleted
}

func ExampleModuleService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/module", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("module", "mod-1", map[string]interface{}{
				"name":             "vpc",
				"description":      "AWS VPC module",
				"provider":         "aws",
				"source":           "https://github.com/example/terraform-aws-vpc",
				"downloadQuantity": 42,
			}),
			jsonapiItem("module", "mod-2", map[string]interface{}{
				"name":             "storage",
				"description":      "GCP storage module",
				"provider":         "google",
				"source":           "https://github.com/example/terraform-gcp-storage",
				"downloadQuantity": 17,
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	modules, err := client.Modules.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range modules {
		fmt.Printf("%s/%s downloads=%d\n", m.Provider, m.Name, m.DownloadQuantity)
	}
	srv.Close()
	// Output:
	// aws/vpc downloads=42
	// google/storage downloads=17
}

func ExampleModuleService_Get() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/module/mod-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiOne("module", "mod-1", map[string]interface{}{
			"name":             "vpc",
			"description":      "AWS VPC module",
			"provider":         "aws",
			"source":           "https://github.com/example/terraform-aws-vpc",
			"downloadQuantity": 42,
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mod, err := client.Modules.Get(context.Background(), "org-1", "mod-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (%s) - %s\n", mod.Name, mod.Provider, mod.Description)
	srv.Close()
	// Output: vpc (aws) - AWS VPC module
}

func ExampleTeamService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/team", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("team", "team-1", map[string]interface{}{
				"name":            "admins",
				"manageWorkspace": true,
				"manageModule":    true,
			}),
			jsonapiItem("team", "team-2", map[string]interface{}{
				"name":            "developers",
				"manageWorkspace": false,
				"manageModule":    false,
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	teams, err := client.Teams.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range teams {
		fmt.Printf("%s manageWorkspace=%v\n", t.Name, t.ManageWorkspace)
	}
	srv.Close()
	// Output:
	// admins manageWorkspace=true
	// developers manageWorkspace=false
}

func ExampleVariableService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/workspace/ws-1/variable", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("variable", "var-1", map[string]interface{}{
				"key":       "AWS_REGION",
				"value":     "us-east-1",
				"category":  "ENV",
				"sensitive": false,
				"hcl":       false,
			}),
			jsonapiItem("variable", "var-2", map[string]interface{}{
				"key":       "DB_PASSWORD",
				"value":     "",
				"category":  "ENV",
				"sensitive": true,
				"hcl":       false,
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	vars, err := client.Variables.List(context.Background(), "org-1", "ws-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range vars {
		fmt.Printf("%s sensitive=%v\n", v.Key, v.Sensitive)
	}
	srv.Close()
	// Output:
	// AWS_REGION sensitive=false
	// DB_PASSWORD sensitive=true
}

func ExampleVariableService_Create() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organization/org-1/workspace/ws-1/variable", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(jsonapiOne("variable", "var-new", map[string]interface{}{
			"key":       "TF_VAR_region",
			"value":     "eu-west-1",
			"category":  "TERRAFORM",
			"sensitive": false,
			"hcl":       false,
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	v, err := client.Variables.Create(context.Background(), "org-1", "ws-1", &terrakube.Variable{
		Key:      "TF_VAR_region",
		Value:    "eu-west-1",
		Category: "TERRAFORM",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s=%s (id=%s)\n", v.Key, v.Value, v.ID)
	srv.Close()
	// Output: TF_VAR_region=eu-west-1 (id=var-new)
}

func ExampleTemplateService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/template", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("template", "tmpl-1", map[string]interface{}{
				"name": "Plan",
				"tcl":  "flow:\n- type: terraformPlan",
			}),
			jsonapiItem("template", "tmpl-2", map[string]interface{}{
				"name": "Apply",
				"tcl":  "flow:\n- type: terraformApply",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	templates, err := client.Templates.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range templates {
		fmt.Println(t.Name)
	}
	srv.Close()
	// Output:
	// Plan
	// Apply
}

func ExampleJobService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/job", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("job", "job-1", map[string]interface{}{
				"status":  "completed",
				"command": "apply",
				"output":  "Apply complete! Resources: 3 added.",
			}),
			jsonapiItem("job", "job-2", map[string]interface{}{
				"status":  "running",
				"command": "plan",
				"output":  "",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := client.Jobs.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, j := range jobs {
		fmt.Printf("%s status=%s\n", j.ID, j.Status)
	}
	srv.Close()
	// Output:
	// job-1 status=completed
	// job-2 status=running
}

func ExampleTeamTokenService_Create() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /access-token/v1/teams", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"tok-1","description":"CI token","days":30,"hours":0,"minutes":0,"group":"TERRAKUBE_DEVELOPERS","token":"tkn_example_abc123"}`))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	token, err := client.TeamTokens.Create(context.Background(), &terrakube.TeamToken{
		Description: "CI token",
		Days:        30,
		Group:       "TERRAKUBE_DEVELOPERS",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("id=%s group=%s\n", token.ID, token.Group)
	srv.Close()
	// Output: id=tok-1 group=TERRAKUBE_DEVELOPERS
}

func ExampleIsNotFound() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"detail":"organization not found","status":"404"}]}`))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Organizations.Get(context.Background(), "org-1")
	fmt.Println(terrakube.IsNotFound(err))
	srv.Close()
	// Output: true
}

func ExampleSSHService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/ssh", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("ssh", "ssh-1", map[string]interface{}{
				"name":       "deploy-key",
				"sshType":    "rsa",
				"privateKey": "-----BEGIN RSA PRIVATE KEY-----\nexample\n-----END RSA PRIVATE KEY-----",
			}),
			jsonapiItem("ssh", "ssh-2", map[string]interface{}{
				"name":       "module-key",
				"sshType":    "ed25519",
				"privateKey": "-----BEGIN OPENSSH PRIVATE KEY-----\nexample\n-----END OPENSSH PRIVATE KEY-----",
			}),
		}))
	})
	srv := httptest.NewServer(mux)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(srv.URL),
		terrakube.WithToken("example-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	keys, err := client.SSH.List(context.Background(), "org-1", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, k := range keys {
		fmt.Printf("%s type=%s\n", k.Name, k.SSHType)
	}
	srv.Close()
	// Output:
	// deploy-key type=rsa
	// module-key type=ed25519
}
