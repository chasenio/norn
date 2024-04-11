package github

import (
	"context"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/pkg/types"
)

type Provider struct {
	ProviderID string
	client     *gh.Client

	commitService       *CommitService
	referenceService    *ReferenceService
	mergeRequestService *PullRequestService
	commentService      *CommentService
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

// NewProviderWithClient creates a new provider with the given client.
func NewProviderWithClient(client *gh.Client) *Provider {
	return &Provider{
		ProviderID:          "github",
		client:              client,
		commitService:       NewCommitService(client),
		referenceService:    NewReferenceService(client),
		mergeRequestService: NewPullRequestService(client),
		commentService:      NewCommentService(client),
	}
}

func (p *Provider) Commit() tp.CommitService {
	return p.commitService
}

func (p *Provider) Reference() tp.ReferenceService {
	return p.referenceService
}

func (p *Provider) MergeRequest() tp.MergeRequestService {
	return p.mergeRequestService
}

func (p *Provider) Comment() tp.CommentService {
	return p.commentService
}
