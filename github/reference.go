package github

import (
	"context"
	"github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
)

type ReferenceService struct {
	client *github.Client
}

func NewReferenceService(client *github.Client) *ReferenceService {
	return &ReferenceService{
		client: client,
	}
}

// Find returns the branch with the specified filters.
func (s *ReferenceService) Find(ctx context.Context, opts *types.FindOptions) ([]types.Branch, error) {
	if opts == nil {
		return nil, ErrInvalidOptions
	}
	return nil, nil
}

func (s *ReferenceService) Create(ctx context.Context, opts *types.CreateOptions) (*types.Branch, error) {
	return nil, nil

}

func (s *ReferenceService) Update(ctx context.Context, opts *types.UpdateOptions) (*types.Branch, error) {
	return nil, nil
}

func (s *ReferenceService) Delete(ctx context.Context, opts *types.DeleteOptions) error {
	return nil
}

func newBranch(branch *github.Branch) *types.Branch {
	return &types.Branch{}
}
