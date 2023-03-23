package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Commit struct {
	sha     string
	tree    *Tree
	message string
}

type Tree struct {
	sha       string
	entries   []types.TreeEntry
	truncated bool
}

type TreeEntry struct {
	sha       string
	path      string
	mode      string
	entryType string
	size      int
	content   string
	url       string
}

type CommitService struct {
	client *gh.Client
}

func NewCommitService(client *gh.Client) *CommitService {
	return &CommitService{
		client: client,
	}
}

func (s *CommitService) Get(ctx context.Context, opt *types.GetCommitOption) (*Commit, error) {
	if opt == nil {
		return nil, ErrInvalidOptions
	}
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	commit, response, err := s.client.Repositories.GetCommit(ctx, repoOpt.Owner, repoOpt.Repo, opt.SHA, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %v", err)
	}
	logrus.Infof("response: %v", response)
	return newCommit(commit), nil
}

func (s *CommitService) Create(ctx context.Context, opt *types.CreateCommitOption) (*Commit, error) {
	if opt == nil {
		return nil, ErrInvalidOptions
	}
	repoOpt, err := parseRepo(*opt.Repo)
	if err != nil {
		return nil, err
	}

	var parents []*gh.Commit

	for _, p := range opt.Parents {
		parents = append(parents, &gh.Commit{
			SHA: gh.String(p.SHA()),
		})
	}

	commit, response, err := s.client.Git.CreateCommit(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Commit{
		Message: opt.PickMessage,
		Tree: &gh.Tree{
			SHA: opt.Tree.SHA(),
		},
		Parents: parents,
	})

	return nil, nil
}

func newCommit(commit *gh.RepositoryCommit) *Commit {
	if commit == nil {
		return nil
	}

	entrys := lo.Map(commit.Commit.Tree.Entries, func(t *gh.TreeEntry, i int) types.TreeEntry {
		return newTreeEntry(*t)
	})

	var truncated bool
	if commit.Commit.Tree.Truncated != nil {
		truncated = *commit.Commit.Tree.Truncated
	}

	return &Commit{
		sha: *commit.SHA,
		tree: &Tree{
			sha:       *commit.Commit.Tree.SHA,
			entries:   entrys,
			truncated: truncated,
		},
		message: *commit.Commit.Message,
	}
}

func newTreeEntry(entry gh.TreeEntry) types.TreeEntry {
	return &TreeEntry{
		sha:       *entry.SHA,
		path:      *entry.Path,
		mode:      *entry.Mode,
		entryType: *entry.Type,
		size:      *entry.Size,
		content:   *entry.Content,
		url:       *entry.URL,
	}
}

// SHA Commit returns the commit for the given path.
func (c *Commit) SHA() string {
	return c.sha
}

func (c *Commit) Tree() types.Tree {
	return c.tree
}

func (c *Commit) Message() string {
	return c.message
}

// SHA Tree returns the tree for the given path.
func (t *Tree) SHA() string {
	return t.sha
}

func (t *Tree) Entries() []types.TreeEntry {
	return t.entries
}

func (t *Tree) Truncated() bool {
	return t.truncated
}

// SHA TreeEntry returns the tree entry for the given path.
func (t *TreeEntry) SHA() string {
	return t.sha
}

func (t *TreeEntry) Path() string {
	return t.path
}

func (t *TreeEntry) Mode() string {
	return t.mode
}

func (t *TreeEntry) Type() string {
	return t.entryType
}

func (t *TreeEntry) Size() int {
	return t.size
}

func (t *TreeEntry) Content() string {
	return t.content
}

func (t *TreeEntry) Url() string {
	return t.url
}
