package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(
		WithAPIKey("test-key"),
		WithAPIHost(strings.TrimPrefix(srv.URL, "https://")),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.httpClient = srv.Client()

	return client
}

func TestGetNetwork(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/network-2" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"network_id":"network-2","name":"two","location":{"provider":"gcp","region":"us-central1"},"cidr_block":"10.1.0.0/16","status":"active","created_at":2}`))
	}))

	got, err := client.GetNetwork(context.Background(), "network-2")
	if err != nil {
		t.Fatalf("GetNetwork() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetNetwork() returned nil network")
	}
	if got.ID != "network-2" {
		t.Fatalf("GetNetwork() ID = %q, want %q", got.ID, "network-2")
	}
	if got.Name != "two" {
		t.Fatalf("GetNetwork() Name = %q, want %q", got.Name, "two")
	}
}

func TestGetNetworkReturnsNotFound(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/missing-network" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))

	_, err := client.GetNetwork(context.Background(), "missing-network")
	if err == nil {
		t.Fatal("GetNetwork() error = nil, want ErrNotFound")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetNetwork() error = %v, want ErrNotFound", err)
	}
}

// captureDatastoreRequest returns a client whose server records the decoded
// JSON body of a datastore create request and replies with the given response.
func captureDatastoreRequest(t *testing.T, response string) (*Client, *map[string]any) {
	t.Helper()

	var body map[string]any
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/datastores" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))

	return client, &body
}

func TestCreateDatastoreSendsFlagPolicyID(t *testing.T) {
	client, body := captureDatastoreRequest(t, `{"datastore_id":"ds-1","status":"pending","config":{"name":"one","flag_policy_id":"fp-1"}}`)

	got, err := client.CreateDatastore(context.Background(), &DatastoreConfig{
		Name:         "one",
		FlagPolicyID: "fp-1",
	})
	if err != nil {
		t.Fatalf("CreateDatastore() error = %v", err)
	}

	if (*body)["flag_policy_id"] != "fp-1" {
		t.Fatalf("request flag_policy_id = %v, want %q", (*body)["flag_policy_id"], "fp-1")
	}
	if got.Config.FlagPolicyID != "fp-1" {
		t.Fatalf("CreateDatastore() Config.FlagPolicyID = %q, want %q", got.Config.FlagPolicyID, "fp-1")
	}
}

func TestCreateDatastoreOmitsUnsetFlagPolicyID(t *testing.T) {
	client, body := captureDatastoreRequest(t, `{"datastore_id":"ds-1","status":"pending","config":{"name":"one"}}`)

	got, err := client.CreateDatastore(context.Background(), &DatastoreConfig{Name: "one"})
	if err != nil {
		t.Fatalf("CreateDatastore() error = %v", err)
	}

	if v, ok := (*body)["flag_policy_id"]; ok {
		t.Fatalf("request flag_policy_id = %v, want field omitted", v)
	}
	if got.Config.FlagPolicyID != "" {
		t.Fatalf("CreateDatastore() Config.FlagPolicyID = %q, want empty", got.Config.FlagPolicyID)
	}
}
