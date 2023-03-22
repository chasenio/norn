package main

import (
	"fmt"
	"github.com/kentio/norn/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	BuildTime   = ""
	BuildNumber = ""
	GitCommit   = ""
	Version     = "0.0.1"
)

func newApp() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "version: %s\n"+"Git Commit: %s\n"+"Build Time: %s\n"+"Build %s\n",
			c.App.Version, GitCommit, BuildTime, BuildNumber)
	}
	return &cli.App{
		Name:    "Norns",
		Version: Version,
		Usage:   "Norns is a CLI tool for cherry-picking commits from one branch to another",
		Commands: []*cli.Command{
			cmd.NewPickCommand(),
		},
		Before: func(context *cli.Context) error {
			debug := os.Getenv("NORN_DEBUG")
			if debug != "" {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
	}
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
