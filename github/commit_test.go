package github

import (
	"context"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"testing"
)

func TestCommitService_Find(t *testing.T) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "your access token"})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	_ = NewCommitService(client)

}
