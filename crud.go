package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// crudService is a generic base for JSON:API CRUD operations.
// It is embedded by resource-specific services.
type crudService[T any] struct {
	client    *Client
	filterKey string // query param key for list filtering; defaults to "filter" when empty
}

// list retrieves a collection of resources at the given path, optionally filtered.
func (s *crudService[T]) list(ctx context.Context, path string, opts *ListOptions) ([]*T, error) {
	var req *http.Request
	var err error

	if opts != nil && opts.Filter != "" {
		key := s.filterKey
		if key == "" {
			key = "filter"
		}
		params := url.Values{key: {opts.Filter}}
		req, err = s.client.requestWithQuery(ctx, http.MethodGet, path, params, nil)
	} else {
		req, err = s.client.request(ctx, http.MethodGet, path, nil)
	}
	if err != nil {
		return nil, err
	}

	var items []*T
	_, err = s.client.do(ctx, req, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// get retrieves a single resource at the given path.
func (s *crudService[T]) get(ctx context.Context, path string) (*T, error) {
	req, err := s.client.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	result := new(T)
	_, err = s.client.do(ctx, req, result)
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
	_, err = s.client.do(ctx, req, created)
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
	_, err = s.client.do(ctx, req, updated)
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

	_, err = s.client.do(ctx, req, nil)
	return err
}
