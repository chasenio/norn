package pick

import (
	"context"
	"errors"
	"fmt"
	"github.com/kentio/norn/internal"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type Service struct {
	provider tp.Provider
	branches []string
}

type CherryPickOptions struct {
	SHA      string
	Repo     string
	Target   string
	RepoPath string
	Pr       int
}

type Mode int

const (
	MergeRequest Mode = iota
	CheeryPick
)

type Task struct {
	Repo           string
	Branches       []string // target branches
	Form           string   // from branch
	SHA            *string
	MergeRequestID string
	IsSummary      bool // generate summary comment
	PickMode       Mode
	RepoPath       string
}

type State string

const (
	SucceedState = "Succeed"
	FailedState  = "Failed"
	PendingState = "Pending"
)

type TaskResult struct {
	State  State
	Branch string
	Reason string
}

func NewPickService(provider tp.Provider, branches []string) *Service {
	return &Service{provider: provider, branches: branches}
}

func (s *Service) FindCommentWithTask(ctx context.Context, task *Task, flag string) ([]tp.Comment, tp.Comment, error) {
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: task.MergeRequestID, Repo: task.Repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, nil, err
	}

	if comment := FindSummaryWithFlag(comments, flag); comment != nil {
		logrus.Warnf("not found pick, end task")
		return comments, comment, nil
	}
	return comments, nil, nil
}

// PerformPickToBranches PerformPick commits from one branches to another
func (s *Service) PerformPickToBranches(ctx context.Context, task *Task, comment tp.Comment) (result []*TaskResult, err error) {

	logrus.Debugf("Start to pick ...")

	// get selected branches
	selected := parseSelectedBranches(comment.Body())

	if len(selected) == 0 {
		logrus.Warnf("no selected branches")
		return nil, nil
	}

	logrus.Infof("Selected branches: %s", selected)

	// PerformPick commits from one branch to another
	for _, branch := range selected {
		var state State
		if branch == task.Form {
			logrus.Debugf("Skip form branch: %s", branch)
			continue // skip the branch, and pick commits from the next branch
		}

		// if select branch not in defined branches, skip
		if !internal.StringInSlice(branch, task.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, task.Branches)
			continue
		}

		logrus.Debugf("Picking %s to %s", *task.SHA, branch)
		// PerformPick commits
		pr, _ := strconv.Atoi(task.MergeRequestID)
		err = s.PerformPick(ctx, &CherryPickOptions{
			SHA:      *task.SHA,
			Repo:     task.Repo,
			Target:   branch,
			RepoPath: task.RepoPath,
			Pr:       pr,
		})
		if err != nil {
			state = FailedState
			result = append(result, &TaskResult{State: state, Branch: branch, Reason: err.Error()})
		} else {
			state = SucceedState
			result = append(result, &TaskResult{State: state, Branch: branch})
		}
		logrus.Infof("Pick %s to %s %s", *task.SHA, branch, state)
	}
	logrus.Infof("Picke Result %v", result)

	if len(result) == 0 {
		logrus.Warnf("No branch to pick")
		return nil, nil
	}

	// generate content
	logrus.Infof("Generate pick result content")
	content, err := NewResultComment(tp.PickResultTemplate, result)
	if err != nil {
		logrus.Errorf("Generate pick result content failed: %s", err)
		return nil, err
	}

	// submit pick result to merge request
	_, err = s.provider.Comment().Create(ctx, &tp.CreateCommentOption{
		Repo:           task.Repo,
		MergeRequestID: task.MergeRequestID,
		Body:           content,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) PerformPick(ctx context.Context, opt *CherryPickOptions) error {
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

	// Check conflict
	err = s.provider.Commit().CheckConflict(ctx, &tp.CheckConflictOption{
		Repo:     opt.Repo,
		Commit:   opt.SHA,
		Target:   opt.Target,
		Mode:     GetCheckConflictMode(s.provider.ProviderID()),
		Pr:       opt.Pr,
		RepoPath: opt.RepoPath,
	})
	if err != nil {
		logrus.Errorf("failed to check conflict: %+v", err)
		return err
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
	summaryComment, err := NewSummaryComment(tp.CherryPickTaskSummaryTemplate, targets)
	if err != nil {
		logrus.Errorf("NewSummaryComment failed: %+v", err)
		return err
	}

	// Check if the comment is existed
	// if exists, regen summary
	_, comment, err := s.FindCommentWithTask(ctx, task, tp.CherryPickSummaryFlag)
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
		// check if pick result is exist, if existed, skip
		comments, result, err := s.FindCommentWithTask(ctx, task, tp.CherryPickResultFlag)
		if result != nil {
			logrus.Warnf("pick result is exist %s.", result)
			return nil
		}
		if err != nil {
			logrus.Warnf("get pick result err: %s", err.Error())
			return err
		}

		// check if summary comment is exist, if not exist, skip
		comment := FindSummaryWithFlag(comments, tp.CherryPickSummaryFlag)
		if comment == nil {
			logrus.Warnf("not found pick summary [%s]", comment)
			return nil
		}
		// summary comment is exist, perform pick
		_, err = s.PerformPickToBranches(ctx, task, comment)
		if err != nil {
			logrus.Errorf("perform pick err: %s", err)
		}
	}
	return err
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
