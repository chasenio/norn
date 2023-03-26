package cmd

import (
	"context"
	"github.com/kentio/norn/feature"
	"github.com/kentio/norn/global"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func NewPickCommand() *cli.Command {
	return &cli.Command{
		Name:  "pick",
		Usage: "pick commits from one branch to another",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "path",
				Usage:    "Path to the git repo",
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
				Usage:    "DoPick commits for a specific branch",
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
		},
		Action: func(c *cli.Context) error {
			logrus.Debugf("Start picking commits")
			ctx := context.Background()
			profile, err := NewProfile(c.String("path"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			vendor, token, mrId := c.String("vendor"), c.String("token"), c.String("merge-request-id")
			logrus.Debugf("Vendor: %s, Token: %s Merge Request ID: %s", vendor, token, mrId)

			if vendor == "" || token == "" {
				return cli.Exit("Vendor or token is empty", 1)
			}

			provider, err := global.NewProvider(ctx, vendor, token)
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

			pick := feature.NewPickFeature(provider, profile.Branches)

			pickOpt := &feature.PickToRefMROpt{
				Repo:           repo,
				Branches:       profile.Branches,
				Form:           from,
				SHA:            sha,
				MergeRequestID: mrId,
				IsSummaryTask:  isSummary,
			}

			/*
				TODO Command 和 Backend 之间的场景不太一样
				 	Command是具体指令，而Backend需要Context
			*/
			if isSummary {
				// Add cherry-pick summary to the merge request
				if err := pick.DoPickSummaryComment(ctx, pickOpt); err != nil {
					return cli.Exit(err.Error(), 1)
				}
				return cli.Exit("", 0)
			}

			if profile.Branches != nil {
				_, _, err := pick.DoPickToBranchesFromMergeRequest(ctx, pickOpt)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}
			}
			return nil
		},
	}
}
