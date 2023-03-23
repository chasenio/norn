package feature

import (
	"context"
	"fmt"
	"github.com/kentio/norn/types"
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

	// 3. create commit
	pickMessage := fmt.Sprintf("Cherry-pick from %s\nSource Commit Message:\n%s", opt.SHA, commit.Message())
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
