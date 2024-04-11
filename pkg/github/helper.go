package github

import (
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"net/http"
)

func NewGithubClient(ctx context.Context, token string) *gh.Client {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return gh.NewClient(tc)
}

func NewGithubClientWithPrivateKey(cred *Credential) (*gh.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, cred.AppID, cred.InstallationID, cred.PrivateKey)
	if err != nil {
		return nil, err
	}
	client := gh.NewClient(&http.Client{Transport: itr})
	return client, nil
}
