package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strconv"
)

type PullRequestService struct {
	client *gh.Client
}

type PullRequest struct {
	id          int
	title       string
	description string
	state       tp.MergeRequestState
}

func (s *PullRequest) MergeId() string {
	return s.title
}

func (s *PullRequest) Title() string {
	return s.title
}

func (s *PullRequest) Description() string {
	return s.description
}

func (s *PullRequest) State() tp.MergeRequestState {
	return s.state
}

func NewPullRequestService(client *gh.Client) *PullRequestService {
	return &PullRequestService{
		client: client,
	}
}

func (s *PullRequestService) Get(ctx context.Context, opt *tp.GetMergeRequestOption) (tp.MergeRequest, error) {
	if opt == nil {
		return nil, tp.ErrInvalidOptions
	}
	logrus.Debugf("Get Pull Request Opt: %+v", *opt)
	repoOpt, err := parseRepo(opt.Repo)
	if err != nil {
		return nil, err
	}

	// pull request to int
	mergeId, err := strconv.Atoi(opt.MergeID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert merge id to int: %v", err)
	}
	pr, response, err := s.client.PullRequests.Get(ctx, repoOpt.Owner, repoOpt.Repo, mergeId)
	logrus.Debugf("Get Pull Request Response: %+v", *response)
	if err != nil {
		logrus.Debugf("Get PR Error: %+v", err)
		return nil, err
	}
	return newPullRequest(pr), nil
}

func newPullRequest(pr *gh.PullRequest) (mr *PullRequest) {
	return &PullRequest{
		id:          pr.GetNumber(),
		title:       pr.GetTitle(),
		description: pr.GetBody(),
		state:       mr.getStateFromGithubPullRequest(pr),
	}
}

func (s *PullRequest) getStateFromGithubPullRequest(pr *gh.PullRequest) tp.MergeRequestState {
	return getStateFromGitHubPullRequestState(pr.GetState())
}

func getStateFromGitHubPullRequestState(state string) tp.MergeRequestState {
	switch state {
	case "open":
		return tp.MergeRequestStateOpen
	case "closed":
		return tp.MergeRequestStateClosed
	case "merged":
		return tp.MergeRequestStateMerged
	default:
		return tp.MergeRequestStateUnknown
	}
}
