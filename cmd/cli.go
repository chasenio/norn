package main

import (
	"fmt"
	"github.com/kentio/norn/cmd/pick"
	"github.com/kentio/norn/pkg/logger"
	"os"
)

var (
	BuildTime   = ""
	BuildNumber = ""
	GitCommit   = ""
	Version     = "0.0.1"
)

func main() {
	logger.SetLogger() // set logger format
	if err := pick.NewApp(&pick.CliInfo{
		BuildTime:   BuildTime,
		BuildNumber: BuildNumber,
		GitCommit:   GitCommit,
		Version:     Version,
	}).Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
