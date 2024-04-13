package main

import (
	"fmt"
	"github.com/kentio/norn/cmd/pick"
	"github.com/kentio/norn/pkg/logger"
	"os"
)

func main() {
	logger.SetLogger() // set logger format
	if err := pick.NewApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
