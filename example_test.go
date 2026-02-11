package terrakube_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	terrakube "github.com/denniswebb/terrakube-go"
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
		w.Write(jsonapiList([]map[string]interface{}{
			jsonapiItem("organization", "org-1", map[string]interface{}{"name": "my-org"}),
		}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

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
	// Output: my-org
}

func ExampleOrganizationService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write(jsonapiList([]map[string]interface{}{
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
	defer srv.Close()

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
	// Output:
	// Alpha (remote)
	// Beta (local)
}

func ExampleOrganizationService_Get() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write(jsonapiOne("organization", "org-1", map[string]interface{}{
			"name":          "Alpha",
			"description":   "Primary organization",
			"executionMode": "remote",
		}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

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
	// Output: Alpha: Primary organization
}

func ExampleOrganizationService_Create() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organization", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonapiOne("organization", "org-new", map[string]interface{}{
			"name":          "NewOrg",
			"executionMode": "remote",
		}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

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
	// Output: NewOrg (id=org-new)
}

func ExampleWorkspaceService_List() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organization/org-1/workspace", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write(jsonapiList([]map[string]interface{}{
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
	defer srv.Close()

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
	// Output:
	// production (branch=main, mode=remote)
	// staging (branch=develop, mode=local)
}
