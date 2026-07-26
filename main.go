package main

import (
	"context"
	"fmt"
	"os"

	"github.com/guessi/eks-managed-node-groups/pkg/constants"
	"github.com/guessi/eks-managed-node-groups/pkg/ui"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:                  constants.AppName,
		Usage:                 constants.AppUsage,
		Version:               constants.GitVersion,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Value:   "us-east-1",
				Usage:   "Region for the clusters",
			},
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "AWS shared config profile to use",
			},
			&cli.StringFlag{
				Name:    "cluster",
				Aliases: []string{"c"},
				Usage:   "Cluster name (skip interactive cluster selection)",
			},
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "Node group type, one of: managed, self-managed (skip interactive type selection)",
			},
			&cli.StringFlag{
				Name:    "nodegroup",
				Aliases: []string{"n"},
				Usage:   "Node group name (skip interactive nodegroup selection)",
			},
			&cli.Int32Flag{
				Name:        "desired",
				Usage:       "Desired size of the node group (requires --min and --max)",
				DefaultText: "interactive",
			},
			&cli.Int32Flag{
				Name:        "min",
				Usage:       "Min size of the node group (requires --desired and --max)",
				DefaultText: "interactive",
			},
			&cli.Int32Flag{
				Name:        "max",
				Usage:       "Max size of the node group (requires --desired and --min)",
				DefaultText: "interactive",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview the change without applying it",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Apply without the confirmation prompt",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "text",
				Usage:   "Output format for dry-run and apply results: text or json",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			opts := ui.Options{
				Region:        c.String("region"),
				Profile:       c.String("profile"),
				ClusterName:   c.String("cluster"),
				NodeGroupType: c.String("type"),
				NodegroupName: c.String("nodegroup"),
				DryRun:        c.Bool("dry-run"),
				Yes:           c.Bool("yes"),
				Output:        c.String("output"),
			}

			if c.IsSet("desired") {
				v := c.Int32("desired")
				opts.DesiredSize = &v
			}
			if c.IsSet("min") {
				v := c.Int32("min")
				opts.MinSize = &v
			}
			if c.IsSet("max") {
				v := c.Int32("max")
				opts.MaxSize = &v
			}

			if err := opts.Validate(); err != nil {
				return err
			}

			return ui.Entry(opts)
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print version number",
				Action: func(context.Context, *cli.Command) error {
					ui.ShowVersion()
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
