package terrakube_test

import (
	"context"
	"net/http"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/terrakube-io/terrakube-go/testutil"
)

func TestFederatedClaimService_List(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/federated/fed-1/claims", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPIList(t, w, http.StatusOK, []*terrakube.FederatedClaim{
			{ID: "claim-1", ClaimKey: "sub", ClaimValue: "repo:org/repo:ref:refs/heads/main"},
			{ID: "claim-2", ClaimKey: "repository", ClaimValue: "org/repo"},
		})
	})

	client := newTestClient(t, srv)
	claims, err := client.FederatedClaims.List(context.Background(), "fed-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(claims) != 2 {
		t.Fatalf("got %d claims, want 2", len(claims))
	}
	if claims[0].ClaimKey != "sub" {
		t.Errorf("ClaimKey = %q, want %q", claims[0].ClaimKey, "sub")
	}
}

func TestFederatedClaimService_Get(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("GET /api/v1/federated/fed-1/claims/claim-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.FederatedClaim{
			ID: "claim-1", ClaimKey: "sub", ClaimValue: "repo:org/repo:ref:refs/heads/main",
		})
	})

	client := newTestClient(t, srv)
	claim, err := client.FederatedClaims.Get(context.Background(), "fed-1", "claim-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim.ID != "claim-1" {
		t.Errorf("ID = %q, want %q", claim.ID, "claim-1")
	}
	if claim.ClaimValue != "repo:org/repo:ref:refs/heads/main" {
		t.Errorf("ClaimValue = %q, want %q", claim.ClaimValue, "repo:org/repo:ref:refs/heads/main")
	}
}

func TestFederatedClaimService_Create(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("POST /api/v1/federated/fed-1/claims", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusCreated, &terrakube.FederatedClaim{
			ID: "claim-new", ClaimKey: "aud", ClaimValue: "terrakube",
		})
	})

	client := newTestClient(t, srv)
	claim, err := client.FederatedClaims.Create(context.Background(), "fed-1", &terrakube.FederatedClaim{
		ClaimKey: "aud", ClaimValue: "terrakube",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim.ID != "claim-new" {
		t.Errorf("ID = %q, want %q", claim.ID, "claim-new")
	}
}

func TestFederatedClaimService_Update(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("PATCH /api/v1/federated/fed-1/claims/claim-1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.WriteJSONAPI(t, w, http.StatusOK, &terrakube.FederatedClaim{
			ID: "claim-1", ClaimKey: "sub", ClaimValue: "repo:org/repo:environment:prod",
		})
	})

	client := newTestClient(t, srv)
	claim, err := client.FederatedClaims.Update(context.Background(), "fed-1", &terrakube.FederatedClaim{
		ID: "claim-1", ClaimKey: "sub", ClaimValue: "repo:org/repo:environment:prod",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim.ClaimValue != "repo:org/repo:environment:prod" {
		t.Errorf("ClaimValue = %q, want %q", claim.ClaimValue, "repo:org/repo:environment:prod")
	}
}

func TestFederatedClaimService_Delete(t *testing.T) {
	t.Parallel()

	srv := testutil.NewServer(t)
	srv.HandleFunc("DELETE /api/v1/federated/fed-1/claims/claim-1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, srv)
	err := client.FederatedClaims.Delete(context.Background(), "fed-1", "claim-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
