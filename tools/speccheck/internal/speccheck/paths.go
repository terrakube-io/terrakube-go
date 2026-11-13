package speccheck

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// APIBasePath is the prefix the client puts before every JSON:API path.
const APIBasePath = "/api/v1"

// Recorder is an http.RoundTripper that records request paths and answers
// every request with 404 without touching the network. Building a client with
// it lets ValidatePaths learn the exact paths each service sends, whatever
// nesting the service uses.
type Recorder struct {
	mu    sync.Mutex
	paths []string
}

// RoundTrip records the request path and returns an empty 404 response.
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.paths = append(r.paths, req.URL.Path)
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header:     http.Header{"Content-Type": []string{"application/vnd.api+json"}},
		Body:       io.NopCloser(strings.NewReader(`{"errors":[]}`)),
		Request:    req,
	}, nil
}

// take returns and clears the recorded paths.
func (r *Recorder) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := r.paths
	r.paths = nil
	return p
}

// ServicePaths are the request paths one service sends, relative to the API
// base path, with placeholder IDs id0, id1, ...
type ServicePaths struct {
	Collection string // path of List; empty when the service has no List
	Item       string // path of Get
}

// CapturePaths calls the List and Get methods of the service field name on
// client with placeholder IDs and returns the paths the service requested.
// client must have been built with an HTTP client that uses rec.
func CapturePaths(client any, name string, rec *Recorder) (ServicePaths, error) {
	svc := reflect.ValueOf(client).Elem().FieldByName(name)
	if !svc.IsValid() || svc.IsNil() {
		return ServicePaths{}, fmt.Errorf("speccheck: client has no service %s", name)
	}
	var list string
	if svc.MethodByName("List").IsValid() {
		var err error
		if list, err = callAndCapture(svc, "List", rec); err != nil {
			return ServicePaths{}, err
		}
	}
	item, err := callAndCapture(svc, "Get", rec)
	if err != nil {
		return ServicePaths{}, err
	}
	return ServicePaths{Collection: list, Item: item}, nil
}

// callAndCapture calls method with a context, one placeholder ID per string
// parameter, and zero values for anything else (nil *ListOptions), then
// returns the single path it requested.
func callAndCapture(svc reflect.Value, method string, rec *Recorder) (string, error) {
	m := svc.MethodByName(method)
	if !m.IsValid() {
		return "", fmt.Errorf("speccheck: %s has no %s method", svc.Type(), method)
	}
	mt := m.Type()
	args := make([]reflect.Value, mt.NumIn())
	ids := 0
	for i := range mt.NumIn() {
		in := mt.In(i)
		switch {
		case in.Implements(reflect.TypeFor[context.Context]()):
			args[i] = reflect.ValueOf(context.Background())
		case in.Kind() == reflect.String:
			args[i] = reflect.ValueOf("id" + strconv.Itoa(ids)).Convert(in)
			ids++
		default:
			args[i] = reflect.Zero(in)
		}
	}
	rec.take()
	m.Call(args)
	paths := rec.take()
	if len(paths) != 1 {
		return "", fmt.Errorf("speccheck: %s.%s sent %d requests, want 1", svc.Type(), method, len(paths))
	}
	return strings.TrimPrefix(paths[0], APIBasePath), nil
}

// ValidatePaths checks that the collection and item paths a service sends
// resolve, in the spec, to the model's resource type: a collection for List
// and a single resource for Get.
func ValidatePaths(spec *Spec, m ModelInfo, sp ServicePaths) []Issue {
	var issues []Issue
	for _, p := range []struct {
		concrete   string
		collection bool
	}{{sp.Collection, true}, {sp.Item, false}} {
		if p.concrete == "" {
			continue
		}
		typ, collection, via, ok := spec.ResolvePath(p.concrete)
		if !ok {
			issues = append(issues, Issue{Kind: PathMissing, Detail: p.concrete})
			continue
		}
		if typ != m.ResourceType || collection != p.collection {
			issues = append(issues, Issue{
				Kind:     PathTypeMismatch,
				GoType:   describeTarget(m.ResourceType, p.collection),
				SpecType: describeTarget(typ, collection),
				Detail:   p.concrete + " (resolved by " + via + ")",
			})
		}
	}
	return issues
}

func describeTarget(typ string, collection bool) string {
	if collection {
		return "collection of " + typ
	}
	return typ
}
