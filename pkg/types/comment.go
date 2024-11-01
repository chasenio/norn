package types

import "context"

type FindCommentOption struct {
	Repo           string
	MergeRequestID string
	CommentIds     []string
}

type CommentService interface {
	Find(ctx context.Context, opt *FindCommentOption) ([]Comment, error)
	Create(ctx context.Context, opt *CreateCommentOption) (Comment, error)
	Update(ctx context.Context, opt *UpdateCommentOption) (Comment, error)
	Delete(ctx context.Context, opt *DeleteCommentOption) error
}

type Comment interface {
	CommentID() string
	Body() string
}
