package main

import (
	"fmt"
	"os"

	"github.com/guessi/eks-managed-node-groups/pkg/constants"
	"github.com/guessi/eks-managed-node-groups/pkg/ui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    constants.AppName,
		Usage:   constants.AppUsage,
		Version: constants.VersionString,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Value:   "us-east-1",
				Usage:   "Region for the clusters",
			},
		},
		Action: func(c *cli.Context) error {
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
				Action: func(cCtx *cli.Context) error {
					ui.ShowVersion()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		return
	}
}
