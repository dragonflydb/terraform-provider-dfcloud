package sdk

import (
	"context"
	"errors"
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
