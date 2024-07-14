package types

import "context"

type Repository interface {
	Name() string
	FullName() string
	GitUrl() string
	DefaultBranch() string
	AllowSquashMerge() *bool
	DeleteBranchOnMerge() *bool
	AllowRebaseMerge() *bool
	Private() *bool
}

type GetRepositoryOption struct {
	Repo string
}

type RepositoryService interface {
	Get(ctx context.Context, opt *GetRepositoryOption) (Repository, error)
}
