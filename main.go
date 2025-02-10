package main

import (
	"fmt"
	"os"

	"github.com/memoio/xspace-server/cmd"
	"github.com/urfave/cli/v2"
)

// @title			Xspace-Server API
// @version		1.0
// @description	This is the API document for xspace server.
// @Host			xspace.docs.org
// @BasePath		/
func main() {
	local := make([]*cli.Command, 0, 1)
	local = append(local, cmd.XspaceServerCmd, cmd.VersionCmd)
	app := cli.App{
		Commands: local,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show application version",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("version") {
				fmt.Println(cmd.Version + "+" + cmd.BuildFlag)
			}
			return nil
		},
	}
	app.Setup()

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err) // nolint:errcheck
		os.Exit(1)
	}
}
