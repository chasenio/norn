package github

import (
	"context"
	gh "github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

func NewGithubClient(ctx context.Context, token string) *gh.Client {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return gh.NewClient(tc)
}
