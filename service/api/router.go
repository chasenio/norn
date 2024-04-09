package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kentio/norn/internal/service"
	webhook "github.com/kentio/norn/service/api/v1"
	"github.com/kentio/norn/service/task"
	"net/http"
)

type Router struct {
	*gin.Engine
	httpServer *http.Server
}

func NewRouter(config *service.Config, tk *task.Service, ctx context.Context) *Router {
	r := gin.Default()

	v1 := r.Group("/v1")
	hook := v1.Group("/webhook")

	{
		hook.POST("/github", webhook.GitHubHandler(ctx, config, tk))
	}

	// if version is not v1, return default webhooks
	r.POST("/webhook/github", webhook.GitHubHandler(ctx, config, tk))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HTTPPort),
		Handler: r.Handler(),
	}

	return &Router{
		Engine:     r,
		httpServer: httpServer,
	}
}
