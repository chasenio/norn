package api

import (
	"context"
	"errors"
	"github.com/kentio/norn/internal/service"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"net/http"
	"time"
)

type Api struct {
	r   *Router
	cfg *service.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func NewApi(lc fx.Lifecycle, router *Router, config *service.Config) *Api {
	api := &Api{
		cfg: config,
		r:   router,
	}

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	api.ctx = ctx
	api.cancel = cancel

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			defer api.cancel()
			logrus.Info("api shutdown.")
			return router.httpServer.Shutdown(ctx)
		},
	})

	return api

}

// Start starts the server
// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-with-context/server.go
func (a *Api) Start() {

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling behind
	go func() {
		if err := a.r.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("listen: %s\n", err)
		}
	}()
}
