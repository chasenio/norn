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

type PickFeature struct {
	provider types.Provider
	branches []string
}

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

func NewPickFeature(provider types.Provider, branches []string) *PickFeature {
	return &PickFeature{provider: provider, branches: branches}
}

func (pick *PickFeature) DoPick(ctx context.Context, opt *PickOption) error {
	if pick.provider == nil || opt == nil {
		return ErrInvalidOptions
	}

	// 1. get reference
	reference, err := pick.provider.Reference().Get(ctx, &types.GetRefOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
	})
	if err != nil {
		return err
	}

	// 2. get commit
	commit, err := pick.provider.Commit().Get(ctx, &types.GetCommitOption{
		Repo: opt.Repo,
		SHA:  opt.SHA,
	})
	if err != nil {
		return err
	}

	// 2.1 if last commit message is same as pick message, skip
	lastCommit, err := pick.provider.Commit().Get(ctx, &types.GetCommitOption{Repo: opt.Repo, SHA: reference.SHA})
	if err != nil {
		logrus.Debugf("failed to get last commit: %+v", err)
		return err
	}
	pickMessage := fmt.Sprintf("%s\n\ncherry-pick from commit (%s)", commit.Message(), opt.SHA[:7])

	// if match message, skip
	lastCommitMessageMd5 := sumMd5(lastCommit.Message())
	pickMessageMd5 := sumMd5(pickMessage)
	logrus.Debugf("last commit message md5: %s, pick message md5: %s", lastCommitMessageMd5, pickMessageMd5)
	if lastCommitMessageMd5 == pickMessageMd5 {
		logrus.Debugf("reference %s last commit message is same as pick message, skip", reference.SHA)
		return nil
	}
	// 3. create commit
	createCommit, err := pick.provider.Commit().Create(ctx, &types.CreateCommitOption{
		Repo:        opt.Repo,
		Tree:        commit.Tree(),
		SHA:         commit.SHA(),
		PickMessage: pickMessage,
		Parents: []string{
			reference.SHA,
		}})

	if err != nil {
		return err
	}

	// 4. update reference
	_, err = pick.provider.Reference().Update(ctx, &types.UpdateOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
		SHA:  createCommit.SHA(),
	})

	if err != nil {
		return err
	}

	return nil
}

// DoPickSummaryComment submit pick summary comment
func (pick *PickFeature) DoPickSummaryComment(ctx context.Context, do *PickToRefMROpt) error {
	// Check if the comment is already submitted
	if isExties, err := pick.IsInMergeRequestComments(ctx, do.Repo, do.MergeRequestID); err != nil {
		logrus.Debugf("IsInMergeRequestComments failed: %+v", err)
		return err
	} else if isExties {
		logrus.Infof("Summary comment already exists, exit")
		return nil
	}

	// generate branch list
	var summaryBranches []string
	var startFlag bool
	for _, branch := range do.Branches {
		// 按照路径顺序， 在前面的被将被跳过
		if branch == do.Form {
			startFlag = true
			continue
		}
		if startFlag {
			summaryBranches = append(summaryBranches, branch)
			continue
		}
		logrus.Debugf("before branch: %s, form: %s ; skip", branch, do.Form)

	}
	logrus.Debugf("Summary branches: %+v", summaryBranches)
	if len(summaryBranches) == 0 {
		logrus.Infof("No summary branches, exit")
		return nil
	}

	// generate comment
	logrus.Debugf("Generate summary comment")
	summaryComment, err := NewMergeReqeustComment(do.IsSummaryTask, &MergeCommentOpt{branches: summaryBranches})
	if err != nil {
		return err
	}
	logrus.Infof("Submit summary comment: %s", summaryComment)
	// submit comment
	_, err = pick.provider.Comment().Create(ctx, &types.CreateCommentOption{
		MergeRequestID: do.MergeRequestID,
		Body:           summaryComment,
		Repo:           do.Repo,
	},
	)
	if err != nil {
		return err
	}
	logrus.Infof("Success to submit summary comment")
	return nil
}

// DoPickToBranchesFromMergeRequest DoPick commits from one branch to another
func (pick *PickFeature) DoPickToBranchesFromMergeRequest(ctx context.Context, do *PickToRefMROpt) (done []string, failed []string, err error) {

	comments, err := pick.provider.Comment().Find(ctx, &types.FindCommentOption{MergeRequestID: do.MergeRequestID, Repo: do.Repo})

	if !IsInMergeRequestComments(comments) {
		logrus.Infof("No pick comment")
		return nil, nil, nil
	}
	logrus.Debugf("Start to pick ...")

	// 获取选中的需要pick的分支
	selectedBranches, err := pick.GetSelectedRefByMergeReqeust(ctx, do.Repo, do.MergeRequestID)
	if err != nil {
		logrus.Debugf("Get Select Ref failed: %+v", err)
		return nil, nil, err
	}

	if len(selectedBranches) == 0 {
		logrus.Infof("No selected branches")
		return nil, nil, nil
	}

	logrus.Infof("Selected branches: %s", selectedBranches)

	// DoPick commits from one branch to another
	for _, branch := range selectedBranches {
		if branch == do.Form {
			logrus.Debugf("Skip form branch: %s", branch)
			continue // skip the branch, and pick commits from the next branch
		}
		logrus.Debugf("Branch: %s", branch)

		// if select branch not in defined branches, skip
		if !global.StringInSlice(branch, do.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, do.Branches)
			continue
		}

		logrus.Debugf("Picking %s to %s", do.SHA, branch)
		// DoPick commits
		pickOption := &PickOption{
			SHA:    do.SHA,
			Repo:   do.Repo,
			Target: branch,
		}
		err = pick.DoPick(ctx, pickOption)
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
	_, err = pick.provider.Comment().Create(ctx, &types.CreateCommentOption{
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
func (pick *PickFeature) GetSelectedRefByMergeReqeust(ctx context.Context, repo string, mergeRequestID string) (selectedBranches []string, err error) {
	// get merge request comments
	comments, err := pick.provider.Comment().Find(ctx, &types.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
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

func (pick *PickFeature) IsInMergeRequestComments(ctx context.Context, repo string, mergeRequestID string) (bool, error) {
	comments, err := pick.provider.Comment().Find(ctx, &types.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
	if err != nil {
		logrus.Debugf("Get merge request comments failed: %s", err)
		return false, err
	}
	return IsInMergeRequestComments(comments), nil
}

// IsInMergeRequestComments check if comment is in merge request
func IsInMergeRequestComments(comments []types.Comment) bool {
	for _, c := range comments {
		if strings.Contains(c.Body(), global.CherryPickSummaryFlag) {
			return true
		}
	}
	return false
}

// ParseSelectedBranches parse selected branches from comment
func ParseSelectedBranches(comment string) (selectedBranches []string) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		if strings.Contains(line, "- [x]") {
			line = strings.ReplaceAll(line, "- [x] ", "") // remove "- [x] "
			line = strings.ReplaceAll(line, " ", "")      // remove " "
			// remove enter
			line = strings.ReplaceAll(line, "\r", "")
			line = strings.ReplaceAll(line, "\n", "")

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
		taskBranchLine, err := NewSelectComment(global.CherryPickTaskSummaryTemplate, opt.branches)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		return taskBranchLine.String(), nil
	}
	var doneString, failedString string

	// render done summary
	if len(opt.done) > 0 {
		logrus.Debugf("render done summary: %s", opt.done)
		taskBranchLine, err := NewItemComment(global.CherryPickTaskDoneTemplate, opt.done)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		doneString = taskBranchLine.String()
	}

	// render failed summary
	if len(opt.failed) > 0 {

		logrus.Debugf("render failed summary: %s", opt.failed)
		taskBranchLine, err := NewItemComment(global.CherryPickTaskFailedTemplate, opt.failed)
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

// NewSelectComment generate comment content
func NewSelectComment(layout string, branches []string) (content strings.Builder, err error) {
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

// NewItemComment generate comment content
func NewItemComment(layout string, branches []string) (content strings.Builder, err error) {
	var taskBranchLine strings.Builder
	type Msg struct {
		Message string `json:"message"`
	}
	for _, branch := range branches {
		taskBranchLine.WriteString("- " + branch + "\n")
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
