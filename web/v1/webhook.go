package webhook

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

func Hello(c *gin.Context) {
	// log request header
	logrus.Infof("request header: %v", c.Request.Header)
	hook := &github.Webhook{}

	// get webhook payload
	payload, err := hook.Parse(c.Request, github.IssuesEvent, github.PullRequestEvent, github.RepositoryEvent,
		github.GitHubAppAuthorizationEvent, github.InstallationRepositoriesEvent,
		github.InstallationEvent, // App Install event, type: github.InstallationPayload; action: created, suspend, unsuspend, deleted
		github.IntegrationInstallationEvent,
		github.IntegrationInstallationRepositoriesEvent)
	if err != nil {
		logrus.Errorf("error: %v", err)
	}
	// log payload and data type
	logrus.Infof("payload: %+v type: %v", payload, reflect.TypeOf(payload).String())

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})

}
