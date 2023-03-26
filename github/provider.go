package github

import (
	"context"
	gh "github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
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

func (p *Provider) Commit() types.CommitService {
	return p.commitService
}

func (p *Provider) Reference() types.ReferenceService {
	return p.referenceService
}

func (p *Provider) MergeRequest() types.MergeRequestService {
	return p.mergeRequestService
}

func (p *Provider) Comment() types.CommentService {
	return p.commentService
}
