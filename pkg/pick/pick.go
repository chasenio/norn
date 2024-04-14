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

type Result struct {
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

// PerformPickToBranches PerformPick commits from one branches to another
func (s *Service) PerformPickToBranches(ctx context.Context, task *Task) (done []string, failed []string, err error) {

	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: task.MergeRequestID, Repo: task.Repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, nil, err
	}

	if FindSummaryWithFlag(comments, tp.CherryPickSummaryFlag) == nil {
		logrus.Errorf("not found pick comment")
		return nil, nil, errors.New("not found pick comment")
	}
	logrus.Debugf("Start to pick ...")

	// 获取选中的分支
	selectedBranches, err := s.GetSelectedBranches(ctx, task.Repo, task.MergeRequestID)
	if err != nil {
		logrus.Warnf("Get Select Ref failed: %+v", err)
		return nil, nil, err
	}

	if len(selectedBranches) == 0 {
		logrus.Infof("No selected branches")
		return nil, nil, nil
	}

	logrus.Infof("Selected branches: %s", selectedBranches)

	// PerformPick commits from one branch to another
	for _, branch := range selectedBranches {
		if branch == task.Form {
			logrus.Debugf("Skip form branch: %s", branch)
			continue // skip the branch, and pick commits from the next branch
		}
		logrus.Debugf("Branch: %s", branch)

		// if select branch not in defined branches, skip
		if !internal.StringInSlice(branch, task.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, task.Branches)
			continue
		}

		logrus.Debugf("Picking %s to %s", *task.SHA, branch)
		// PerformPick commits
		pickOption := &CherryPick{
			SHA:    *task.SHA,
			Repo:   task.Repo,
			Target: branch,
		}
		err = s.PerformPick(ctx, pickOption)
		if err != nil {
			failed = append(failed, branch)
			continue
		}
		done = append(done, branch)
		logrus.Infof("Picked %s to %s", pickOption.SHA, pickOption.Target)
	}

	logrus.Infof("Done: %s Failed: %s", done, failed)

	if len(done) == 0 && len(failed) == 0 {
		logrus.Warnf("No branch to pick")
		return nil, nil, nil
	}

	// generate comment
	logrus.Infof("Generate pick result comment")
	pickResultComment, err := NewSummaryComment(task.IsSummary, &Result{done: done, failed: failed})
	if err != nil {
		logrus.Errorf("Generate pick result comment failed: %s", err)
		return nil, nil, err
	}

	// submit pick result to merge request
	_, err = s.provider.Comment().Create(ctx, &tp.CreateCommentOption{
		Repo:           task.Repo,
		MergeRequestID: task.MergeRequestID,
		Body:           pickResultComment,
	})
	if err != nil {
		return done, failed, err
	}
	return done, failed, nil
}

func (s *Service) PerformPick(ctx context.Context, opt *CherryPick) error {
	if s.provider == nil || opt == nil {
		logrus.Error("provider or opt is nil")
		return ErrInvalidOptions
	}

	// 1. get reference
	reference, err := s.provider.Reference().Get(ctx, &tp.GetRefOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
	})
	if err != nil {
		logrus.Errorf("failed to get reference: %+v", err)
		return err
	}

	// 2. get commit
	commit, err := s.provider.Commit().Get(ctx, &tp.GetCommitOption{
		Repo: opt.Repo,
		SHA:  opt.SHA,
	})
	if err != nil {
		logrus.Errorf("failed to get commit: %+v", err)
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
		logrus.Errorf("failed to create commit: %+v", err)
		return err
	}

	// 4. update reference
	_, err = s.provider.Reference().Update(ctx, &tp.UpdateOption{
		Repo: opt.Repo,
		Ref:  fmt.Sprintf("refs/heads/%s", opt.Target),
		SHA:  createCommit.SHA(),
	})

	if err != nil {
		logrus.Errorf("failed to update reference: %+v", err)
		return err
	}
	return nil
}

// CreateSummaryWithTask submit pick summary comment
func (s *Service) CreateSummaryWithTask(ctx context.Context, task *Task) error {
	// generate branch list of comment body
	targets := generateTargetBranches(task)
	logrus.Debugf("Summary branches: %+v", targets)
	if len(targets) == 0 {
		logrus.Infof("No summary branches, exit")
		return nil
	}

	// generate comment body
	summaryComment, err := NewSummaryComment(task.IsSummary, &Result{branches: targets})
	if err != nil {
		logrus.Errorf("NewSummaryComment failed: %+v", err)
		return err
	}

	// Check if the comment is already submitted
	// if exists, regen summary
	comment, err := s.CheckSummaryExist(ctx, task.Repo, task.MergeRequestID)
	if err != nil {
		logrus.Debugf("CheckSummaryExist failed: %+v", err)
		return err
	}

	switch comment {
	case nil:
		// if not exists, submit summary comment
		// submit comment
		_, err = s.provider.Comment().Create(ctx, &tp.CreateCommentOption{
			MergeRequestID: task.MergeRequestID,
			Body:           summaryComment,
			Repo:           task.Repo,
		},
		)
		if err != nil {
			return err
		}
	default:
		// diff summary branches and exist branches, if different, update the comment
		// if same, skip
		existSelected := parseSelectedBranches(comment.Body())
		if EqualSlice(existSelected, targets) {
			logrus.Infof("Summary branches are same as exist, skip")
			return nil
		}

		// update the comment
		logrus.Info("pick comment already exists, regenerate summary comment.")
		_, err = s.provider.Comment().Update(ctx, &tp.UpdateCommentOption{
			CommentID: comment.CommentID(),
			Body:      summaryComment,
			Repo:      task.Repo,
		})
		if err != nil {
			return err
		}
	}
	logrus.Infof("Submit summary comment: %s", summaryComment)
	return nil
}

func (s *Service) ProcessPick(ctx context.Context, task *Task) error {
	var err error
	if task.IsSummary {
		err = s.CreateSummaryWithTask(ctx, task)
		if err != nil {
			logrus.Errorf("create summary err: %s", err)
		}
	} else {
		_, _, err = s.PerformPickToBranches(ctx, task)
		if err != nil {
			logrus.Errorf("perform pick err: %s", err)
		}
	}
	return err
}

// GetSelectedBranches get selected reference by merge request
func (s *Service) GetSelectedBranches(ctx context.Context, repo string, mergeRequestID string) ([]string, error) {
	// get merge request comments
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: mergeRequestID, Repo: repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, err
	}

	var selected []string

	// find comment with flag
	for _, comment := range comments {
		if strings.Contains(comment.Body(), tp.CherryPickSummaryFlag) {
			// parse selected reference
			selected = parseSelectedBranches(comment.Body())
			return selected, nil
		}
	}
	return nil, nil
}

// CheckSummaryExist check if summary comment is exist
func (s *Service) CheckSummaryExist(ctx context.Context, repo string, mergeRequestID string) (tp.Comment, error) {
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

// NewSummaryComment generate task comment for pr
func NewSummaryComment(isSummary bool, opt *Result) (string, error) {

	// generate comment for summary, before pick
	if isSummary {
		content, err := NewSelectComment(tp.CherryPickTaskSummaryTemplate, opt.branches)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		return content.String(), nil
	}

	// generate comment for done and failed, after pick
	var doneString, failedString string

	// render done summary
	if len(opt.done) > 0 {
		logrus.Debugf("render done summary: %s", opt.done)
		content, err := NewResultComment(tp.CherryPickTaskDoneTemplate, opt.done)
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
		doneString = content.String()
	}

	// render failed summary
	if len(opt.failed) > 0 {

		logrus.Debugf("render failed summary: %s", opt.failed)
		content, err := NewResultComment(tp.CherryPickTaskFailedTemplate, opt.failed)
		if err != nil {
			logrus.Errorf("failed to execute template: %s \n branches: %s \n err: %+v", tp.CherryPickTaskFailedTemplate, opt.failed, err)
			return "", err
		}
		failedString = content.String()
	}
	var summary string

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

// NewResultComment generate comment content
func NewResultComment(layout string, branches []string) (content strings.Builder, err error) {
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
