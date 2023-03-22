package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
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

func parseRepo(repo string) (*RepoOption, error) {
	// parse "kentio/norn" 字符串分割
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo: %s", repo)
	}
	return &RepoOption{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

func (c *PickClient) Pick(ctx context.Context, repo string, opt *PickOption) error {
	repoOpt, err := parseRepo(repo)
	if err != nil {
		return err
	}
	if repoOpt == nil || opt == nil {
		return ErrInvalidOptions
	}

	branchRef, _, err := c.client.Git.GetRef(ctx, repoOpt.Owner, repoOpt.Repo, fmt.Sprintf("refs/heads/%s", opt.Branch))
	if err != nil {
		return fmt.Errorf("failed to get branch ref: %v", err)
	}

	// 获取源提交(commitSHA)的提交对象
	commit, _, err := c.client.Repositories.GetCommit(ctx, repoOpt.Owner, repoOpt.Repo, opt.SHA, nil)
	if err != nil {
		return fmt.Errorf("failed to get commit: %v", err)
	}

	// 创建新的提交对象
	cherryPickMessage := fmt.Sprintf("Cherry-pick from %s\nSource Commit Message:\n%s", *commit.SHA, *commit.Commit.Message)
	createCommit, _, err := c.client.Git.CreateCommit(ctx, repoOpt.Owner, repoOpt.Repo, &github.Commit{
		Message: github.String(cherryPickMessage),
		Tree:    commit.Commit.Tree,
		Parents: []*github.Commit{{SHA: branchRef.Object.SHA}, {SHA: commit.SHA}},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	// Update target branch reference
	reference, response, err := c.client.Git.UpdateRef(ctx, repoOpt.Owner, repoOpt.Repo, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", opt.Branch)),
		Object: &github.GitObject{
			SHA: createCommit.SHA,
		},
	},
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to update ref: %v", err)
	}
	if response.StatusCode == http.StatusUnprocessableEntity {
		return fmt.Errorf("reference: %v", *reference.Ref)
	}
	logrus.Infof("success to pick commit: %s to ref: %s", *createCommit.SHA, *reference.Ref)
	return nil
}
