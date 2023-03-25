package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	"github.com/kentio/norn/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"strconv"
)

type CommentService struct {
	client *gh.Client
}

type Comment struct {
	commentId int64
	body      string
}

func NewCommentService(client *gh.Client) *CommentService {
	return &CommentService{
		client: client,
	}
}

func (s *CommentService) Create(ctx context.Context, opt *types.CreateCommentOption) (types.Comment, error) {
	if opt == nil {
		return nil, types.ErrInvalidOptions
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
		})
	logrus.Debugf("Add Comment Response: %+v", *response)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %v", err)
	}
	logrus.Debugf("Add Comment : %+v", *prComment)
	return newIssueComment(prComment), nil
}

func (s *CommentService) Find(ctx context.Context, opt *types.FindCommentOption) ([]types.Comment, error) {
	if opt == nil {
		return nil, types.ErrInvalidOptions
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

	return lo.Map(comments, func(c *gh.IssueComment, _ int) types.Comment {
		return newIssueComment(c)
	}), nil
}

func newIssueComment(comment *gh.IssueComment) *Comment {
	return &Comment{
		commentId: comment.GetID(),
		body:      comment.GetBody(),
	}
}

func newPRComment(comment *gh.PullRequestComment) *Comment {
	return &Comment{
		commentId: comment.GetID(),
		body:      comment.GetBody(),
	}
}

func (c *Comment) CommentID() any {
	return c.commentId
}

func (c *Comment) Body() string {
	return c.body
}
