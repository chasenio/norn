package types

type Provider interface {
	Commit() CommitService
	Reference() ReferenceService
	MergeRequest() MergeRequestService
	Comment() CommentService
}
