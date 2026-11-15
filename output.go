package terrakube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Output is a Terraform output of a workspace.
type Output struct {
	Name      string `json:"name"`
	Value     any    `json:"value"`
	Type      any    `json:"type"`
	Sensitive bool   `json:"sensitive"`
}

// OutputService reads workspace outputs from the workspace state.
type OutputService struct {
	client *Client
}

// stateOutput is an output as Terraform writes it in a state file.
type stateOutput struct {
	Value     any  `json:"value"`
	Type      any  `json:"type"`
	Sensitive bool `json:"sensitive"`
}

// List returns the outputs of a workspace, sorted by name. It reads
// values.outputs from the JSON state of the newest history entry (by
// createdDate) that names a state file. If there is no such entry, its JSON
// state has no outputs, or that state is stored outside the API (see
// [StateService.Download]: nothing is sent there), it reads outputs from the
// current raw Terraform state instead. A workspace without history has no
// outputs. Sensitive values are returned in clear text with Sensitive set.
// Numbers in values are json.Number, so large integers keep every digit.
//
// It returns a *ValidationError if orgID or wsID is empty and a *APIError on
// server errors.
func (s *OutputService) List(ctx context.Context, orgID, wsID string) ([]*Output, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("workspaceID", wsID); err != nil {
		return nil, err
	}

	history, err := s.client.History.List(ctx, orgID, wsID, &ListOptions{Sort: "-createdDate"})
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return []*Output{}, nil
	}

	// The server saves some entries (rollback, remote state, import) before
	// their state file, with an empty Output: those are skipped.
	if newest := newestHistory(history); newest != nil {
		var showJSON struct {
			Values struct {
				Outputs map[string]stateOutput `json:"outputs"`
			} `json:"values"`
		}
		body, err := s.client.State.Download(ctx, newest, StateJSON)
		if errors.Is(err, ErrForeignHost) {
			err = ErrStateNotFound // fall back to the raw state, read through the API
		}
		if err := decodeState(body, err, &showJSON); err != nil {
			return nil, err
		}
		if len(showJSON.Values.Outputs) > 0 {
			return sortedOutputs(showJSON.Values.Outputs), nil
		}
	}

	var raw struct {
		Outputs map[string]stateOutput `json:"outputs"`
	}
	body, err := s.client.State.Current(ctx, orgID, wsID)
	if err := decodeState(body, err, &raw); err != nil {
		return nil, err
	}
	return sortedOutputs(raw.Outputs), nil
}

// decodeState decodes the state in body into v and closes body. err is the
// error from opening body; a missing state (ErrStateNotFound) decodes as nothing.
func decodeState(body io.ReadCloser, err error, v any) error {
	if errors.Is(err, ErrStateNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	defer body.Close() //nolint:errcheck // response body close errors are inconsequential
	dec := json.NewDecoder(body)
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decoding state: %w", err)
	}
	return nil
}

// newestHistory returns the entry with the latest createdDate among those with
// an Output, or nil for none. Entries without a date sort before every dated
// entry.
func newestHistory(history []*History) *History {
	var newest *History
	for _, h := range history {
		switch {
		case h.Output == "":
		case newest == nil:
			newest = h
		case h.CreatedDate == nil:
		case newest.CreatedDate == nil || h.CreatedDate.After(*newest.CreatedDate):
			newest = h
		}
	}
	return newest
}

// sortedOutputs converts state outputs to a slice sorted by name.
func sortedOutputs(m map[string]stateOutput) []*Output {
	out := make([]*Output, 0, len(m))
	for name, o := range m {
		out = append(out, &Output{Name: name, Value: o.Value, Type: o.Type, Sensitive: o.Sensitive})
	}
	slices.SortFunc(out, func(a, b *Output) int { return strings.Compare(a.Name, b.Name) })
	return out
}
