package types

type GetCommitOption struct {
	Repo string
	SHA  string
}

type CreateCommitOption struct {
	Repo        *string
	Commit      *Commit
	PickMessage *string
	Parents     []Commit
}

type Commit interface {
	SHA() string
	Tree() Tree
	Message() string
}

type Tree interface {
	SHA() string
	Entries() []TreeEntry
	Truncated() bool
}

type TreeEntry interface {
	SHA() string
	Path() string
	Mode() string
	Type() string
	Size() int
	Content() string
	Url() string
}
