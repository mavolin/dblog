// Command dblog provides a simple CLI tool to generate a wrapper type that
// logs all calls to a repository to one or multiple loggers.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/mavolin/dblog/generator"
	"github.com/mavolin/dblog/internal/meta"
	"github.com/mavolin/dblog/logger"
	"github.com/mavolin/dblog/logger/sentry"
)

type app struct {
	cli *cli.App

	sentryLogger *sentry.Logger
}

func main() {
	ver := meta.Version
	if meta.Commit != meta.UnknownCommit {
		ver += " (" + meta.Commit + ")"
	}

	var app app

	app.cli = &cli.App{
		Name:        "dblog",
		Usage:       "Generate a wrapper type that logs all calls to a repository interface.",
		Version:     ver,
		Description: "Utility to generate a logger for repositories.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "o",
				Value: "dblog/wrapper.go",
				Usage: "The path to the output file.",
			},
			&cli.StringFlag{
				Name:  "t",
				Value: "Wrapper",
				Usage: "The name of the generated type.",
			},

			&cli.BoolFlag{
				Name:  "sentry",
				Value: false,
				Usage: "Enable sentry logging.",
			},
		},
		ArgsUsage: "interfaceName",
		Action:    app.run,
	}

	app.sentryLogger = sentry.NewLogger(app.cli)

	if err := app.cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (app *app) run(cctx *cli.Context) error {
	var ls []logger.Logger
	if cctx.Bool("sentry") {
		ls = append(ls, app.sentryLogger)
	}

	g, err := generator.New(".", cctx.Args().Get(0), ls...)
	if err != nil {
		return err
	}

	if err = g.Generate(cctx.String("o"), cctx.String("t")); err != nil {
		return err
	}

	goFmt := exec.Command("go", "fmt", cctx.String("o"))
	if err = goFmt.Run(); err != nil {
		fmt.Println("failed to format generated file:", err)
	}

	return nil
}
