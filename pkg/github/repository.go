package github

import (
	"context"
	gh "github.com/google/go-github/v62/github"
	tp "github.com/kentio/norn/pkg/types"
)

type Repository struct {
	name                string
	fullName            string
	gitUrl              string
	defaultBranch       string
	allowSquashMerge    *bool
	deleteBranchOnMerge *bool
	allowRebaseMerge    *bool
	private             *bool
}

type RepositoryService struct {
	client *gh.Client
}

func NewRepositoryService(client *gh.Client) *RepositoryService {
	return &RepositoryService{
		client: client,
	}
}

func (r RepositoryService) Get(ctx context.Context, opt *tp.GetRepositoryOption) (tp.Repository, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}

	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}

	repo, _, err := r.client.Repositories.Get(ctx, repoOpt.Owner, repoOpt.Repo)
	if err != nil {
		return nil, err
	}
	return newRepository(repo), nil

}

func newRepository(repo *gh.Repository) *Repository {
	return &Repository{
		name:                *repo.Name,
		fullName:            *repo.FullName,
		gitUrl:              *repo.GitURL,
		defaultBranch:       *repo.DefaultBranch,
		allowSquashMerge:    repo.AllowSquashMerge,
		deleteBranchOnMerge: repo.DeleteBranchOnMerge,
		allowRebaseMerge:    repo.AllowRebaseMerge,
		private:             repo.Private,
	}
}

func (r *Repository) Name() string {
	return r.name
}

func (r *Repository) FullName() string {
	return r.fullName
}

func (r *Repository) GitUrl() string {
	return r.gitUrl
}

func (r *Repository) DefaultBranch() string {
	return r.defaultBranch
}

func (r *Repository) AllowSquashMerge() *bool {
	return r.allowSquashMerge
}

func (r *Repository) DeleteBranchOnMerge() *bool {
	return r.deleteBranchOnMerge
}

func (r *Repository) AllowRebaseMerge() *bool {
	return r.allowRebaseMerge
}

func (r *Repository) Private() *bool {
	return r.private
}
