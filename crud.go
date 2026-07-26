package terrakube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/google/jsonapi"
)

// crudService is a generic base for JSON:API CRUD operations.
// It is embedded by resource-specific services.
type crudService[T any] struct {
	client    *Client
	filterKey string // query param key for list filtering; defaults to "filter" when empty
}

// list retrieves a collection of resources at the given path, optionally filtered.
// When ListOptions.Page and ListOptions.PageSize are both zero, all pages are
// fetched automatically. Otherwise a single page is returned.
func (s *crudService[T]) list(ctx context.Context, path string, opts *ListOptions) ([]*T, error) {
	params := s.buildListParams(opts)

	// Manual pagination: fetch a single page.
	if opts != nil && (opts.Page > 0 || opts.PageSize > 0) {
		req, err := s.client.requestWithQuery(ctx, http.MethodGet, path, params, nil)
		if err != nil {
			return nil, err
		}
		return doList[T](s.client, req)
	}

	// Auto-paginate: fetch all pages.
	return s.fetchAllPages(ctx, path, params)
}

// buildListParams converts ListOptions into URL query parameters.
func (s *crudService[T]) buildListParams(opts *ListOptions) url.Values {
	params := url.Values{}
	if opts == nil {
		return params
	}

	if opts.Filter != "" {
		key := s.filterKey
		if key == "" {
			key = "filter"
		}
		params.Set(key, opts.Filter)
	}
	if opts.Include != "" {
		params.Set("include", opts.Include)
	}
	if opts.Page > 0 {
		params.Set("page[number]", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 {
		params.Set("page[size]", strconv.Itoa(opts.PageSize))
	}

	return params
}

// fetchAllPages iterates through all pages by following JSON:API links.next.
func (s *crudService[T]) fetchAllPages(ctx context.Context, path string, params url.Values) ([]*T, error) {
	var allItems []*T

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, err
	}

	for {
		items, nextURL, err := doListWithLinks[T](s.client, req)
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, items...)

		if nextURL == "" {
			break
		}

		// Build request for the next page using the absolute URL from links.next.
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building next page request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+s.client.token)
		req.Header.Set("User-Agent", s.client.userAgent)
		req.Header.Set("Accept", mediaType)
	}

	return allItems, nil
}

// doListWithLinks executes a request and decodes a JSON:API list response,
// returning both the items and the next page URL (if present).
func doListWithLinks[T any](c *Client, req *http.Request) ([]*T, string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are inconsequential

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", c.parseAPIError(req, resp.StatusCode, bodyBytes)
	}

	if len(bodyBytes) == 0 {
		return nil, "", nil
	}

	items, err := jsonapi.UnmarshalManyPayload(bytes.NewReader(bodyBytes), reflect.TypeOf((*T)(nil)))
	if err != nil {
		return nil, "", fmt.Errorf("decoding JSON:API list response: %w", err)
	}

	result := make([]*T, len(items))
	for i, item := range items {
		result[i] = item.(*T)
	}

	// Extract links.next for pagination.
	var links struct {
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	if json.Unmarshal(bodyBytes, &links) == nil {
		return result, links.Links.Next, nil
	}

	return result, "", nil
}

// get retrieves a single resource at the given path.
func (s *crudService[T]) get(ctx context.Context, path string) (*T, error) {
	req, err := s.client.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	result := new(T)
	_, err = s.client.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// create posts a new resource to the given path.
func (s *crudService[T]) create(ctx context.Context, path string, entity *T) (*T, error) {
	req, err := s.client.request(ctx, http.MethodPost, path, entity)
	if err != nil {
		return nil, err
	}

	created := new(T)
	_, err = s.client.do(req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// update patches an existing resource at the given path.
func (s *crudService[T]) update(ctx context.Context, path string, entity *T) (*T, error) {
	req, err := s.client.request(ctx, http.MethodPatch, path, entity)
	if err != nil {
		return nil, err
	}

	updated := new(T)
	_, err = s.client.do(req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// del removes the resource at the given path.
func (s *crudService[T]) del(ctx context.Context, path string) error {
	req, err := s.client.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
