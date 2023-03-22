package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	BuildTime   = ""
	BuildNumber = ""
	GitCommit   = ""
	Version     = "1.0.0"
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
	}
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
