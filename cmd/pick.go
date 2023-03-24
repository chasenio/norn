package cmd

import (
	"context"
	"github.com/kentio/norn/github"
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
		},
		Action: func(c *cli.Context) error {
			ctx := context.Background()
			client := NewClient(ctx, c.String("vendor"), c.String("token"))
			if client == nil {
				return cli.Exit("Unknown vendor", 1)
			}
			err := client.Pick(ctx, c.String("repo"), &github.PickOption{
				SHA:    c.String("sha"),
				Branch: c.String("target"),
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return nil
		},
	}
}

// NewClient returns a new client for the given vendor.
func NewClient(ctx context.Context, vendor string, token string) *github.PickClient {
	switch vendor {
	case "gh":
		return github.NewClient(ctx, token)
	default:
		return nil
	}
}
