package types

import "context"

type MergeRequestState int

type MergeRequestService interface {
	Get(ctx context.Context, opt *GetMergeRequestOption) (MergeRequest, error)
}

type MergeRequest interface {
	State() MergeRequestState
	MergeId() string
	Title() string
	Description() string
}

type GetMergeRequestOption struct {
	Repo    string
	MergeID string
}

type CreateCommentOption struct {
	// MergeRequestID is the ID of the merge request to comment on. also known as IssueID
	MergeRequestID string
	Repo           string
	Body           string
}

type UpdateCommentOption struct {
	// CommentID is the ID of the comment to update.
	CommentID      string
	Repo           string
	Body           string
	MergeRequestID string
}
