package webhook

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/kentio/norn/feature"
	"github.com/kentio/norn/internal/common"
	"github.com/kentio/norn/internal/service"
	pgh "github.com/kentio/norn/pkg/github"
	"github.com/kentio/norn/pkg/types"
	"github.com/kentio/norn/service/task"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strconv"
)

const (
	GitHubHeader = "X-GitHub-Event"
)

func GitHubHandler(ctx context.Context, cfg *service.Config, tk *task.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// log request header
		logrus.Infof("request header: %v", c.Request.Header)
		hook, _ := github.New(github.Options.Secret(cfg.Github.Secret))
		event := github.Event(c.Request.Header.Get(GitHubHeader))

		// get webhook payload
		payload, err := hook.Parse(c.Request, github.IssuesEvent, github.PullRequestEvent, github.RepositoryEvent,
			github.GitHubAppAuthorizationEvent, github.InstallationRepositoriesEvent,
			github.InstallationEvent, // App Install event, type: github.InstallationPayload; action: created, suspend, unsuspend, deleted
			github.IntegrationInstallationEvent,
			github.IntegrationInstallationRepositoriesEvent)
		if err != nil && !errors.Is(err, github.ErrEventNotFound) || payload == nil {
			logrus.Errorf("parse payload error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "parse payload error.",
			})
			return
		}

		payloadType := reflect.TypeOf(payload).String()
		action := reflect.ValueOf(payload).FieldByName("Action").Interface().(string)
		logrus.Infof("payload type: %v action: %v", payloadType, action)
		if cfg.Dev {
			common.WebhookToFile(payload, action)
		}

		// Push Task
		if event == github.PullRequestEvent && IsEffectiveAction(action) {
			pl := payload.(github.PullRequestPayload)
			opt := NewPickOptFromEvent(pl, cfg.Branches)
			tk.Push(func() {
				provider, err := NewProvider(ctx, cfg, pl.Installation.ID)
				if err != nil {
					logrus.Errorf("New Provider Err: %s", err)
					return
				}
				srv := feature.NewPickService(provider, cfg.Branches)
				err = srv.DoWithOpt(ctx, opt)
				if err != nil {
					logrus.Errorf("do with opt err: %s", err)
				}
			})
			logrus.Infof("send task id: %s", opt.MergeRequestID)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success.",
		})

	}

}

// GitHub Action Types
const (
	Opened   string = "opened"
	Reopened string = "reopened"
	Edited   string = "edited"
	Closed   string = "closed"
	Merged   string = "merged"
)

func NewPickOptFromEvent(pl github.PullRequestPayload, branches []string) *feature.PickToRefMROpt {
	opt := &feature.PickToRefMROpt{
		Repo:           pl.Repository.FullName,
		Branches:       branches,
		Form:           pl.PullRequest.Base.Ref,
		SHA:            *pl.PullRequest.MergeCommitSha,
		MergeRequestID: strconv.FormatInt(pl.PullRequest.Number, 10),
	}
	switch pl.Action {
	case Opened, Reopened, Edited: // generate summary plan
		opt.IsSummaryTask = true
	case Merged:
		opt.IsSummaryTask = false
	}
	return opt
}

func NewProvider(ctx context.Context, cfg *service.Config, installId int64) (types.Provider, error) {

	appID, _ := strconv.Atoi(cfg.Github.AppID)
	cred := &pgh.Credential{
		AppID:          int64(appID),
		InstallationID: installId,
		PrivateKey:     []byte(cfg.Github.PrivateKey),
	}
	provider, err := pgh.NewProviderWithOpt(ctx, cred)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func IsEffectiveAction(action string) bool {
	switch action {
	case Opened, Reopened, Edited, Merged:
		return true
	default:
		return false
	}
}
