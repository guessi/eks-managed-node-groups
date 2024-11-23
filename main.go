package main

import (
	"log"
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
		Action: func(*cli.Context) error {
			if err := ui.Entry(); err != nil {
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
		log.Fatal(err)
	}
}
