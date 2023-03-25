package cmd

import (
	"context"
	"github.com/kentio/norn/feature"
	"github.com/kentio/norn/github"
	"github.com/kentio/norn/types"
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
				Usage:    "Pick commits for a specific branch",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			logrus.Debugf("Start picking commits")
			ctx := context.Background()
			profile, err := NewProfile(c.String("path"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			vendor, token := c.String("vendor"), c.String("token")
			logrus.Debugf("Vendor: %s, Token: %s", vendor, token)

			if vendor == "" || token == "" {
				return cli.Exit("Vendor or token is empty", 1)
			}

			provider, err := NewProvider(ctx, vendor, token)
			if err != nil {
				return cli.Exit("Unknown provider", 1)
			}

			repo, from := c.String("repo"), c.String("for")
			logrus.Debugf("Repo: %s, From: %s", repo, from)

			if repo == "" || from == "" {
				return cli.Exit("Repo is empty", 1)
			}

			launchPick := false

			logrus.Debugf("Branchs: %s", profile.Branches)

			// Pick commits from one branch to another
			for _, branch := range profile.Branches {
				logrus.Debugf("Branch: %s", branch)
				if branch == from {
					logrus.Debugf("Launch pick: %s", branch)
					launchPick = true
					continue // skip the branch, and pick commits from the next branch
				}
				if !launchPick {
					logrus.Debugf("Skip pick: %s", branch)
					continue
				}

				logrus.Debugf("Picking %s to %s", c.String("sha"), branch)
				// Pick commits
				pickOption := &feature.PickOption{
					SHA:    c.String("sha"),
					Repo:   repo,
					Target: branch,
				}
				err := feature.Pick(ctx, provider, pickOption)
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}
				logrus.Infof("Picked %s to %s", pickOption.SHA, pickOption.Target)
			}
			return nil
		},
	}
}

// NewProvider NewClient returns a new client for the given vendor.
func NewProvider(ctx context.Context, vendor string, token string) (types.Provider, error) {
	logrus.Debugf("New provider: %s", vendor)

	switch vendor {
	case "gh":
		return github.NewProvider(ctx, token)
	default:
		return nil, types.ErrUnknownProvider
	}
}
