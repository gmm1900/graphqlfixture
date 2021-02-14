package graphqlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Request represents the request to GraphQL endpoint.
type Request struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// Client is a GraphQL client.
type Client struct {
	url           string // GraphQL server URL.
	httpClient    *http.Client
	customHeaders http.Header
}

const (
	defaultHTTPTimeout = 30 * time.Second
)

// New creates a GraphQL client targeting the specified GraphQL server URL.
// If httpClient is nil, then http.DefaultClient with timeout is used.
func New(url string, httpClient *http.Client, customHeaders http.Header) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultHTTPTimeout,
		}
	}

	return &Client{
		url:           url,
		httpClient:    httpClient,
		customHeaders: customHeaders,
	}
}

// Do sends the graphql request and json-unmarshal the graphql response to the
// given `response` object (expect to be a pointer of the desired response type).  The given
// `response` object is expected to be able to handle the `Data` and `Errors` top level fields.
func (c *Client) Do(ctx context.Context, graphqlRequest Request, response interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(graphqlRequest)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.url, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	c.customHeaders.Set("Content-Type", "application/json")
	req.Header = c.customHeaders

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("bad response status code: %v body: %q", resp.Status, body)
	}

	switch t := response.(type) {
	case *[]byte:
		body, _ := ioutil.ReadAll(resp.Body)
		*t = body
	case *string:
		body, _ := ioutil.ReadAll(resp.Body)
		*t = string(body)
	default:
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return fmt.Errorf("error decoding response body: %w", err)
		}
	}

	return nil
}
