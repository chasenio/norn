package feature

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/kentio/norn/types"
	"github.com/sirupsen/logrus"
)

type PickOption struct {
	SHA    string
	Repo   string
	Target string
}

func Pick(ctx context.Context, provider types.Provider, opt *PickOption) error {
	if provider == nil || opt == nil {
		return ErrInvalidOptions
	}

	// 1. get reference
	reference, err := provider.Reference().Get(ctx, &types.GetRefOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
	})
	if err != nil {
		return err
	}

	// 2. get commit
	commit, err := provider.Commit().Get(ctx, &types.GetCommitOption{
		Repo: opt.Repo,
		SHA:  opt.SHA,
	})
	if err != nil {
		return err
	}

	// 2.1 if last commit message is same as pick message, skip
	lastCommit, err := provider.Commit().Get(ctx, &types.GetCommitOption{Repo: opt.Repo, SHA: reference.SHA})
	if err != nil {
		logrus.Debugf("failed to get last commit: %+v", err)
		return err
	}
	pickMessage := fmt.Sprintf("Cherry-pick from %s\nSource Commit Message:\n%s", opt.SHA, commit.Message())

	// if match message, skip
	lastCommitMessageMd5 := sumMd5(lastCommit.Message())
	pickMessageMd5 := sumMd5(pickMessage)
	logrus.Debugf("last commit message md5: %s, pick message md5: %s", lastCommitMessageMd5, pickMessageMd5)
	if lastCommitMessageMd5 == pickMessageMd5 {
		logrus.Debugf("reference %s last commit message is same as pick message, skip", reference.SHA)
		return nil
	}
	// 3. create commit
	createCommit, err := provider.Commit().Create(ctx, &types.CreateCommitOption{
		Repo:        opt.Repo,
		Tree:        commit.Tree(),
		SHA:         commit.SHA(),
		PickMessage: pickMessage,
		Parents: []string{
			reference.SHA,
			commit.SHA(),
		}})

	if err != nil {
		return err
	}

	// 4. update reference
	_, err = provider.Reference().Update(ctx, &types.UpdateOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
		SHA:  createCommit.SHA(),
	})

	if err != nil {
		return err
	}

	return nil
}

func sumMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
