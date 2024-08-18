package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v62/github"
	"github.com/kentio/norn/pkg/types"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type PickService struct {
	client *gh.Client
}

type RepoOption struct {
	Owner string
	Repo  string
}

func NewPickService(client *gh.Client) *PickService {
	return &PickService{
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

func (c *PickService) Pick(ctx context.Context, repo string, opt *tp.PickOption) error {
	repoOpt, err := parseRepo(repo)
	if err != nil {
		return err
	}
	if repoOpt == nil || opt == nil {
		return types.ErrInvalidOptions
	}
	// get target ref details
	targetRef, _, err := c.client.Git.GetRef(ctx, repoOpt.Owner, repoOpt.Repo, "refs/heads/"+opt.Branch)
	if err != nil {
		return tp.NotFound
	}

	// get target latest commit details
	latestCommit, _, err := c.client.Git.GetCommit(ctx, repoOpt.Owner, repoOpt.Repo, targetRef.Object.GetSHA())
	if err != nil {
		logrus.Errorf("1 get target commit")
		return err
	}

	// 需要 cherry-pick 的 commit
	sourceCommit, _, err := c.client.Git.GetCommit(ctx, repoOpt.Owner, repoOpt.Repo, opt.SHA)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return err
	}

	// create a temporary ref
	tempRef := fmt.Sprintf("refs/heads/pick-%s-%s", opt.Branch, opt.SHA)
	// Delete the temporary ref
	defer func() {
		_, err := c.client.Git.DeleteRef(ctx, repoOpt.Owner, repoOpt.Repo, tempRef)
		if err != nil {
			logrus.Errorf("Failed to delete temporary ref %s: %v", tempRef, err)
		}
	}()
	_, _, err = c.client.Git.CreateRef(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Reference{
		Ref: gh.String(tempRef),
		Object: &gh.GitObject{
			SHA: gh.String(targetRef.Object.GetSHA()),
		},
	})

	if err != nil {
		logrus.Errorf("Failed to create temporary ref %s: %v", tempRef, err)
		return err
	}

	// 创建一个新的 sibling commit
	siblingCommit, _, err := c.client.Git.CreateCommit(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Commit{
		Author:    sourceCommit.Author,
		Committer: sourceCommit.Committer,
		Message:   gh.String(fmt.Sprintf("Sibling of %s", sourceCommit.GetSHA())),
		Tree:      &gh.Tree{SHA: latestCommit.Tree.SHA},
		Parents:   []*gh.Commit{{SHA: sourceCommit.Parents[0].SHA}},
	}, nil)

	if err != nil {
		logrus.Errorf("Failed to create new commit: %v with temp ref", err)
		return err
	}

	// update temp ref to sibling commit
	_, _, err = c.client.Git.UpdateRef(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Reference{
		Ref: gh.String(tempRef),
		Object: &gh.GitObject{
			SHA: siblingCommit.SHA,
		},
	}, true)

	if err != nil {
		logrus.Errorf("Failed to update temp ref %s", tempRef)
		return err
	}
	// merge pick commit to temp branch
	base := strings.Split(tempRef, "/")[2]
	logrus.Warnf("temp ref: %s base: %s", tempRef, base)
	// merge pick commit to temp branch
	mergeSha, err := c.Merge(ctx, &MergeOption{
		Owner: repoOpt.Owner,
		Repo:  repoOpt.Repo,
		Base:  tempRef,
		SHA:   opt.SHA,
	})
	if err != nil {
		return err
	}

	// create the final pick commit
	message := fmt.Sprintf("%s\n\n(cherry picked from commit %s)", *sourceCommit.Message, sourceCommit.GetSHA()[:7])
	newCommit, _, err := c.client.Git.CreateCommit(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Commit{
		Author:    sourceCommit.Author,
		Committer: sourceCommit.Committer,
		Message:   gh.String(message),
		Tree:      &gh.Tree{SHA: mergeSha, Truncated: gh.Bool(false)},
		Parents:   []*gh.Commit{{SHA: latestCommit.SHA}},
	}, nil)

	if err != nil {
		logrus.Errorf("creating commit with different tree")
		return err
	}

	// update the ref to the new commit with temp ref
	_, _, err = c.client.Git.UpdateRef(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Reference{
		Ref: gh.String(tempRef),
		Object: &gh.GitObject{
			SHA: newCommit.SHA,
		},
	}, true)
	if err != nil {
		logrus.Error("update temp branch error")
		return err
	}

	// update target branch
	_, _, err = c.client.Git.UpdateRef(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Reference{
		Ref: targetRef.Ref,
		Object: &gh.GitObject{
			SHA: newCommit.SHA,
		},
	}, true)
	if err != nil {
		logrus.Errorf("update target branch error %s", *targetRef.Ref)
		return err
	}

	return nil
}

type MergeOption struct {
	Owner string
	Repo  string
	Base  string
	SHA   string
}

func (c *PickService) Merge(ctx context.Context, opt *MergeOption) (*string, error) {
	logrus.Infof("merge new commit %s to temp branch %s", opt.SHA, opt.Base)
	mergeCommit, mergeResp, err := c.client.Repositories.Merge(ctx, opt.Owner, opt.Repo, &gh.RepositoryMergeRequest{
		Base:          gh.String(opt.Base),
		Head:          gh.String(opt.SHA),
		CommitMessage: gh.String(fmt.Sprintf("Merge %s into %s", opt.SHA, opt.Base)),
	})
	// merge conflict
	if mergeResp.StatusCode == http.StatusConflict && strings.Contains(err.Error(), "conflict") {
		logrus.Warnf("merge sha %s to %s conflict: %s in %s/%s", opt.SHA, opt.Base, mergeResp.Status, opt.Owner, opt.Repo)
		return nil, types.ErrConflict
	}
	if err != nil {
		logrus.Errorf("merge new commit to temp branch error %s", opt.Base)
		return nil, err
	}
	return mergeCommit.Commit.Tree.SHA, nil
}
