package types

type ProviderType string

const (
	GitHubProvider ProviderType = "github"
)

type Provider interface {
	Commit() CommitService
	Reference() ReferenceService
	MergeRequest() MergeRequestService
	Comment() CommentService
	ProviderID() ProviderType
	Pick() PickService
}
