package github

import (
	"context"
	"testing"
)

func TestPickClient_Pick(t *testing.T) {
	ctx := context.Background()
	SHA := "9f320ef77b4a8818bf48a7c28a93f9e2faeecdc7"
	Branch := "release/23.04"
	token := ""
	client := NewClient(context.Background(), token)
	err := client.Pick(ctx, "",
		&PickOption{SHA: SHA, Branch: Branch})
	t.Logf("err: %v", err)
}
