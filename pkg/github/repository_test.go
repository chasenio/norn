package github

import (
	"context"
	tp "github.com/kentio/norn/pkg/types"
	"testing"
)

func TestRepositoryService_Get(t *testing.T) {
	ctx := context.Background()
	token := ""
	client := NewGithubClient(ctx, token)

	repoClient := NewRepositoryService(client)

	repo, err := repoClient.Get(ctx, &tp.GetRepositoryOption{
		Repo: "",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
		return
	}
	t.Logf("repo: %+v", repo)
}
