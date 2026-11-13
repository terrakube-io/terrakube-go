package terrakube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// crudService is a generic base for JSON:API CRUD operations.
// It is embedded by resource-specific services.
type crudService[T any] struct {
	client    *Client
	filterKey string // query param key for list filtering; defaults to "filter" when empty
	paged     bool   // the server accepts page parameters for this resource type
}

// list retrieves a collection of resources at the given path, optionally filtered.
// With ListOptions.Page or ListOptions.PageSize set, that one page is returned.
// Otherwise the collection is listed: completely and page by page for resource
// types the server can page, and in a single request for all others. Every
// request goes to the configured endpoint; links in responses are not followed.
func (s *crudService[T]) list(ctx context.Context, path string, opts *ListOptions) ([]*T, error) {
	params := s.buildListParams(opts)

	if !s.paged || (opts != nil && (opts.Page > 0 || opts.PageSize > 0)) {
		items, _, err := s.listPage(ctx, path, params)
		return items, err
	}
	return s.listAll(ctx, path, params)
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

// listPage sends one list request and returns its items and meta.page.
func (s *crudService[T]) listPage(ctx context.Context, path string, params url.Values) ([]*T, *pageInfo, error) {
	req, err := s.client.requestWithQuery(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, nil, err
	}
	return s.client.doList[T](req)
}

// listAll fetches every page of a collection using the server's meta.page.
// The first request asks only for totals, so the server picks the page size;
// later pages repeat that size.
func (s *crudService[T]) listAll(ctx context.Context, path string, params url.Values) ([]*T, error) {
	params.Set("page[totals]", "")
	all, page, err := s.listPage(ctx, path, params)
	if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.StatusCode == http.StatusBadRequest {
		// Which resource types a server can page depends on its version.
		// If it refuses page parameters, list the collection in one request.
		params.Del("page[totals]")
		all, _, err = s.listPage(ctx, path, params)
		return all, err
	}
	if err != nil {
		return nil, err
	}

	for n := 2; page != nil && page.Limit > 0 && n <= page.TotalPages; n++ {
		params.Set("page[number]", strconv.Itoa(n))
		params.Set("page[size]", strconv.Itoa(page.Limit))

		var items []*T
		items, page, err = s.listPage(ctx, path, params)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
	}
	return all, nil
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
