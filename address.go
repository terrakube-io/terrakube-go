package terrakube

import "context"

// Address represents a Terrakube job address resource.
type Address struct {
	ID          string  `jsonapi:"primary,address"`
	Name        string  `jsonapi:"attr,name"`
	Type        string  `jsonapi:"attr,type"`
	CreatedBy   *string `jsonapi:"attr,createdBy"`
	CreatedDate *string `jsonapi:"attr,createdDate"`
	UpdatedBy   *string `jsonapi:"attr,updatedBy"`
	UpdatedDate *string `jsonapi:"attr,updatedDate"`
	Job         *Job    `jsonapi:"relation,job,omitempty"`
}

// AddressService handles communication with the job address endpoints.
type AddressService struct {
	crudService[Address]
}

// List returns all addresses for a job.
func (s *AddressService) List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*Address, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "address")
	return s.list(ctx, path, opts)
}

// Get returns a single address by ID.
func (s *AddressService) Get(ctx context.Context, orgID, jobID, id string) (*Address, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("address ID", id); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "address", id)
	return s.get(ctx, path)
}

// Create creates a new address for a job.
func (s *AddressService) Create(ctx context.Context, orgID, jobID string, address *Address) (*Address, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "address")
	return s.create(ctx, path, address)
}

// Update modifies an existing address. The address's ID field must be set.
func (s *AddressService) Update(ctx context.Context, orgID, jobID string, address *Address) (*Address, error) {
	if err := validateID("organization ID", orgID); err != nil {
		return nil, err
	}
	if err := validateID("job ID", jobID); err != nil {
		return nil, err
	}
	if err := validateID("address ID", address.ID); err != nil {
		return nil, err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "address", address.ID)
	return s.update(ctx, path, address)
}

// Delete removes an address by ID.
func (s *AddressService) Delete(ctx context.Context, orgID, jobID, id string) error {
	if err := validateID("organization ID", orgID); err != nil {
		return err
	}
	if err := validateID("job ID", jobID); err != nil {
		return err
	}
	if err := validateID("address ID", id); err != nil {
		return err
	}

	path := s.client.apiPath("organization", orgID, "job", jobID, "address", id)
	return s.del(ctx, path)
}
