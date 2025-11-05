package github

import (
	"context"
	tp "github.com/chasenio/norn/pkg/types"
	gh "github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

func NewGithubClient(ctx context.Context, token string) *gh.Client {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return gh.NewClient(tc)
}

func NewGitHubWithBaseUrl(ctx context.Context, opt *tp.CreateProviderOption) *gh.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opt.Token})
	tc := oauth2.NewClient(ctx, ts)
	client, _ := gh.NewClient(tc).WithEnterpriseURLs(*opt.BaseUrl, *opt.UploadUrl)

	return client
}
