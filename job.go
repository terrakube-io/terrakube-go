package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Job represents a Terrakube job resource.
type Job struct {
	ID      string `jsonapi:"primary,job"`
	Command string `jsonapi:"attr,command"`
	Output  string `jsonapi:"attr,output"`
	Status  string `jsonapi:"attr,status"`
}

// JobService handles communication with the job related methods of the
// Terrakube API.
type JobService struct {
	client *Client
}

// List returns all jobs for the given organization.
func (s *JobService) List(ctx context.Context, orgID string, opts *ListOptions) ([]*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "job")

	var params url.Values
	if opts != nil && opts.Filter != "" {
		params = url.Values{"filter[job]": {opts.Filter}}
	}

	req, err := s.client.requestWithQuery(ctx, http.MethodGet, p, params, nil)
	if err != nil {
		return nil, err
	}

	var jobs []*Job
	_, err = s.client.do(ctx, req, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// Get returns a single job by ID.
func (s *JobService) Get(ctx context.Context, orgID, id string) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", id); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "job", id)

	req, err := s.client.request(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	job := &Job{}
	_, err = s.client.do(ctx, req, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// Create creates a new job in the given organization.
func (s *JobService) Create(ctx context.Context, orgID string, job *Job) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "job")

	req, err := s.client.request(ctx, http.MethodPost, p, job)
	if err != nil {
		return nil, err
	}

	created := &Job{}
	_, err = s.client.do(ctx, req, created)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Update modifies an existing job. The job's ID field must be set.
func (s *JobService) Update(ctx context.Context, orgID string, job *Job) (*Job, error) {
	if err := validateID("organizationID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("jobID", job.ID); err != nil {
		return nil, err
	}

	p := s.client.apiPath("organization", orgID, "job", job.ID)

	req, err := s.client.request(ctx, http.MethodPatch, p, job)
	if err != nil {
		return nil, err
	}

	updated := &Job{}
	_, err = s.client.do(ctx, req, updated)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete removes a job by ID.
func (s *JobService) Delete(ctx context.Context, orgID, id string) error {
	if err := validateID("organizationID", orgID); err != nil {
		return err
	}
	if err := validateID("jobID", id); err != nil {
		return err
	}

	p := s.client.apiPath("organization", orgID, "job", id)

	req, err := s.client.request(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	return err
}
