package types

type Provider interface {
	Commit() CommitService
	Reference() ReferenceService
}
