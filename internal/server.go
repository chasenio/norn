package internal

import (
	"context"
	"github.com/kentio/norn/internal/service"
	"github.com/kentio/norn/service/api"
	"github.com/kentio/norn/service/task"
	"go.uber.org/fx"
)

type Server struct {
	api  *api.Api
	task *task.Service

	config *service.Config
}

var Model = fx.Options(
	api.Model,

	fx.Provide(service.NewConfig),
	fx.Provide(task.NewService),
	fx.Provide(NewServer),

	fx.Invoke(StartAppHook),
)

func NewServer(config *service.Config, api *api.Api, task *task.Service) *Server {
	return &Server{
		config: config,
		api:    api,
		task:   task,
	}
}

func StartAppHook(ctx context.Context, app *Server) error {
	return app.Start(ctx)
}

func (s *Server) Start(ctx context.Context) error {
	s.api.Start(ctx)

	s.task.Start()

	return nil
}

func NewApp(ctx context.Context) *fx.App {
	app := fx.New(
		Model,
		fx.Provide(
			func() context.Context {
				return ctx
			},
		),
	)
	return app
}
