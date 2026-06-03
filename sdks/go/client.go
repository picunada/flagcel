package flagcel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/picunada/flagcel/evalcore"
)

type definitionsHTTPClient struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type definitionsEnvelope struct {
	Data evalcore.Definitions `json:"data"`
}

type fetchResult struct {
	definitions evalcore.Definitions
	etag        string
	unchanged   bool
}

func newDefinitionsHTTPClient(endpoint, apiKey string, httpClient *http.Client) (*definitionsHTTPClient, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("flagcel: endpoint is required")
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("flagcel: parse endpoint: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("flagcel: endpoint must be an absolute URL")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &definitionsHTTPClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}, nil
}

func (c *definitionsHTTPClient) fetchDefinitions(ctx context.Context, etag string) (fetchResult, error) {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: joinURLPath(c.baseURL.Path, "/eval/definitions")})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fetchResult{}, fmt.Errorf("flagcel: create definitions request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fetchResult{}, fmt.Errorf("flagcel: fetch definitions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return fetchResult{etag: responseETag(resp, etag), unchanged: true}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fetchResult{}, fmt.Errorf("flagcel: fetch definitions: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope definitionsEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fetchResult{}, fmt.Errorf("flagcel: decode definitions: %w", err)
	}

	return fetchResult{
		definitions: envelope.Data,
		etag:        responseETag(resp, etag),
	}, nil
}

func responseETag(resp *http.Response, previous string) string {
	if etag := resp.Header.Get("ETag"); etag != "" {
		return etag
	}
	return previous
}

func joinURLPath(basePath, path string) string {
	basePath = strings.TrimRight(basePath, "/")
	if basePath == "" {
		return path
	}
	return basePath + path
}
