package webhook

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/kentio/norn/internal/common"
	"github.com/kentio/norn/internal/service"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

func GitHubHandler(config *service.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// log request header
		logrus.Infof("request header: %v", c.Request.Header)
		hook, _ := github.New(github.Options.Secret(config.Github.Secret))

		// get webhook payload
		payload, err := hook.Parse(c.Request, github.IssuesEvent, github.PullRequestEvent, github.RepositoryEvent,
			github.GitHubAppAuthorizationEvent, github.InstallationRepositoriesEvent,
			github.InstallationEvent, // App Install event, type: github.InstallationPayload; action: created, suspend, unsuspend, deleted
			github.IntegrationInstallationEvent,
			github.IntegrationInstallationRepositoriesEvent)
		if err != nil {
			logrus.Errorf("parse payload error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "parse payload error.",
			})
			return
		}

		payloadType := reflect.TypeOf(payload).String()
		action := reflect.ValueOf(payload).FieldByName("Action").Interface().(string)
		logrus.Infof("payload type: %v action: %v", payloadType, action)

		common.WebhookToFile(payload, action)

		c.JSON(http.StatusOK, gin.H{
			"message": "success.",
		})

	}

}
