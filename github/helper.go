package github

import (
	"context"
	"errors"
	gh "github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

var ErrInvalidOptions = errors.New("invalid parameter, please check your request")

var NotFound = errors.New("not found")

func NewGithubClient(ctx context.Context, token string) *gh.Client {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return gh.NewClient(tc)
}
