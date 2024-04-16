package github

import (
	"context"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/pkg/types"
)

type Provider struct {
	providerID tp.ProviderType
	client     *gh.Client

	commitService       *CommitService
	referenceService    *ReferenceService
	mergeRequestService *PullRequestService
	commentService      *CommentService
	treeService         *TreeService
}

func NewProvider(ctx context.Context, token string) *Provider {
	client := NewGithubClient(ctx, token)
	return &Provider{
		providerID:          tp.GitHubProvider,
		commitService:       NewCommitService(client),
		referenceService:    NewReferenceService(client),
		mergeRequestService: NewPullRequestService(client),
		commentService:      NewCommentService(client),
		treeService:         NewTreeService(client),
	}
}

// NewProviderWithClient creates a new provider with the given client.
func NewProviderWithClient(client *gh.Client) *Provider {
	return &Provider{
		providerID:          tp.GitHubProvider,
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

func (p *Provider) ProviderID() tp.ProviderType {
	return p.providerID
}
