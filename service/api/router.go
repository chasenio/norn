package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kentio/norn/internal/service"
	webhook "github.com/kentio/norn/service/api/v1"
	"github.com/kentio/norn/service/task"
	"go.uber.org/fx"
	"net/http"
)

type Router struct {
	*gin.Engine
	httpServer *http.Server
}

func NewRouter(lc fx.Lifecycle, config *service.Config, tk *task.Service) *Router {
	r := gin.New()

	v1 := r.Group("/v1")
	hook := v1.Group("/webhook")

	{
		hook.POST("/github", webhook.GitHubHandler(config, tk))
	}

	// if version is not v1, return default webhooks
	r.POST("/webhook/github", webhook.GitHubHandler(config, tk))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HTTPPort),
		Handler: r.Handler(),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return httpServer.Shutdown(ctx)
		},
	})

	return &Router{
		Engine:     r,
		httpServer: httpServer,
	}
}
