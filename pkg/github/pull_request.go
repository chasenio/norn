package github

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v50/github"
	"github.com/kentio/norn/pkg/types"
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
	state       types.MergeRequestState
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

func (s *PullRequest) State() types.MergeRequestState {
	return s.state
}

func NewPullRequestService(client *gh.Client) *PullRequestService {
	return &PullRequestService{
		client: client,
	}
}

func (s *PullRequestService) Get(ctx context.Context, opt *types.GetMergeRequestOption) (types.MergeRequest, error) {
	if opt == nil {
		return nil, types.ErrInvalidOptions
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
		return nil, fmt.Errorf("failed to get pull request: %v", err)
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

func (s *PullRequest) getStateFromGithubPullRequest(pr *gh.PullRequest) types.MergeRequestState {
	return getStateFromGitHubPullRequestState(pr.GetState())
}

func getStateFromGitHubPullRequestState(state string) types.MergeRequestState {
	switch state {
	case "open":
		return types.MergeRequestStateOpen
	case "closed":
		return types.MergeRequestStateClosed
	case "merged":
		return types.MergeRequestStateMerged
	default:
		return types.MergeRequestStateUnknown
	}
}
