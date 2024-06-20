package github

import (
	"context"
	"errors"
	gh "github.com/google/go-github/v62/github"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"os"
)

type Commit struct {
	sha     string
	tree    *Tree
	message string
}

type Tree struct {
	sha       string
	entries   []tp.TreeEntry
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

// Get Commit returns the commit for the given path.
func (s *CommitService) Get(ctx context.Context, opt *tp.GetCommitOption) (tp.Commit, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Get Commit Opt: %+v", *opt)
	commit, response, err := s.client.Repositories.GetCommit(ctx, repoOpt.Owner, repoOpt.Repo, opt.SHA, nil)
	if err != nil {
		logrus.Debugf("Get Commit Error: %+v", err)
		return nil, err
	}
	logrus.Debugf("Get Commit Response: %+v", *response)
	return newCommit(commit), nil
}

// CheckConflict Check Conflict
func (s *CommitService) CheckConflict(ctx context.Context, opts *tp.CheckConflictOption) error {

	if opts.Mode != tp.WithCommand {
		return errors.New("not support check conflict with API")
	}

	repoOpt, err := parseRepo(opts.Repo)
	// create patch
	content, err := CreatePatchWithClient(ctx, &CreatePatchOption{
		Client: s.client,
		Commit: opts.Commit,
		Owner:  repoOpt.Owner,
		Repo:   repoOpt.Repo,
		Pr:     opts.Pr,
	})
	if err != nil {
		return err
	}

	// write patch to temp file
	patch, err := os.CreateTemp(os.TempDir(), "patch")
	if err != nil {
		return err
	}
	_, err = patch.Write([]byte(content))
	if err != nil {
		return err
	}
	// remove patch file
	defer func() {
		err := os.Remove(patch.Name())
		if err != nil {
			logrus.Warnf("Remove Patch File Error: %v", err)
		}
		err = patch.Close()
		if err != nil {
			logrus.Warnf("Close Patch File Error: %v", err)
		}
	}()

	// checkout target branch
	err = Checkout(&CheckoutOption{
		Branch:   opts.Target,
		RepoPath: opts.RepoPath, // default path is current directory
	})
	if err != nil {
		return err
	}

	// check patch
	err = ApplyPatch(&ApplyPatchOption{
		Patch:    patch.Name(),
		RepoPath: opts.RepoPath,
		Check:    true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Create Commit creates a new commit.
func (s *CommitService) Create(ctx context.Context, opt *tp.CreateCommitOption) (tp.Commit, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}

	var parents []*gh.Commit
	logrus.Debugf("Create Commit Opt: %+v", *opt)
	for _, p := range opt.Parents {
		parents = append(parents, &gh.Commit{
			SHA: gh.String(p),
		})
	}

	commit, _, err := s.client.Git.CreateCommit(ctx, repoOpt.Owner, repoOpt.Repo, &gh.Commit{
		Message: gh.String(opt.PickMessage),
		Parents: parents,
		Tree: &gh.Tree{
			SHA:       gh.String(opt.Tree.SHA()),
			Entries:   nil,
			Truncated: gh.Bool(false),
		},
	}, nil)
	if err != nil {
		logrus.Errorf("Create Commit Error: %v", err)
		return nil, err
	}

	return newGithubCommit(commit), nil
}

func newGithubCommit(c *gh.Commit) *Commit {
	return &Commit{
		sha:     *c.SHA,
		message: *c.Message,
	}
}

func newCommit(commit *gh.RepositoryCommit) *Commit {
	if commit == nil {
		return nil
	}

	entrys := lo.Map(commit.Commit.Tree.Entries, func(t *gh.TreeEntry, i int) tp.TreeEntry {
		return NewTreeEntry(*t)
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

func NewTree(tree gh.Tree) tp.Tree {
	entry := lo.Map(tree.Entries, func(t *gh.TreeEntry, i int) tp.TreeEntry {
		return NewTreeEntry(*t)
	})

	var truncated bool
	if tree.Truncated != nil {
		truncated = *tree.Truncated
	}

	return &Tree{
		sha:       *tree.SHA,
		entries:   entry,
		truncated: truncated,
	}
}

func NewTreeEntry(entry gh.TreeEntry) tp.TreeEntry {
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

func (c *Commit) Tree() tp.Tree {
	return c.tree
}

func (c *Commit) Message() string {
	return c.message
}

// SHA Tree returns the tree for the given path.
func (t *Tree) SHA() string {
	return t.sha
}

func (t *Tree) Entries() []tp.TreeEntry {
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
