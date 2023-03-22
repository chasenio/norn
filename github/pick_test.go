package github

import (
	"context"
	"testing"
)

func TestPickClient_Pick(t *testing.T) {
	ctx := context.Background()
	SHA := "ad3719a041af1a374eb88262df79aa784b2d0fc0"
	Branch := "release/23.04"
	token := ""
	client := NewClient(context.Background(), token)
	err := client.Pick(ctx, &RepoOption{Owner: "kentio", Repo: "test_cherry_pick"},
		&PickOption{SHA: SHA, Branch: Branch})
	t.Logf("err: %v", err)
}
