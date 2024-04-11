package github

import (
	"context"
	gh "github.com/google/go-github/v50/github"
	types2 "github.com/kentio/norn/types"
)

type Provider struct {
	ProviderID string
	client     *gh.Client

	commitService       *CommitService
	referenceService    *ReferenceService
	mergeRequestService *PullRequestService
	commentService      *CommentService
}

type Credential struct {
	Token          string
	AppID          int64
	InstallationID int64
	PrivateKey     []byte
}

func NewProvider(ctx context.Context, token string) *Provider {
	client := NewGithubClient(ctx, token)
	return &Provider{
		ProviderID:          "github",
		commitService:       NewCommitService(client),
		referenceService:    NewReferenceService(client),
		mergeRequestService: NewPullRequestService(client),
		commentService:      NewCommentService(client),
	}
}

// NewProviderWithOpt creates a new provider with the given options.
func NewProviderWithOpt(ctx context.Context, cred *Credential) (*Provider, error) {
	if cred.Token != "" {
		return NewProvider(ctx, cred.Token), nil
	}
	client, err := NewGithubClientWithPrivateKey(cred)
	if err != nil {
		return nil, err
	}
	return &Provider{
		ProviderID:          "github",
		commitService:       NewCommitService(client),
		referenceService:    NewReferenceService(client),
		mergeRequestService: NewPullRequestService(client),
		commentService:      NewCommentService(client),
	}, nil
}

func (p *Provider) Commit() types2.CommitService {
	return p.commitService
}

func (p *Provider) Reference() types2.ReferenceService {
	return p.referenceService
}

func (p *Provider) MergeRequest() types2.MergeRequestService {
	return p.mergeRequestService
}

func (p *Provider) Comment() types2.CommentService {
	return p.commentService
}
