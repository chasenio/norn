package main

import (
	"context"
	"github.com/kentio/norn/internal"
)

func main() {
	app := internal.NewApp(context.Background())
	app.Run()
}
