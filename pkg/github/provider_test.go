package github

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"github.com/kentio/norn/pkg/types"
	"net/http"
	"testing"
)

func TestNewProvider(t *testing.T) {
	token := ""
	provider := NewProvider(nil, token)

	t.Logf("provider: %v", provider)
}

func TestGithubAppClient(t *testing.T) {
	keypath := " "
	var appID int64 = 1
	var installationID int64 = 1
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appID, installationID, keypath)
	if err != nil {
		t.Errorf("NewKeyFromFile error: %v", err)
	}

	client := github.NewClient(&http.Client{Transport: itr})
	t.Logf("client: %+v", client)

	ctx := context.Background()

	srv := NewCommentService(client)

	comment, err := srv.Create(ctx, &types.CreateCommentOption{
		Repo:           " ",
		MergeRequestID: " ",
		Body:           "test comment",
	})
	if err != nil {
		t.Errorf("Create error: %v", err)
	}
	t.Logf("comment: %+v", comment)
}
