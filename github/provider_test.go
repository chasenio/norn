package github

import (
	"context"
	"github.com/kentio/norn/types"
	"testing"
)

func TestNewProvider(t *testing.T) {
	token := ""
	provider := NewProvider(nil, token)

	t.Logf("provider: %v", provider)
}

func TestGithubAppClient(t *testing.T) {

	opt := &Credential{
		AppID:          0,
		InstallationID: 0,
		PrivateKey: []byte(
			``),
	}

	gh, err := NewProviderWithOpt(context.Background(), opt)
	if err != nil {
		t.Errorf("New Client error: %v", err)
	}
	t.Logf("client: %+v", gh)

	ctx := context.Background()

	comment, err := gh.Comment().Create(ctx, &types.CreateCommentOption{
		Repo:           "",
		MergeRequestID: "",
		Body:           "test comment",
	})
	if err != nil {
		t.Errorf("Create error: %v", err)
	}
	t.Logf("comment: %+v", comment)
}
