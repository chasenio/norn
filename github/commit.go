package github

import (
	"context"
	"github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
	"github.com/sirupsen/logrus"
)

type CommitService struct {
	client *github.Client
}

func NewCommitService(client *github.Client) *CommitService {
	return &CommitService{
		client: client,
	}
}

// Find returns the branch with the specified filters.
func (s *CommitService) Get(ctx context.Context, opt *types.GetCommitOption) (*types.CommitInfo, error) {
	if opt == nil {
		return nil, ErrInvalidOptions
	}
	commmits, response, err := s.client.Repositories.GetCommit(ctx, opt.Owner, opt.Repo, opt.Sha, nil)
	if err != nil {
		return nil, err
	}
	logrus.Infof("response: %v", response)
	return newCommit(commmits), nil
}

func newCommit(commit *github.RepositoryCommit) *types.CommitInfo {
	return &types.CommitInfo{
		SHA: commit.SHA,
		Author: &types.UserSpec{
			Name: commit.Author.Name,
		},
		Committer: &types.UserSpec{
			Name: commit.Committer.Name,
		},
		HTMLURL: commit.HTMLURL,
	}
}
