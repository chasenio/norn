package pick

import (
	"context"
	"fmt"
	"github.com/kentio/norn/internal"
	"github.com/kentio/norn/pkg/common"
	"github.com/kentio/norn/pkg/pick"
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

func NewApp() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer,
			"version: %s\n"+
				"Git Commit: %s\n"+
				"Build Time: %s\n"+
				"Build %s\n",
			c.App.Version, GitCommit, BuildTime, BuildNumber)
	}
	return &cli.App{
		Name:    "Norns",
		Version: Version,
		Usage:   "Norns is a CLI tool for cherry-picking commits from one ref to another",
		Commands: []*cli.Command{
			NewPickCommand(),
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

func NewPickCommand() *cli.Command {
	return &cli.Command{
		Name:  "pick",
		Usage: "pick commits from one branch to another",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "path",
				Usage:    "RepoPath to the git repo",
				Aliases:  []string{"p"},
				Required: false,
				Value:    ".cherry-pick-path.yml",
			},
			&cli.StringFlag{
				Name:    "vendor",
				Usage:   "Git vendor, such as gh(github)",
				Value:   "gh",
				Aliases: []string{"v"},
			},
			&cli.StringFlag{
				Name:     "repo",
				Usage:    "Git repo, such as kentio/norn",
				Aliases:  []string{"r"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "token",
				Usage:    "Personal access token",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "sha",
				Usage:    "Commit sha",
				Aliases:  []string{"s"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "for",
				Usage:    "PerformPick commits for a specific branch",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "merge-request-id",
				Usage:    "The merge request id",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "is-summary",
				Usage: "Add Cherry-pick summary to the merge request",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "repo-path",
				Usage: "RepoPath to the git repo",
				Value: ".",
			},
		},
		Action: func(c *cli.Context) error {
			logrus.Debugf("Start picking commits")
			ctx := context.Background()
			profile, err := internal.NewProfile(c.String("path"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			vendor, token, mrId := c.String("vendor"), c.String("token"), c.String("merge-request-id")
			logrus.Debugf("Vendor: %s, Token: %s Merge Request ID: %s", vendor, token, mrId)

			if vendor == "" || token == "" {
				return cli.Exit("Vendor or token is empty", 1)
			}

			provider, err := common.NewProvider(ctx, vendor, token)
			if err != nil {
				return cli.Exit("Unknown provider", 1)
			}

			repo, from := c.String("repo"), c.String("for")
			logrus.Debugf("Repo: %s, From: %s", repo, from)

			if repo == "" {
				return cli.Exit("Repo is empty", 1)
			}

			sha, isSummary := c.String("sha"), c.Bool("is-summary")
			logrus.Debugf("SHA: %s, IsSummary: %t", sha, isSummary)

			p := pick.NewPickService(provider, profile.Branches)

			pickOpt := &pick.Task{
				Repo:           repo,
				Branches:       profile.Branches,
				Form:           from,
				SHA:            &sha,
				MergeRequestID: mrId,
				IsSummary:      isSummary,
				RepoPath:       c.String("repo-path"),
			}

			/*
				TODO Command 和 Backend 之间的场景不太一样
				 	Command是具体指令，而Backend需要Context
			*/

			err = p.ProcessPick(ctx, pickOpt)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return cli.Exit("Success!", 0)
		},
	}
}
