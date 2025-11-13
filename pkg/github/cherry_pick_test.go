package github

import (
	"context"
	tp "github.com/chasenio/norn/pkg/types"
	"testing"
)

func TestPickClient_Pick(t *testing.T) {
	ctx := context.Background()
	SHA := ""
	Branch := "master"
	token := ""

	client := NewGithubClient(ctx, token)
	pickServuce := NewCherryPickService(client)
	err := pickServuce.CherryPick(ctx, "chasenio/pick",
		&tp.Option{SHA: SHA, Branch: Branch, Prefix: "cherry-kit"})
	if err != nil {
		t.Errorf("err: %v", err)
	} else {

		t.Logf("err: %v", err)
	}
}
