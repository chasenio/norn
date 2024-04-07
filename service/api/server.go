package api

import (
	"context"
	"errors"
	"github.com/kentio/norn/internal/service"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Api struct {
	router *Router
	config *service.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func NewApi(router *Router, config *service.Config) *Api {
	return &Api{
		config: config,
		router: router,
	}

}

// Start starts the server
// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-with-context/server.go
func (a *Api) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.ctx = ctx
	a.cancel = cancel

	go func() {
		if err := a.router.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("listen: %s\n", err)
		}
	}()
}
