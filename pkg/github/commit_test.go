package github

import (
	"context"
	"github.com/google/go-github/v60/github"
	"github.com/kentio/norn/pkg/types"
	"golang.org/x/oauth2"
	"testing"
)

func TestCommitService_Find(t *testing.T) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "your access token"})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	_ = NewCommitService(client)

}

func TestCommitService_Get(t *testing.T) {
	ctx := context.Background()
	token := ""
	client := NewGithubClient(ctx, token)
	commitClient := NewCommitService(client)

	commit, err := commitClient.Get(ctx, &types.GetCommitOption{
		Repo: "",
		SHA:  "",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("commit: %+v", commit)
}
