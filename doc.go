// Package terrakube provides a Go client library for the Terrakube API.
//
// Terrakube is an open-source platform for managing infrastructure as code
// workspaces, modules, teams, and related resources via a JSON:API interface.
//
// Create a client with NewClient and use the service fields to interact
// with each resource type:
//
//	client, err := terrakube.NewClient(
//	    terrakube.WithEndpoint("https://terrakube.example.com"),
//	    terrakube.WithToken("your-api-token"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	orgs, err := client.Organizations.List(ctx, nil)
package terrakube
