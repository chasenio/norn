package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
)

type ReferenceService struct {
	client *github.Client
}

type GetRefOption struct {
	Repo string
	Ref  string
}

func NewReferenceService(client *github.Client) *ReferenceService {
	return &ReferenceService{
		client: client,
	}
}

// Get reference
func (s *ReferenceService) Get(ctx context.Context, opt *GetRefOption) (*types.BranchRef, error) {
	if opt == nil {
		return nil, ErrInvalidOptions
	}

	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}

	branchRef, _, err := s.client.Git.GetRef(ctx, repoOpt.Owner, repoOpt.Repo, opt.Ref)
	if err != nil {
		return nil, fmt.Errorf("get reference failed: %v", err)
	}
	return newBranch(branchRef), nil
}

// Find returns the branch with the specified filters.
func (s *ReferenceService) Find(ctx context.Context, opts *types.FindOptions) ([]types.BranchRef, error) {
	if opts == nil {
		return nil, ErrInvalidOptions
	}
	return nil, nil
}

func (s *ReferenceService) Create(ctx context.Context, opts *types.CreateOptions) (*types.BranchRef, error) {
	return nil, nil
}

func (s *ReferenceService) Update(ctx context.Context, opts *types.UpdateOptions) (*types.BranchRef, error) {
	return nil, nil
}

func (s *ReferenceService) Delete(ctx context.Context, opts *types.DeleteOptions) error {
	return nil
}

func newBranch(branchRef *github.Reference) *types.BranchRef {
	return &types.BranchRef{
		Ref: branchRef.Ref,
		SHA: branchRef.Object.SHA,
	}
}
