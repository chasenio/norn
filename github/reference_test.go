package github

import (
	"context"
	"github.com/kentio/norn/types"
	"testing"
)

func TestReferenceService_Get(t *testing.T) {
	token := ""
	ctx := context.Background()
	client := NewGithubClient(ctx, token)
	referenceClient := NewReferenceService(client)

	reference, err := referenceClient.Get(ctx, &types.GetRefOption{
		Repo: "kentio/norn",
		Ref:  "heads/topic/jeff/add_def",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("reference: %+v", reference)
}
