package github

import (
	"context"
	"testing"
)

func TestPickClient_Pick(t *testing.T) {
	ctx := context.Background()
	SHA := "696b3168704d0d5b811d80615b3e1a6a31b2d2a5"
	Branch := "release/23.04"
	token := ""
	client := NewClient(context.Background(), token)
	err := client.Pick(ctx, "kentio/test_cherry_pick",
		&PickOption{SHA: SHA, Branch: Branch})
	t.Logf("err: %v", err)
}
