package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	"github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ReferenceService struct {
	client *gh.Client
}

func NewReferenceService(client *gh.Client) *ReferenceService {
	return &ReferenceService{
		client: client,
	}
}

// Get reference
func (s *ReferenceService) Get(ctx context.Context, opt *types.GetRefOption) (*types.Reference, error) {
	if opt == nil {
		return nil, types.ErrInvalidOptions
	}

	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Get Reference Opt: %+v", opt)

	branchRef, response, err := s.client.Git.GetRef(ctx, repoOpt.Owner, repoOpt.Repo, opt.Ref)
	if err != nil {
		logrus.Debugf("Get Reference Response: %+v", response)
		return nil, fmt.Errorf("get reference failed: %v", err)
	}
	logrus.Debugf("Get Reference: %+v", *branchRef)
	return newBranch(branchRef), nil
}

// Find returns the branch with the specified filters.
func (s *ReferenceService) Find(ctx context.Context, opts *types.FindOptions) ([]types.Reference, error) {
	if opts == nil {
		return nil, types.ErrInvalidOptions
	}
	return nil, nil
}

func (s *ReferenceService) Create(ctx context.Context, opt *types.CreateOptions) (*types.Reference, error) {
	return nil, nil
}

// Update updates the reference with the specified options.
func (s *ReferenceService) Update(ctx context.Context, opt *types.UpdateOption) (*types.Reference, error) {
	if opt == nil {
		return nil, types.ErrInvalidOptions
	}
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Update Reference Opt: %+v", opt)
	ref, response, err := s.client.Git.UpdateRef(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Reference{
		Ref: gh.String(opt.Ref),
		Object: &gh.GitObject{
			SHA: gh.String(opt.SHA),
		},
	}, false)
	logrus.Debugf("Update Reference Response: %+v", response)
	if err != nil {
		return nil, fmt.Errorf("update reference failed: %v", err)
	}
	if response.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("reference: %v", *ref.Ref)
	}
	logrus.Debugf("Update Reference: %+v", *ref)
	return &types.Reference{Ref: *ref.Ref}, nil
}

func (s *ReferenceService) Delete(ctx context.Context, opt *types.DeleteOptions) error {
	return nil
}

func newBranch(branchRef *gh.Reference) *types.Reference {
	return &types.Reference{
		Ref: *branchRef.Ref,
		SHA: *branchRef.Object.SHA,
	}
}
