package api

import "go.uber.org/fx"

var Model = fx.Options(
	fx.Provide(NewRouter),
	fx.Provide(NewApi),
)
