package types

type ProviderType string

const (
	GitHubProvider ProviderType = "github"
	GitlabProvider ProviderType = "gitlab"
)

type CreateProviderOption struct {
	Token     string
	BaseUrl   *string
	UploadUrl *string // GitHub Enterprise only
}

type Provider interface {
	Commit() CommitService
	Reference() ReferenceService
	MergeRequest() MergeRequestService
	Comment() CommentService
	ProviderID() ProviderType
	Pick() PickService
}
