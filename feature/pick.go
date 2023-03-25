package feature

import (
	"context"
	"fmt"
	"github.com/kentio/norn/global"
	"github.com/kentio/norn/types"
	"github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

type PickOption struct {
	SHA    string
	Repo   string
	Target string
}

type MergeCommentOpt struct {
	branches []string
	done     []string
	failed   []string
}

type PickToRefMROpt struct {
	Repo           string
	Branches       []string
	Form           string
	SHA            string
	MergeRequestID string
	IsSummaryTask  bool
}

func Pick(ctx context.Context, provider types.Provider, opt *PickOption) error {
	if provider == nil || opt == nil {
		return ErrInvalidOptions
	}

	// 1. get reference
	reference, err := provider.Reference().Get(ctx, &types.GetRefOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
	})
	if err != nil {
		return err
	}

	// 2. get commit
	commit, err := provider.Commit().Get(ctx, &types.GetCommitOption{
		Repo: opt.Repo,
		SHA:  opt.SHA,
	})
	if err != nil {
		return err
	}

	// 2.1 if last commit message is same as pick message, skip
	lastCommit, err := provider.Commit().Get(ctx, &types.GetCommitOption{Repo: opt.Repo, SHA: reference.SHA})
	if err != nil {
		logrus.Debugf("failed to get last commit: %+v", err)
		return err
	}
	pickMessage := fmt.Sprintf("Cherry-pick from %s\nSource Commit Message:\n%s", opt.SHA, commit.Message())

	// if match message, skip
	lastCommitMessageMd5 := sumMd5(lastCommit.Message())
	pickMessageMd5 := sumMd5(pickMessage)
	logrus.Debugf("last commit message md5: %s, pick message md5: %s", lastCommitMessageMd5, pickMessageMd5)
	if lastCommitMessageMd5 == pickMessageMd5 {
		logrus.Debugf("reference %s last commit message is same as pick message, skip", reference.SHA)
		return nil
	}
	// 3. create commit
	createCommit, err := provider.Commit().Create(ctx, &types.CreateCommitOption{
		Repo:        opt.Repo,
		Tree:        commit.Tree(),
		SHA:         commit.SHA(),
		PickMessage: pickMessage,
		Parents: []string{
			reference.SHA,
			commit.SHA(),
		}})

	if err != nil {
		return err
	}

	// 4. update reference
	_, err = provider.Reference().Update(ctx, &types.UpdateOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
		SHA:  createCommit.SHA(),
	})

	if err != nil {
		return err
	}

	return nil
}

// DoPickToBranchesFromMergeRequest Pick commits from one branch to another
func DoPickToBranchesFromMergeRequest(ctx context.Context, provider types.Provider, do *PickToRefMROpt) (done []string, failed []string, err error) {

	logrus.Debugf("Branches: %s", do.Branches)

	if do.IsSummaryTask {
		// Add Comment to merge request
		logrus.Infof("Add comment to merge request")
		// generate comment
		logrus.Debugf("Generate comment")
		summaryComment, err := NewMergeReqeustComment(do.IsSummaryTask, &MergeCommentOpt{branches: do.Branches})
		if err != nil {
			return nil, nil, err
		}
		// submit comment
		_, err = provider.Comment().Create(ctx, &types.CreateCommentOption{
			MergeRequestID: do.MergeRequestID,
			Body:           summaryComment,
			Repo:           do.Repo,
		},
		)
		logrus.Debugf("Comment: %s", summaryComment)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}
	logrus.Debugf("Pick commits from %s to %s", do.Form, do.Branches)

	// 获取选中的需要pick的分支
	selectedBranches, err := GetSelectedRefByMergeReqeust(ctx, provider, do.Repo, do.MergeRequestID)
	if err != nil {
		logrus.Debugf("GetSelectedRefByMergeReqeust failed: %+v", err)
		return nil, nil, err
	}

	if len(selectedBranches) == 0 {
		logrus.Infof("No selected branches")
		return nil, nil, nil
	}

	logrus.Infof("Selected branches: %s", selectedBranches)

	launchPick := false
	// Pick commits from one branch to another
	for _, branch := range selectedBranches {
		logrus.Debugf("Branch: %s", branch)
		if branch == do.Form {
			logrus.Debugf("Launch pick: %s", branch)
			launchPick = true
			continue // skip the branch, and pick commits from the next branch
		}
		if !launchPick {
			logrus.Debugf("Skip pick: %s", branch)
			continue
		}

		// if select branch not in defined branches, skip
		if !global.StringInSlice(branch, do.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, do.Branches)
			continue
		}

		logrus.Debugf("Picking %s to %s", do.SHA, branch)
		// Pick commits
		pickOption := &PickOption{
			SHA:    do.SHA,
			Repo:   do.Repo,
			Target: branch,
		}
		err = Pick(ctx, provider, pickOption)
		if err != nil {
			failed = append(failed, branch)
			continue
		}
		done = append(done, branch)
		logrus.Infof("Picked %s to %s", pickOption.SHA, pickOption.Target)
	}

	logrus.Debugf("Done: %s Failed: %s", done, failed)

	if len(done) == 0 && len(failed) == 0 {
		logrus.Debugf("No branch to pick")
		return nil, nil, nil
	}

	// generate comment
	logrus.Debugf("Generate pick result comment")
	pickResultComment, err := NewMergeReqeustComment(do.IsSummaryTask, &MergeCommentOpt{done: done, failed: failed})
	if err != nil {
		logrus.Debugf("Generate pick result comment failed: %s", err)
		return nil, nil, err
	}

	// submit pick result to merge request
	_, err = provider.Comment().Create(ctx, &types.CreateCommentOption{
		Repo:           do.Repo,
		MergeRequestID: do.MergeRequestID,
		Body:           pickResultComment,
	})
	if err != nil {
		return done, failed, err
	}
	return done, failed, nil
}

// GetSelectedRefByMergeReqeust get selected reference by merge request
func GetSelectedRefByMergeReqeust(ctx context.Context, p types.Provider, repo string, mergeRequestID string) (selectedBranches []string, err error) {
	// get merge request comments
	comments, err := p.Comment().Find(ctx, &types.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
	if err != nil {
		logrus.Debugf("Get merge request comments failed: %s", err)
		return nil, err
	}

	// find comment with flag
	for _, comment := range comments {
		if strings.Contains(comment.Body(), global.CherryPickSummaryFlag) {
			// parse selected reference
			selectedBranches = ParseSelectedBranches(comment.Body())
			return selectedBranches, nil
		}
	}
	return nil, nil
}

// ParseSelectedBranches parse selected branches from comment
func ParseSelectedBranches(comment string) (selectedBranches []string) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		if strings.Contains(line, "- [x]") {
			line = strings.ReplaceAll(line, "- [x] ", "") // remove "- [x] "
			line = strings.ReplaceAll(line, " ", "")      // remove " "
			if line == "" {
				continue
			}
			selectedBranches = append(selectedBranches, line)
		}
	}
	return selectedBranches
}

// NewMergeReqeustComment generate comment for merge request
func NewMergeReqeustComment(isSummary bool, opt *MergeCommentOpt) (summary string, err error) {
	if isSummary {
		taskBranchLine, err := NewCommentContent(global.CherryPickTaskSummaryTemplate, opt.branches)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		return taskBranchLine.String(), nil
	}
	var doneString, failedString string

	// render done summary
	if len(opt.done) > 0 {
		logrus.Debugf("render done summary: %s", opt.done)
		taskBranchLine, err := NewCommentContent(global.CherryPickTaskDoneTemplate, opt.done)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		doneString = taskBranchLine.String()
	}

	// render failed summary
	if len(opt.failed) > 0 {

		logrus.Debugf("render failed summary: %s", opt.failed)
		taskBranchLine, err := NewCommentContent(global.CherryPickTaskFailedTemplate, opt.failed)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		failedString = taskBranchLine.String()
	}

	if len(opt.done) > 0 {
		summary += doneString
	}

	// if both done and failed, add a separator
	if len(opt.failed) > 0 && len(opt.done) > 0 {
		summary += "---\n" +
			failedString
	} else {
		summary += failedString
	}

	return summary, nil
}

// NewCommentContent generate comment content
func NewCommentContent(layout string, branches []string) (content strings.Builder, err error) {
	var taskBranchLine strings.Builder
	type Msg struct {
		Message string `json:"message"`
	}
	for _, branch := range branches {
		taskBranchLine.WriteString("- [x] " + branch + "\n")
	}
	tpl := template.Must(template.New("message").Parse(layout))
	data := Msg{
		Message: taskBranchLine.String(),
	}
	err = tpl.Execute(&content, data)
	if err != nil {
		logrus.Debugf("Failed to execute template: %s \n branches: %s", layout, branches)
		return content, fmt.Errorf("failed to execute template: %w", err)
	}
	return content, nil
}
