package web

import (
	"github.com/gin-gonic/gin"
	"github.com/kentio/norn/web/v1"
)

func NewApp() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	hook := v1.Group("/webhook")

	{
		hook.POST("/github", webhook.Hello)
	}

	// if version is not v1, return default webhooks
	r.POST("/webhook/github", webhook.Hello)

	return r

}
