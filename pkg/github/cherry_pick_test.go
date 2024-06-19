package github

import (
	"context"
	tp "github.com/kentio/norn/pkg/types"
	"testing"
)

func TestPickClient_Pick(t *testing.T) {
	ctx := context.Background()
	SHA := ""
	Branch := ""
	token := ""
	client := NewGithubClient(ctx, token)
	pickServuce := NewPickService(client)
	err := pickServuce.Pick(ctx, "",
		&tp.PickOption{SHA: SHA, Branch: Branch})
	if err != nil {
		t.Errorf("err: %v", err)
	} else {

		t.Logf("err: %v", err)
	}
}
