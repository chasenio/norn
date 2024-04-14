package pick

import (
	"context"
	"errors"
	"fmt"
	"github.com/kentio/norn/internal"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

type Service struct {
	provider tp.Provider
	branches []string
}

type CherryPick struct {
	SHA    string
	Repo   string
	Target string
}

type MergeCommentOpt struct {
	branches []string
	done     []string
	failed   []string
}

type Task struct {
	Repo           string
	Branches       []string // target branches
	Form           string   // from branch
	SHA            *string
	MergeRequestID string
	IsSummary      bool // generate summary comment
}

func NewPickService(provider tp.Provider, branches []string) *Service {
	return &Service{provider: provider, branches: branches}
}

func (s *Service) DoPick(ctx context.Context, opt *CherryPick) error {
	if s.provider == nil || opt == nil {
		return ErrInvalidOptions
	}

	// 1. get reference
	reference, err := s.provider.Reference().Get(ctx, &tp.GetRefOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
	})
	if err != nil {
		return err
	}

	// 2. get commit
	commit, err := s.provider.Commit().Get(ctx, &tp.GetCommitOption{
		Repo: opt.Repo,
		SHA:  opt.SHA,
	})
	if err != nil {
		return err
	}

	// 2.1 if last commit message is same as pick message, skip
	// TODO enhance the logic, last commit message may not be the pick message
	lastCommit, err := s.provider.Commit().Get(ctx, &tp.GetCommitOption{Repo: opt.Repo, SHA: reference.SHA})
	if err != nil {
		logrus.Warnf("failed to get last commit: %+v", err)
		return err
	}
	pickMessage := fmt.Sprintf("%s\n\n(cherry picked from commit %s)", commit.Message(), opt.SHA[:7])

	// if match message, skip
	lastCommitMessageMd5 := sumMd5(lastCommit.Message())
	pickMessageMd5 := sumMd5(pickMessage)
	logrus.Debugf("last commit message md5: %s, pick message md5: %s", lastCommitMessageMd5, pickMessageMd5)
	if lastCommitMessageMd5 == pickMessageMd5 {
		logrus.Warnf("reference %s last commit message is same as pick message, skip", reference.SHA)
		return errors.New("last commit message is same as pick message")
	}
	// 3. create commit
	createCommit, err := s.provider.Commit().Create(ctx, &tp.CreateCommitOption{
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
	_, err = s.provider.Reference().Update(ctx, &tp.UpdateOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
		SHA:  createCommit.SHA(),
	})

	if err != nil {
		return err
	}

	return nil
}

// RenderComment submit pick summary comment
func (s *Service) RenderComment(ctx context.Context, do *Task) error {
	// Check if the comment is already submitted
	// if exists, regen summary
	comment, err := s.ExistSummary(ctx, do.Repo, do.MergeRequestID)
	if err != nil {
		logrus.Debugf("ExistSummary failed: %+v", err)
		return err
	}
	// generate branch list of comment body
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
	summaryComment, err := NewSummaryComment(do.IsSummary, &MergeCommentOpt{branches: summaryBranches})
	if err != nil {
		return err
	}
	logrus.Infof("Submit summary comment: %s", summaryComment)

	switch comment {
	case nil:
		// if not exists, submit summary comment
		// submit comment
		_, err = s.provider.Comment().Create(ctx, &tp.CreateCommentOption{
			MergeRequestID: do.MergeRequestID,
			Body:           summaryComment,
			Repo:           do.Repo,
		},
		)
		if err != nil {
			return err
		}
	default:
		// if exists, update the comment
		logrus.Infof("pick comment already exists, regenerate summary comment.")
		_, err = s.provider.Comment().Update(ctx, &tp.UpdateCommentOption{
			CommentID: comment.CommentID(),
			Body:      summaryComment,
			Repo:      do.Repo,
		})
		if err != nil {
			return err
		}
	}

	logrus.Infof("Success to submit summary comment")
	return nil
}

// DoPickToBranches DoPick commits from one branches to another
func (s *Service) DoPickToBranches(ctx context.Context, do *Task) (done []string, failed []string, err error) {

	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: do.MergeRequestID, Repo: do.Repo})

	if FindSummaryWithFlag(comments, tp.CherryPickSummaryFlag) == nil {
		logrus.Errorf("not found pick comment")
		return nil, nil, errors.New("not found pick comment")
	}
	logrus.Debugf("Start to pick ...")

	// 获取选中的分支
	selectedBranches, err := s.GetSelectedRefByMergeReqeust(ctx, do.Repo, do.MergeRequestID)
	if err != nil {
		logrus.Warnf("Get Select Ref failed: %+v", err)
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
		if !internal.StringInSlice(branch, do.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, do.Branches)
			continue
		}

		logrus.Debugf("Picking %s to %s", do.SHA, branch)
		// DoPick commits
		pickOption := &CherryPick{
			SHA:    *do.SHA,
			Repo:   do.Repo,
			Target: branch,
		}
		err = s.DoPick(ctx, pickOption)
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
	pickResultComment, err := NewSummaryComment(do.IsSummary, &MergeCommentOpt{done: done, failed: failed})
	if err != nil {
		logrus.Errorf("Generate pick result comment failed: %s", err)
		return nil, nil, err
	}

	// submit pick result to merge request
	_, err = s.provider.Comment().Create(ctx, &tp.CreateCommentOption{
		Repo:           do.Repo,
		MergeRequestID: do.MergeRequestID,
		Body:           pickResultComment,
	})
	if err != nil {
		return done, failed, err
	}
	return done, failed, nil
}

func (s *Service) DoWithOpt(ctx context.Context, opt *Task) error {
	var err error
	if opt.IsSummary {
		err = s.RenderComment(ctx, opt)
		if err != nil {
			logrus.Errorf("do summary err: %s", err)
		}
	} else {
		_, _, err = s.DoPickToBranches(ctx, opt)
		if err != nil {
			logrus.Errorf("do pick err: %s", err)
		}
	}
	return err
}

// GetSelectedRefByMergeReqeust get selected reference by merge request
func (s *Service) GetSelectedRefByMergeReqeust(ctx context.Context, repo string, mergeRequestID string) (selectedBranches []string, err error) {
	// get merge request comments
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, err
	}

	// find comment with flag
	for _, comment := range comments {
		if strings.Contains(comment.Body(), tp.CherryPickSummaryFlag) {
			// parse selected reference
			selectedBranches = ParseSelectedBranches(comment.Body())
			return selectedBranches, nil
		}
	}
	return nil, nil
}

// ExistSummary check if summary comment is exist
func (s *Service) ExistSummary(ctx context.Context, repo string, mergeRequestID string) (tp.Comment, error) {
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, err
	}
	return FindSummaryWithFlag(comments, tp.CherryPickSummaryFlag), nil
}

// FindSummaryWithFlag check if comment is in merge request
func FindSummaryWithFlag(comments []tp.Comment, flag string) tp.Comment {
	for _, c := range comments {
		if strings.Contains(c.Body(), flag) {
			return c
		}
	}
	return nil
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

// NewSummaryComment generate task comment for pr
func NewSummaryComment(isSummary bool, opt *MergeCommentOpt) (string, error) {
	var summary string
	// generate comment for summary, before pick
	if isSummary {
		taskBranchLine, err := NewSelectComment(tp.CherryPickTaskSummaryTemplate, opt.branches)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		return taskBranchLine.String(), nil
	}
	// generate comment for done and failed, after pick
	var doneString, failedString string

	// render done summary
	if len(opt.done) > 0 {
		logrus.Debugf("render done summary: %s", opt.done)
		taskBranchLine, err := NewItemComment(tp.CherryPickTaskDoneTemplate, opt.done)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		doneString = taskBranchLine.String()
	}

	// render failed summary
	if len(opt.failed) > 0 {

		logrus.Debugf("render failed summary: %s", opt.failed)
		taskBranchLine, err := NewItemComment(tp.CherryPickTaskFailedTemplate, opt.failed)
		if err != nil {
			logrus.Errorf("failed to execute template: %s \n branches: %s \n err: %+v", tp.CherryPickTaskFailedTemplate, opt.failed, err)
			return "", err
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
		logrus.Warnf("Failed to execute template: %s \n branches: %s \n err: %+v", layout, branches, err)
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
		logrus.Warnf("Failed to execute template: %s \n branches: %s err: %+v", layout, branches, err)
		return content, err
	}
	return content, nil
}
