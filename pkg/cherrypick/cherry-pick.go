package cherrypick

import (
	"context"
	"errors"
	"github.com/chasenio/norn/internal"
	tp "github.com/chasenio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

// Service provides cherry-pick operations with configurable templates.
// It manages the creation of summary and result comments for cherry-pick tasks.
type Service struct {
	provider tp.Provider
}

type Options struct {
	SHA      string // commit sha
	Repo     string
	Target   string // target branch
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
	From           string   // from branch
	SHA            *string
	MergeRequestID string
	IsSummary      bool // generate summary comment
	PickMode       Mode
	RepoPath       string
	BranchPrefix   string // branch prefix for temp branch
}

type Status string

const (
	SucceedStatus = "Succeed"
	FailedStatus  = "Failed"
	PendingStatus = "Pending"
	SkipStatus    = "Skip"
)

type TaskResult struct {
	Status Status
	Branch string
	Reason string
}

// NewService creates a new cherry-pick service with the given provider.
// It uses default templates. Use NewServiceWithTemplates for custom templates.
//
// Parameters:
//   - provider: The git provider implementation (e.g., GitHub, GitLab)
//
// Returns:
//   - *Service: A configured cherry-pick service instance with default templates
func NewService(provider tp.Provider) *Service {
	return &Service{
		provider: provider,
	}
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

func (s *Service) GetSelected(ctx context.Context, task *Task, template *MessageTemplate) ([]string, error) {
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: task.MergeRequestID, Repo: task.Repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return nil, err
	}

	comment := FindSummaryWithFlag(comments, template.UniqueID)
	if comment == nil {
		logrus.Warnf("not found pick summary [%s]", comment)
		return nil, nil
	}

	// get selected branches
	selected := parseSelectedBranches(comment.Body())
	return selected, nil
}

// PerformPickToBranches PerformPick commits from one branches to another
func (s *Service) PerformPickToBranches(ctx context.Context, task *Task, selected []string, template *MessageTemplate) (result []*TaskResult, err error) {

	logrus.Debugf("Start to pick ...")

	logrus.Infof("Selected branches: %s", selected)

	// PerformPick commits from one branch to another
	for _, branch := range selected {
		var status Status
		if branch == task.From {
			logrus.Debugf("Skip form branch: %s", branch)
			continue // skip the branch, and pick commits from the next branch
		}

		// if select branch not in defined branches, skip
		if !internal.StringInSlice(branch, task.Branches) {
			logrus.Debugf("Skip pick: %s, not in defined %s", branch, task.Branches)
			continue
		}

		logrus.Debugf("cherry-pick [%s] to [%s]", *task.SHA, branch)
		// cherry-pick commit
		err := s.provider.Cherry().CherryPick(ctx, task.Repo, &tp.Option{
			Branch: branch,
			SHA:    *task.SHA,
			Prefix: task.BranchPrefix,
		})
		if err != nil {
			logrus.Warnf("Cherry-pick %s to %s failed: %s", *task.SHA, branch, err)
			status = FailedStatus
			if errors.Is(err, tp.NotFound) {
				status = SkipStatus
			}
			var e *tp.ProviderError
			if !errors.As(err, &e) { // 如果不是 ProviderError 需要对信息做处理
				// format error message, 如果能够通过空格分割 1 次，取后面的部分
				message := strings.Split(err.Error(), " ")
				if len(message) > 1 {
					logrus.Warnf("source error: %s", err)
					err = errors.New(strings.Join(message[1:], " "))
				}
			}
			result = append(result, &TaskResult{Status: status, Branch: branch, Reason: err.Error()})
		} else {
			status = SucceedStatus
			result = append(result, &TaskResult{Status: status, Branch: branch})
		}
		logrus.Infof("cherry-pick %s to %s %s", *task.SHA, branch, status)
	}
	logrus.Infof("Picke Result %v", result)

	if len(result) == 0 {
		logrus.Warnf("No branch to pick")
		return nil, nil
	}

	// generate content
	logrus.Infof("Generate pick result content")
	content, err := NewResultComment(template.Text, result)
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
	logrus.Infof("Submit Result Comment: \n%s", content)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateSummary submit pick summary comment
func (s *Service) CreateSummary(ctx context.Context, task *Task, template *Templates) error {
	if template == nil {
		template = DefaultTemplates()
	}
	// generate branch list of comment body
	targets := generateTargetBranches(task)
	logrus.Debugf("Summary branches: %+v", targets)
	if len(targets) == 0 {
		logrus.Debug("No target branches, delete summary comment if exists")
		s.DeleteSummaryWithFlag(ctx, task)
		return nil
	}

	// generate comment body
	summaryComment, err := NewSummaryComment(template.Summary.Text, targets)
	if err != nil {
		logrus.Errorf("NewSummaryComment failed: %+v", err)
		return err
	}

	// Check if the comment is existed
	// if exists, regen summary
	_, comment, err := s.FindCommentWithTask(ctx, task, template.Summary.UniqueID)
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

func (s *Service) CherryPick(ctx context.Context, task *Task, template *Templates) error {
	var err error
	if template == nil {
		template = DefaultTemplates()
	}
	if task.IsSummary {
		err = s.CreateSummary(ctx, task, template)
		if err != nil {
			logrus.Errorf("create summary err: %s", err)
		}
	} else {
		// get selected branches
		selected, err := s.GetSelected(ctx, task, template.Summary)

		if err != nil {
			logrus.Errorf("get selected branches err: %s", err)
			return err
		}

		if len(selected) == 0 {
			logrus.Warnf("no selected branches")
			return nil
		}

		// summary comment is exist, perform pick
		_, err = s.PerformPickToBranches(ctx, task, selected, template.CherryPickResult)
		if err != nil {
			logrus.Errorf("perform pick err: %s", err)
		}
	}
	return err
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

// DeleteSummaryWithFlag delete summary comment
// - find summary comment with flag
// - delete summary comment
func (s *Service) DeleteSummaryWithFlag(ctx context.Context, task *Task) {
	comments, err := s.provider.Comment().Find(ctx, &tp.FindCommentOption{MergeRequestID: task.MergeRequestID, Repo: task.Repo})
	if err != nil {
		logrus.Warnf("Get merge request comments failed: %s", err)
		return
	}
	comment := FindSummaryWithFlag(comments, tp.CherryPickSummaryFlag)
	if comment != nil {
		err = s.provider.Comment().Delete(ctx, &tp.DeleteCommentOption{
			CommentID: comment.CommentID(),
			Repo:      task.Repo,
		})
		if err != nil {
			logrus.Warnf("Delete summary comment failed: %s", err)
		}
	}
}
