package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
)

type PickClient struct {
	client *github.Client
}

type PickOption struct {
	SHA    string
	Branch string
}

type RepoOption struct {
	Owner string
	Repo  string
}

func NewClient(ctx context.Context, token string) *PickClient {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &PickClient{
		client: client,
	}
}

func (c *PickClient) Pick(ctx context.Context, repo *RepoOption, opt *PickOption) error {

	if repo == nil || opt == nil {
		return ErrInvalidOptions
	}

	commit, response, err := c.client.Repositories.GetCommit(ctx, repo.Owner, repo.Repo, opt.SHA, nil)
	if err != nil {
		logrus.Infof("response: %v", response)
		return err
	}
	// Get Branch
	_, response, err = c.client.Repositories.GetBranch(ctx, repo.Owner, repo.Repo, opt.Branch, false)
	if err != nil {
		logrus.Infof("response: %v", response)
		return err
	}

	// get tree

	cherryPickMessage := fmt.Sprintf("Cherry-pick from %s Source Message: %s", *commit.SHA, *commit.Commit.Message)
	createCommit, response, err := c.client.Git.CreateCommit(ctx, repo.Owner, repo.Repo, &github.Commit{
		//SHA:       commit.SHA,
		Author:    commit.Commit.Author,
		Committer: commit.Commit.Committer,
		Message:   github.String(cherryPickMessage),
		Tree:      commit.Commit.Tree,
		//Parents:      branch.Commit.Parents,
		Stats:        commit.Stats,
		URL:          commit.Commit.URL,
		Verification: commit.Commit.Verification,
		NodeID:       commit.NodeID,
	})
	if err != nil {
		logrus.Infof("response: %v", response)
		return err
	}
	// Update branch
	reference, response, err := c.client.Git.UpdateRef(ctx, repo.Owner, repo.Repo, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", opt.Branch)),
		Object: &github.GitObject{
			SHA: createCommit.SHA,
		},
	},
		false,
	)
	if err != nil {
		logrus.Infof("response: %v", response)
		return err
	}
	if response.StatusCode == http.StatusUnprocessableEntity {
		return fmt.Errorf("reference: %v", reference)
	}
	logrus.Infof("reference: %v", reference)
	return nil
}

// 实现一个github的pick功能
