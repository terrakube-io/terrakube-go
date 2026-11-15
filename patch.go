package terrakube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/google/jsonapi"
)

// Patch is a sparse JSON:API write. Only the listed attributes and
// relationships are sent, so the server keeps its value (on update) or its
// default (on create) for everything else.
//
// Attribute keys are the JSON:API attribute names, as in the struct tags
// (for example "terraformVersion" for Workspace.IaCVersion). Values are
// encoded with encoding/json; a nil value sends null.
type Patch struct {
	Attributes    map[string]any
	Relationships map[string]Relationship
}

// Relationship is a relationship in a Patch.
//
// A to-one relationship sets ID: the target's ID, or nil to clear the link
// ({"data": null}). A to-many relationship sets Many and lists the targets in
// IDs; the list replaces the whole set, so an empty list clears it
// ({"data": []}). Type is required except to clear a to-one link.
//
//	"vcs":       {Type: "vcs", ID: &vcsID}                        // link
//	"vcs":       {}                                               // unlink
//	"templates": {Type: "template", Many: true, IDs: []string{a, b}} // replace the set
//	"templates": {Type: "template", Many: true}                   // empty the set
type Relationship struct {
	Type string   // JSON:API type of the target, for example "vcs".
	ID   *string  // To-one: ID of the target; nil clears the link.
	Many bool     // The relationship is to-many: IDs is sent and ID must be nil.
	IDs  []string // To-many: IDs of all targets; empty clears the set.
}

// patchResource is the resource object of a sparse write.
type patchResource struct {
	Type          string                     `json:"type"`
	ID            string                     `json:"id,omitempty"`
	Attributes    map[string]any             `json:"attributes,omitempty"`
	Relationships map[string]patchRelLinkage `json:"relationships,omitempty"`
}

// patchRelLinkage is a relationship object. Data is nil (encoded as null), a
// *patchRelData or a []patchRelData.
type patchRelLinkage struct {
	Data any `json:"data"`
}

type patchRelData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// isEmpty reports whether the patch names no attribute and no relationship.
func (p Patch) isEmpty() bool {
	return len(p.Attributes) == 0 && len(p.Relationships) == 0
}

// resource builds the resource object for p, validating its relationships in
// name order, so the same Patch always reports the same error.
func (p Patch) resource(jsonapiType, id string) (*patchResource, error) {
	res := &patchResource{Type: jsonapiType, ID: id, Attributes: p.Attributes}
	if len(p.Relationships) > 0 {
		res.Relationships = make(map[string]patchRelLinkage, len(p.Relationships))
	}
	for _, name := range slices.Sorted(maps.Keys(p.Relationships)) {
		linkage, err := p.Relationships[name].linkage(name)
		if err != nil {
			return nil, err
		}
		res.Relationships[name] = linkage
	}
	return res, nil
}

// linkage validates rel and returns its relationship object.
func (rel Relationship) linkage(name string) (patchRelLinkage, error) {
	field := "Patch.Relationships[" + name + "]"
	switch {
	case name == "":
		return patchRelLinkage{}, &ValidationError{Field: "Patch.Relationships", Message: "must not contain an empty name"}
	case rel.Many && rel.ID != nil:
		return patchRelLinkage{}, &ValidationError{Field: field, Message: "must not set both ID (to-one) and Many (to-many)"}
	case !rel.Many && len(rel.IDs) > 0:
		return patchRelLinkage{}, &ValidationError{Field: field + ".IDs", Message: "must be empty unless Many is set"}
	case !rel.Many && rel.ID == nil:
		return patchRelLinkage{}, nil // clears a to-one link
	case rel.Type == "":
		return patchRelLinkage{}, &ValidationError{Field: field + ".Type", Message: "must not be empty"}
	case !rel.Many && *rel.ID == "":
		return patchRelLinkage{}, &ValidationError{Field: field + ".ID", Message: "must not be empty; use a nil ID to clear the link"}
	case !rel.Many:
		return patchRelLinkage{Data: &patchRelData{Type: rel.Type, ID: *rel.ID}}, nil
	}
	data := make([]patchRelData, len(rel.IDs))
	for i, id := range rel.IDs {
		if id == "" {
			return patchRelLinkage{}, &ValidationError{Field: fmt.Sprintf("%s.IDs[%d]", field, i), Message: "must not be empty"}
		}
		data[i] = patchRelData{Type: rel.Type, ID: id}
	}
	return patchRelLinkage{Data: data}, nil
}

// requestDocument builds an authenticated JSON:API request whose body is the
// resource object res, encoded with encoding/json.
func (c *Client) requestDocument(ctx context.Context, method, reqPath string, res *patchResource) (*http.Request, error) {
	body, err := json.Marshal(map[string]any{"data": res})
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpointURL(reqPath).String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Accept", mediaType)
	return req, nil
}

// jsonapiTypeOf returns the JSON:API type in the primary tag of T.
func jsonapiTypeOf[T any]() string {
	for field := range reflect.TypeFor[T]().Fields() {
		if name, ok := strings.CutPrefix(field.Tag.Get("jsonapi"), "primary,"); ok {
			name, _, _ = strings.Cut(name, ",")
			return name
		}
	}
	return ""
}

// updateFields sends a sparse PATCH of the resource id at path and returns its
// new state. An empty Patch is a *ValidationError. The server answers a PATCH
// with 204 and no body, so the state is then read with a GET of the same path.
// If the update made the resource unreadable (a disabled organization answers
// 404), the attributes that were sent are returned.
func (s *crudService[T]) updateFields(ctx context.Context, path, id string, p Patch) (*T, error) {
	if p.isEmpty() {
		return nil, &ValidationError{Field: "Patch", Message: "must set at least one attribute or relationship"}
	}
	res, err := p.resource(jsonapiTypeOf[T](), id)
	if err != nil {
		return nil, err
	}
	req, err := s.client.requestDocument(ctx, http.MethodPatch, path, res)
	if err != nil {
		return nil, err
	}

	updated := new(T)
	resp, err := s.client.do(req, updated)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
		fresh, err := s.get(ctx, path)
		if IsNotFound(err) {
			return sentEntity[T](res), nil // the update itself succeeded
		}
		return fresh, err
	}
	return updated, nil
}

// sentEntity decodes the sent attributes and id into a T. If the attributes do
// not decode into T, only the id is kept.
func sentEntity[T any](res *patchResource) *T {
	out := new(T)
	sent := map[string]any{"type": res.Type, "id": res.ID, "attributes": res.Attributes}
	if b, err := json.Marshal(map[string]any{"data": sent}); err == nil && jsonapi.UnmarshalPayload(bytes.NewReader(b), out) == nil {
		return out
	}
	out = new(T)
	b, _ := json.Marshal(map[string]any{"data": map[string]any{"type": res.Type, "id": res.ID}}) //nolint:errchkjson // two strings always encode
	_ = jsonapi.UnmarshalPayload(bytes.NewReader(b), out)                                        // an id-only resource object always decodes
	return out
}

// createFields sends a sparse POST to path and returns the created resource.
// Only the keys present in p are sent, so the server applies its defaults to
// every other attribute.
func (s *crudService[T]) createFields(ctx context.Context, path string, p Patch) (*T, error) {
	res, err := p.resource(jsonapiTypeOf[T](), "")
	if err != nil {
		return nil, err
	}
	req, err := s.client.requestDocument(ctx, http.MethodPost, path, res)
	if err != nil {
		return nil, err
	}

	created := new(T)
	if _, err := s.client.do(req, created); err != nil {
		return nil, err
	}
	return created, nil
}
