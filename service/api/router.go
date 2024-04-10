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

func NewRouter(cfg *service.Config, tk *task.Service, ctx context.Context) *Router {
	if !cfg.Dev {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	v1 := r.Group("/v1")
	hook := v1.Group("/webhook")

	{
		hook.POST("/github", webhook.GitHubHandler(ctx, cfg, tk))
	}

	// if version is not v1, return default webhooks
	r.POST("/webhook/github", webhook.GitHubHandler(ctx, cfg, tk))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: r.Handler(),
	}

	return &Router{
		Engine:     r,
		httpServer: httpServer,
	}
}
