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
	MergeRequestID string
	Repo           string
	Body           string
}
