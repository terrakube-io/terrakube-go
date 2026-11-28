package terrakube_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

// The methods tested here address resources by their own ID on the server's top-level
// paths (/runTrigger/{id}, /policy_set/{id}/parameters, ...), for callers that do not
// know the organization or workspace a resource belongs to.

func TestByID_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		jsonapiType string
		call        func(*terrakube.Client, string) (string, error)
	}{
		{"RunTriggers.GetByID", "/api/v1/runTrigger/id-1", "runTrigger", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.RunTriggers.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"WebhookEvents.GetByID", "/api/v1/webhook_event/id-1", "webhook_event", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.WebhookEvents.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"Webhooks.GetByID", "/api/v1/webhook/id-1", "webhook", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.Webhooks.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"Workspaces.GetByID", "/api/v1/workspace/id-1", "workspace", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.Workspaces.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicySets.GetByID", "/api/v1/policy_set/id-1", "policy_set", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.PolicySets.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicyExemptions.GetByID", "/api/v1/policy_exemption/id-1", "policy_exemption", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.PolicyExemptions.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"NotificationConfigurations.GetByID", "/api/v1/notification_configuration/id-1", "notification_configuration", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.NotificationConfigurations.GetByID(context.Background(), id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicySetParameters.GetByPolicySet", "/api/v1/policy_set/ps-1/parameters/id-1", "policy_set_parameter", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.PolicySetParameters.GetByPolicySet(context.Background(), "ps-1", id)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicyAttachments.GetByPolicySet", "/api/v1/policy_set/ps-1/attachments/id-1", "policy_attachment", func(c *terrakube.Client, id string) (string, error) {
			v, err := c.PolicyAttachments.GetByPolicySet(context.Background(), "ps-1", id)
			return idOf(v, err, func() string { return v.ID })
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			srv.HandleFunc("GET "+tt.path, func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
					t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
				}
				writeResource(w, http.StatusOK, tt.jsonapiType, "id-1")
			})

			got, err := tt.call(newTestClient(t, srv), "id-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "id-1" {
				t.Errorf("ID = %q, want %q", got, "id-1")
			}
		})
	}
}

func TestByID_Get_NotFound(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/runTrigger/id-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteError(t, w, http.StatusNotFound, "not found")
	})

	_, err := newTestClient(t, srv).RunTriggers.GetByID(context.Background(), "id-1")
	if !terrakube.IsNotFound(err) {
		t.Fatalf("want a not-found error, got %v", err)
	}
}

func TestByID_UpdateFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		jsonapiType string
		call        func(*terrakube.Client) (string, error)
	}{
		{"RunTriggers.UpdateFieldsByID", "/api/v1/runTrigger/id-1", "runTrigger", func(c *terrakube.Client) (string, error) {
			v, err := c.RunTriggers.UpdateFieldsByID(context.Background(), "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicySets.UpdateFieldsByID", "/api/v1/policy_set/id-1", "policy_set", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicySets.UpdateFieldsByID(context.Background(), "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicyExemptions.UpdateFieldsByID", "/api/v1/policy_exemption/id-1", "policy_exemption", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicyExemptions.UpdateFieldsByID(context.Background(), "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"NotificationConfigurations.UpdateFieldsByID", "/api/v1/notification_configuration/id-1", "notification_configuration", func(c *terrakube.Client) (string, error) {
			v, err := c.NotificationConfigurations.UpdateFieldsByID(context.Background(), "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicySetParameters.UpdateFieldsByPolicySet", "/api/v1/policy_set/ps-1/parameters/id-1", "policy_set_parameter", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicySetParameters.UpdateFieldsByPolicySet(context.Background(), "ps-1", "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicyAttachments.UpdateFieldsByPolicySet", "/api/v1/policy_set/ps-1/attachments/id-1", "policy_attachment", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicyAttachments.UpdateFieldsByPolicySet(context.Background(), "ps-1", "id-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			handleUpdateFields(t, srv, tt.path, tt.jsonapiType)

			got, err := tt.call(newTestClient(t, srv))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "id-1" {
				t.Errorf("ID = %q, want %q", got, "id-1")
			}
		})
	}
}

func TestByID_CreateFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		jsonapiType string
		call        func(*terrakube.Client) (string, error)
	}{
		{"RunTriggers.CreateFieldsTopLevel", "/api/v1/runTrigger", "runTrigger", func(c *terrakube.Client) (string, error) {
			v, err := c.RunTriggers.CreateFieldsTopLevel(context.Background(), testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicySetParameters.CreateFieldsByPolicySet", "/api/v1/policy_set/ps-1/parameters", "policy_set_parameter", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicySetParameters.CreateFieldsByPolicySet(context.Background(), "ps-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"PolicyAttachments.CreateFieldsByPolicySet", "/api/v1/policy_set/ps-1/attachments", "policy_attachment", func(c *terrakube.Client) (string, error) {
			v, err := c.PolicyAttachments.CreateFieldsByPolicySet(context.Background(), "ps-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
		{"NotificationTriggers.CreateFieldsByConfiguration", "/api/v1/notification_configuration/cfg-1/triggers", "notification_trigger", func(c *terrakube.Client) (string, error) {
			v, err := c.NotificationTriggers.CreateFieldsByConfiguration(context.Background(), "cfg-1", testPatch)
			return idOf(v, err, func() string { return v.ID })
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			handleCreateFields(t, srv, tt.path, tt.jsonapiType)

			got, err := tt.call(newTestClient(t, srv))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "new-1" {
				t.Errorf("ID = %q, want %q", got, "new-1")
			}
		})
	}
}

func TestByID_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		call func(*terrakube.Client) error
	}{
		{"RunTriggers.DeleteByID", "/api/v1/runTrigger/id-1", func(c *terrakube.Client) error {
			return c.RunTriggers.DeleteByID(context.Background(), "id-1")
		}},
		{"PolicySets.DeleteByID", "/api/v1/policy_set/id-1", func(c *terrakube.Client) error {
			return c.PolicySets.DeleteByID(context.Background(), "id-1")
		}},
		{"PolicyExemptions.DeleteByID", "/api/v1/policy_exemption/id-1", func(c *terrakube.Client) error {
			return c.PolicyExemptions.DeleteByID(context.Background(), "id-1")
		}},
		{"NotificationConfigurations.DeleteByID", "/api/v1/notification_configuration/id-1", func(c *terrakube.Client) error {
			return c.NotificationConfigurations.DeleteByID(context.Background(), "id-1")
		}},
		{"PolicySetParameters.DeleteByPolicySet", "/api/v1/policy_set/ps-1/parameters/id-1", func(c *terrakube.Client) error {
			return c.PolicySetParameters.DeleteByPolicySet(context.Background(), "ps-1", "id-1")
		}},
		{"PolicyAttachments.DeleteByPolicySet", "/api/v1/policy_set/ps-1/attachments/id-1", func(c *terrakube.Client) error {
			return c.PolicyAttachments.DeleteByPolicySet(context.Background(), "ps-1", "id-1")
		}},
		{"NotificationTriggers.DeleteByConfiguration", "/api/v1/notification_configuration/cfg-1/triggers/id-1", func(c *terrakube.Client) error {
			return c.NotificationTriggers.DeleteByConfiguration(context.Background(), "cfg-1", "id-1")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			srv.HandleFunc("DELETE "+tt.path, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			if err := tt.call(newTestClient(t, srv)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestByID_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		entities any
		call     func(*terrakube.Client) (int, error)
	}{
		{"RunTriggers.ListTopLevel", "/api/v1/runTrigger", []*terrakube.RunTrigger{{ID: "a"}, {ID: "b"}}, func(c *terrakube.Client) (int, error) {
			v, err := c.RunTriggers.ListTopLevel(context.Background(), nil)
			return len(v), err
		}},
		{"PolicySetParameters.ListByPolicySet", "/api/v1/policy_set/ps-1/parameters", []*terrakube.PolicySetParameter{{ID: "a"}, {ID: "b"}}, func(c *terrakube.Client) (int, error) {
			v, err := c.PolicySetParameters.ListByPolicySet(context.Background(), "ps-1", nil)
			return len(v), err
		}},
		{"PolicyAttachments.ListByPolicySet", "/api/v1/policy_set/ps-1/attachments", []*terrakube.PolicyAttachment{{ID: "a"}, {ID: "b"}}, func(c *terrakube.Client) (int, error) {
			v, err := c.PolicyAttachments.ListByPolicySet(context.Background(), "ps-1", nil)
			return len(v), err
		}},
		{"NotificationTriggers.ListByConfiguration", "/api/v1/notification_configuration/cfg-1/triggers", []*terrakube.NotificationTrigger{{ID: "a"}, {ID: "b"}}, func(c *terrakube.Client) (int, error) {
			v, err := c.NotificationTriggers.ListByConfiguration(context.Background(), "cfg-1", nil)
			return len(v), err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := testutil.NewServer(t)
			srv.HandleFunc("GET "+tt.path, func(w http.ResponseWriter, _ *http.Request) {
				testutil.WriteJSONAPIList(t, w, http.StatusOK, tt.entities)
			})

			got, err := tt.call(newTestClient(t, srv))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != 2 {
				t.Errorf("got %d items, want 2", got)
			}
		})
	}
}

func TestByID_EmptyID(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	c := newTestClient(t, srv)
	ctx := context.Background()

	tests := []struct {
		name  string
		field string
		call  func() error
	}{
		{"RunTriggers.GetByID", "runTriggerID", func() error { _, err := c.RunTriggers.GetByID(ctx, ""); return err }},
		{"RunTriggers.UpdateFieldsByID", "runTriggerID", func() error { _, err := c.RunTriggers.UpdateFieldsByID(ctx, "", testPatch); return err }},
		{"RunTriggers.DeleteByID", "runTriggerID", func() error { return c.RunTriggers.DeleteByID(ctx, "") }},
		{"WebhookEvents.GetByID", "eventID", func() error { _, err := c.WebhookEvents.GetByID(ctx, ""); return err }},
		{"Webhooks.GetByID", "webhookID", func() error { _, err := c.Webhooks.GetByID(ctx, ""); return err }},
		{"Workspaces.GetByID", "workspaceID", func() error { _, err := c.Workspaces.GetByID(ctx, ""); return err }},
		{"PolicySets.GetByID", "policySetID", func() error { _, err := c.PolicySets.GetByID(ctx, ""); return err }},
		{"PolicySets.UpdateFieldsByID", "policySetID", func() error { _, err := c.PolicySets.UpdateFieldsByID(ctx, "", testPatch); return err }},
		{"PolicySets.DeleteByID", "policySetID", func() error { return c.PolicySets.DeleteByID(ctx, "") }},
		{"PolicyExemptions.GetByID", "policyExemptionID", func() error { _, err := c.PolicyExemptions.GetByID(ctx, ""); return err }},
		{"PolicyExemptions.UpdateFieldsByID", "policyExemptionID", func() error { _, err := c.PolicyExemptions.UpdateFieldsByID(ctx, "", testPatch); return err }},
		{"PolicyExemptions.DeleteByID", "policyExemptionID", func() error { return c.PolicyExemptions.DeleteByID(ctx, "") }},
		{"NotificationConfigurations.GetByID", "notificationConfigurationID", func() error { _, err := c.NotificationConfigurations.GetByID(ctx, ""); return err }},
		{"NotificationConfigurations.UpdateFieldsByID", "notificationConfigurationID", func() error {
			_, err := c.NotificationConfigurations.UpdateFieldsByID(ctx, "", testPatch)
			return err
		}},
		{"NotificationConfigurations.DeleteByID", "notificationConfigurationID", func() error { return c.NotificationConfigurations.DeleteByID(ctx, "") }},
		{"PolicySetParameters.ListByPolicySet", "policySetID", func() error { _, err := c.PolicySetParameters.ListByPolicySet(ctx, "", nil); return err }},
		{"PolicySetParameters.GetByPolicySet/policySet", "policySetID", func() error { _, err := c.PolicySetParameters.GetByPolicySet(ctx, "", "id-1"); return err }},
		{"PolicySetParameters.GetByPolicySet/id", "policySetParameterID", func() error { _, err := c.PolicySetParameters.GetByPolicySet(ctx, "ps-1", ""); return err }},
		{"PolicySetParameters.CreateFieldsByPolicySet", "policySetID", func() error {
			_, err := c.PolicySetParameters.CreateFieldsByPolicySet(ctx, "", testPatch)
			return err
		}},
		{"PolicySetParameters.UpdateFieldsByPolicySet", "policySetParameterID", func() error {
			_, err := c.PolicySetParameters.UpdateFieldsByPolicySet(ctx, "ps-1", "", testPatch)
			return err
		}},
		{"PolicySetParameters.DeleteByPolicySet", "policySetParameterID", func() error { return c.PolicySetParameters.DeleteByPolicySet(ctx, "ps-1", "") }},
		{"PolicyAttachments.ListByPolicySet", "policySetID", func() error { _, err := c.PolicyAttachments.ListByPolicySet(ctx, "", nil); return err }},
		{"PolicyAttachments.GetByPolicySet/id", "policyAttachmentID", func() error { _, err := c.PolicyAttachments.GetByPolicySet(ctx, "ps-1", ""); return err }},
		{"PolicyAttachments.CreateFieldsByPolicySet", "policySetID", func() error {
			_, err := c.PolicyAttachments.CreateFieldsByPolicySet(ctx, "", testPatch)
			return err
		}},
		{"PolicyAttachments.UpdateFieldsByPolicySet", "policySetID", func() error {
			_, err := c.PolicyAttachments.UpdateFieldsByPolicySet(ctx, "", "id-1", testPatch)
			return err
		}},
		{"PolicyAttachments.DeleteByPolicySet", "policyAttachmentID", func() error { return c.PolicyAttachments.DeleteByPolicySet(ctx, "ps-1", "") }},
		{"NotificationTriggers.ListByConfiguration", "notificationConfigurationID", func() error {
			_, err := c.NotificationTriggers.ListByConfiguration(ctx, "", nil)
			return err
		}},
		{"NotificationTriggers.CreateFieldsByConfiguration", "notificationConfigurationID", func() error {
			_, err := c.NotificationTriggers.CreateFieldsByConfiguration(ctx, "", testPatch)
			return err
		}},
		{"NotificationTriggers.DeleteByConfiguration", "notificationTriggerID", func() error {
			return c.NotificationTriggers.DeleteByConfiguration(ctx, "cfg-1", "")
		}},
		{"PolicyAttachments.GetByPolicySet/policySet", "policySetID", func() error { _, err := c.PolicyAttachments.GetByPolicySet(ctx, "", "id-1"); return err }},
		{"PolicyAttachments.UpdateFieldsByPolicySet/id", "policyAttachmentID", func() error {
			_, err := c.PolicyAttachments.UpdateFieldsByPolicySet(ctx, "ps-1", "", testPatch)
			return err
		}},
		{"PolicyAttachments.DeleteByPolicySet/policySet", "policySetID", func() error { return c.PolicyAttachments.DeleteByPolicySet(ctx, "", "id-1") }},
		{"PolicySetParameters.UpdateFieldsByPolicySet/policySet", "policySetID", func() error {
			_, err := c.PolicySetParameters.UpdateFieldsByPolicySet(ctx, "", "id-1", testPatch)
			return err
		}},
		{"PolicySetParameters.DeleteByPolicySet/policySet", "policySetID", func() error { return c.PolicySetParameters.DeleteByPolicySet(ctx, "", "id-1") }},
		{"NotificationTriggers.DeleteByConfiguration/config", "notificationConfigurationID", func() error {
			return c.NotificationTriggers.DeleteByConfiguration(ctx, "", "id-1")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidationError(t, tt.call(), tt.field)
		})
	}
}

func TestByID_List_Options(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/policy_set/ps-1/parameters", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filter[policy_set_parameter]"); got != "key==x" {
			t.Errorf("filter[policy_set_parameter] = %q, want %q", got, "key==x")
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.PolicySetParameter{})
	})
	c := newTestClient(t, srv)

	if _, err := c.PolicySetParameters.ListByPolicySet(context.Background(), "ps-1", &terrakube.ListOptions{Filter: "key==x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.HandleFunc("GET /api/v1/runTrigger", func(w http.ResponseWriter, r *http.Request) {
		want := `sourceWorkspace.id=="s";destinationWorkspace.id=="d"`
		if got := r.URL.Query().Get("filter[runTrigger]"); got != want {
			t.Errorf("filter[runTrigger] = %q, want %q", got, want)
		}
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.RunTrigger{})
	})
	if _, err := c.RunTriggers.ListTopLevel(context.Background(), &terrakube.ListOptions{Filter: `sourceWorkspace.id=="s";destinationWorkspace.id=="d"`}); err != nil {
		t.Fatalf("ListTopLevel: unexpected error: %v", err)
	}

	// These collections are not paged, so a page request is refused before it is sent.
	_, err := c.PolicyAttachments.ListByPolicySet(context.Background(), "ps-1", &terrakube.ListOptions{Page: 1})
	if !isValidationErr(err) {
		t.Fatalf("ListByPolicySet with a page: want a *ValidationError, got %v", err)
	}
}

func isValidationErr(err error) bool {
	_, ok := errors.AsType[*terrakube.ValidationError](err)
	return ok
}

// idOf returns the ID of a result, or the error that came with it.
func idOf[T any](v *T, err error, id func() string) (string, error) {
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	return id(), nil
}
