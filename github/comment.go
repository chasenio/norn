package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"strconv"
)

type CommentService struct {
	client *gh.Client
}

type Comment struct {
	commentId string
	body      string
}

func NewCommentService(client *gh.Client) *CommentService {
	return &CommentService{
		client: client,
	}
}

// Create Comment creates a new comment on the given merge request.
func (s *CommentService) Create(ctx context.Context, opt *tp.CreateCommentOption) (tp.Comment, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	logrus.Debugf("Add Comment Opt: %+v", *opt)
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}

	mergeId, err := strconv.Atoi(opt.MergeRequestID)
	prComment, response, err := s.client.Issues.CreateComment(ctx,
		repoOpt.Owner,
		repoOpt.Repo, mergeId,
		&gh.IssueComment{
			Body: gh.String(opt.Body),
			User: &gh.User{
				Name:  gh.String("github-action[bot]"),
				Email: gh.String("github-actions[bot]@users.noreply.github.com"),
			},
		})
	logrus.Debugf("Add Comment Response: %+v", response)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %v", err)
	}
	if response.StatusCode != 201 {
		return nil, fmt.Errorf("failed to add comment: %v", response.Status)
	}
	logrus.Debugf("Add Comment : %+v", *prComment)
	return newIssueComment(prComment), nil
}

// Find Comment finds comments on the given merge request.
func (s *CommentService) Find(ctx context.Context, opt *tp.FindCommentOption) ([]tp.Comment, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	logrus.Debugf("Find Comment Opt: %+v", *opt)
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	// pull request to int
	mrId, err := strconv.Atoi(opt.MergeRequestID)
	if err != nil {
		logrus.Debugf("failed to convert merge id to int: %v", err)
		return nil, fmt.Errorf("failed to convert merge id to int: %v", err)
	}

	logrus.Debugf("Merge Reqeust ID: %v", mrId)
	// find comment
	comments, response, err := s.client.Issues.ListComments(ctx, repoOpt.Owner, repoOpt.Repo, mrId, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments request: %v", err)
	}
	logrus.Debugf("Find Comment Response: %d", response.StatusCode)

	return lo.Map(comments, func(c *gh.IssueComment, _ int) tp.Comment {
		return newIssueComment(c)
	}), nil
}

// Update Comment updates a comment on the given merge request.
func (s *CommentService) Update(ctx context.Context, opt *tp.UpdateCommentOption) (tp.Comment, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Update Comment Opt: %+v", *opt)
	_commentID, err := strconv.Atoi(opt.CommentID)
	commentID := int64(_commentID)
	if err != nil {
		logrus.Debugf("failed to convert comment id to int: %v", err)
		return nil, err
	}
	comment, response, err := s.client.Issues.EditComment(ctx, repoOpt.Owner, repoOpt.Repo, commentID, &gh.IssueComment{
		ID:   &commentID,
		Body: gh.String(opt.Body),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list comments request: %v", err)
	}
	logrus.Debugf("Update Comment %s Response: %d", opt.CommentID, response.StatusCode)
	return newIssueComment(comment), nil
}

func newIssueComment(comment *gh.IssueComment) *Comment {
	return &Comment{
		commentId: strconv.FormatInt(comment.GetID(), 10),
		body:      comment.GetBody(),
	}
}

func newPRComment(comment *gh.PullRequestComment) *Comment {
	return &Comment{
		commentId: strconv.FormatInt(comment.GetID(), 10),
		body:      comment.GetBody(),
	}
}

func (c *Comment) CommentID() string {
	return c.commentId
}

func (c *Comment) Body() string {
	return c.body
}
