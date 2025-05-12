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
		Name:    constants.AppName,
		Usage:   constants.AppUsage,
		Version: constants.GitVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Value:   "us-east-1",
				Usage:   "Region for the clusters",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			region := c.String("region")
			if err := ui.Entry(region); err != nil {
				return err
			}
			return nil
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
		fmt.Println(err.Error())
		return
	}
}
