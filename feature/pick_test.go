package feature

import (
	"context"
	"github.com/kentio/norn/github"
	"testing"
)

func TestPick(t *testing.T) {
	ctx := context.Background()
	provider, _ := github.NewProvider(ctx, "")

	err := Pick(ctx, provider, &PickOption{
		SHA:    "696b3168704d0d5b811d80615b3e1a6a31b2d2a5",
		Repo:   "kentio/test_cherry_pick",
		Target: "master"})

	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}
