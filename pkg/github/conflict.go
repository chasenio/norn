package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

type CreatePatchOption struct {
	Client *gh.Client
	Commit string
	Owner  string
	Repo   string
	Pr     int
}

func CreatePatchWithClient(ctx context.Context, opt *CreatePatchOption) (string, error) {
	patch, _, err := opt.Client.PullRequests.GetRaw(ctx, opt.Owner, opt.Repo, opt.Pr, gh.RawOptions{
		Type: gh.Patch,
	})
	if err != nil {
		return "", err
	}
	return patch, nil
}

type CheckoutOption struct {
	Branch   string
	RepoPath string // repo path
}

func Checkout(opt *CheckoutOption) error {
	remote := fmt.Sprintf("remotes/origin/%s", opt.Branch)
	cmd := exec.Command("git", "checkout", "-b", opt.Branch, remote, "-f")
	cmd.Dir = opt.RepoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("checkout branch failed: %s", err.Error())
		return err
	}
	return nil
}

type ApplyPatchOption struct {
	Patch    string // patch file path
	RepoPath string // repo path
	Check    bool   // check apply patch, but not apply
}

func ApplyPatch(opt *ApplyPatchOption) error {
	// check apply patch, but not apply
	var cmd = &exec.Cmd{}
	if opt.Check {
		cmd = exec.Command("git", "apply", opt.Patch, "--check")
	} else {
		cmd = exec.Command("git", "apply", opt.Patch)
	}
	cmd.Dir = opt.RepoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return tp.ErrConflict
	}
	return nil
}

type CherryPickOption struct {
	RepoPath string
	Commit   string
}

func CherryPick(opt *CherryPickOption) error {
	defer func() {
		cmd := exec.Command("git", "cherry-pick", "--abort")
		cmd.Dir = opt.RepoPath
		_ = cmd.Run()
	}()

	cmd := exec.Command("git", "cherry-pick", opt.Commit)
	cmd.Dir = opt.RepoPath
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
