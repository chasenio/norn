package github

import (
	"context"
	"errors"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
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
		logrus.Errorf("create path failed: %s", err.Error())
		return "", err
	}
	return patch, nil
}

type CheckoutOption struct {
	Branch   string
	RepoPath string // repo path
}

func Checkout(opt *CheckoutOption) error {
	var stdout strings.Builder
	// check branch exist
	d := exec.Command("git", "branch", "-d", opt.Branch, "-f")
	d.Dir = opt.RepoPath
	d.Stdout = &stdout
	d.Stderr = &stdout
	if err := d.Run(); err != nil {
		logrus.Warnf("delete branch err: %s", stdout.String())
	}
	stdout.Reset()
	// checkout branch
	remote := fmt.Sprintf("remotes/origin/%s", opt.Branch)
	cmd := exec.Command("git", "checkout", "-b", opt.Branch, remote, "-f")
	cmd.Dir = opt.RepoPath
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		logrus.Errorf("checkout branch failed: %s\nerr: %s", stdout.String(), err.Error())
		return errors.New("checkout branch failed")
	}
	return nil
}

type ApplyPatchOption struct {
	Patch    string // patch file path
	RepoPath string // repo path
	Check    bool   // check apply patch, but not apply
}

func ApplyPatch(opt *ApplyPatchOption) error {
	var stdout strings.Builder
	// check apply patch, but not apply
	var cmd = &exec.Cmd{}
	if opt.Check {
		cmd = exec.Command("git", "apply", opt.Patch, "--check")
	} else {
		cmd = exec.Command("git", "apply", opt.Patch)
	}
	cmd.Dir = opt.RepoPath
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	if err != nil {
		logrus.Warnf("apply patch failed: %s err: %s", stdout.String(), err.Error())
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
