package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type errorResponse struct {
	Error string `json:"error"`
}

type clientOptions struct {
	apiKey  string
	apiHost string
	timeout time.Duration
}

type ClientOption interface {
	apply(*clientOptions)
}

type apiKeyOption string

func (o apiKeyOption) apply(opts *clientOptions) {
	opts.apiKey = string(o)
}

// WithAPIKey configures the client to authenticate with Dragonfly cloud using
// the given API key.
//
// You can get an API key from the Dragonfly cloud dashboard.
func WithAPIKey(key string) ClientOption {
	return apiKeyOption(key)
}

// WithAPIKeyFromEnv is a shortcut for calling [WithAPIKey] with the
// value of the DFCLOUD_API_KEY environment variable.
func WithAPIKeyFromEnv() ClientOption {
	return WithAPIKey(os.Getenv("DFCLOUD_API_KEY"))
}

type timeoutOption time.Duration

func (o timeoutOption) apply(opts *clientOptions) {
	opts.timeout = time.Duration(o)
}

// WithTimeout configures the client request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return timeoutOption(timeout)
}

type apiHostOption string

func (o apiHostOption) apply(opts *clientOptions) {
	opts.apiHost = string(o)
}

// WithAPIHost configures the client to use the given API URL.
func WithAPIHost(url string) ClientOption {
	return apiHostOption(url)
}

// Client represents a REST client for the Dragonfly cloud API.
type Client struct {
	apiKey  string
	apiHost string

	httpClient *http.Client
}

// NewClient creates a Dragonfly cloud client.
//
// The client options must include either [WithAPIKey] or [WithAPIKeyFromEnv]
// to authenticate with Dragonfly cloud.
func NewClient(opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		timeout: time.Second * 15,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	if options.apiKey == "" {
		return nil, fmt.Errorf("missing api key")
	}

	if options.apiHost == "" {
		// use default
		options.apiHost = "api.dragonflydb.cloud"
	}

	return &Client{
		apiKey: options.apiKey,
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
		apiHost: options.apiHost,
	}, nil
}

func (c *Client) GetDatastore(ctx context.Context, id string) (*Datastore, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/datastores/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var datastore *Datastore
	if err := json.NewDecoder(r).Decode(&datastore); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if datastore.Key == "" {
		// only way to disable passkey
		datastore.Config.DisablePasskey = true
	}

	return datastore, nil
}

func (c *Client) CreateDatastore(ctx context.Context, config *DatastoreConfig) (*Datastore, error) {
	b, _ := json.Marshal(&config)

	r, err := c.request(ctx, http.MethodPost, "/v1/datastores", b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var datastore Datastore
	if err := json.NewDecoder(r).Decode(&datastore); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if datastore.Key == "" {
		// only way to disable passkey
		datastore.Config.DisablePasskey = true
	}

	return &datastore, nil
}

func (c *Client) UpdateDatastore(ctx context.Context, id string, config *DatastoreConfig) (*Datastore, error) {
	b, _ := json.Marshal(&config)

	r, err := c.request(ctx, http.MethodPut, "/v1/datastores/"+id, b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var datastore Datastore
	if err := json.NewDecoder(r).Decode(&datastore); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if datastore.Key == "" {
		// only way to disable passkey
		datastore.Config.DisablePasskey = true
	}
	return &datastore, nil
}

// ListDatastores lists all the customers datastores.
func (c *Client) ListDatastores(ctx context.Context) ([]*Datastore, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/datastores", nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var datastores []*Datastore
	if err := json.NewDecoder(r).Decode(&datastores); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return datastores, nil
}

func (c *Client) DeleteDatastore(ctx context.Context, id string) error {
	r, err := c.request(ctx, http.MethodDelete, "/v1/datastores/"+id, nil)
	if err != nil {
		return err
	}
	defer r.Close()

	return nil
}

func (c *Client) GetNetwork(ctx context.Context, id string) (*Network, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/networks/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var network *Network
	if err := json.NewDecoder(r).Decode(&network); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return network, nil
}

func (c *Client) CreateNetwork(ctx context.Context, config *NetworkConfig) (*Network, error) {
	b, _ := json.Marshal(&config)

	r, err := c.request(ctx, http.MethodPost, "/v1/networks", b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var network Network
	if err := json.NewDecoder(r).Decode(&network); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &network, nil
}

// ListNetworks lists all the customers networks.
func (c *Client) ListNetworks(ctx context.Context) ([]*Network, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/networks", nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var networks []*Network
	if err := json.NewDecoder(r).Decode(&networks); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return networks, nil
}

func (c *Client) DeleteNetwork(ctx context.Context, id string) error {
	r, err := c.request(ctx, http.MethodDelete, "/v1/networks/"+id, nil)
	if err != nil {
		return err
	}
	defer r.Close()

	return nil
}

func (c *Client) GetConnection(ctx context.Context, id string) (*Connection, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/connections/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var conn *Connection
	if err := json.NewDecoder(r).Decode(&conn); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return conn, nil
}

func (c *Client) CreateConnection(ctx context.Context, config *ConnectionConfig) (*Connection, error) {
	b, _ := json.Marshal(&config)

	r, err := c.request(ctx, http.MethodPost, "/v1/connections", b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var conn Connection
	if err := json.NewDecoder(r).Decode(&conn); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &conn, nil
}

// ListConnection lists all the customers connections.
func (c *Client) ListConnections(ctx context.Context) ([]*Connection, error) {
	r, err := c.request(ctx, http.MethodGet, "/v1/connections", nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var conns []*Connection
	if err := json.NewDecoder(r).Decode(&conns); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return conns, nil
}

func (c *Client) DeleteConnection(ctx context.Context, id string) error {
	r, err := c.request(ctx, http.MethodDelete, "/v1/connections/"+id, nil)
	if err != nil {
		return err
	}
	defer r.Close()

	return nil
}

func (c *Client) request(
	ctx context.Context,
	method string,
	path string,
	body []byte,
) (io.ReadCloser, error) {
	url := &url.URL{
		Scheme: "https",
		Host:   c.apiHost,
		Path:   path,
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), b)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest &&
			resp.StatusCode < http.StatusInternalServerError {
			var errResp errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
				return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
			}
			return nil, fmt.Errorf("bad status: %d: %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
