// Copyright (C) 2018 Miquel Sabaté Solà <mikisabate@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"

	"github.com/mssola/openhub/lib"

	"gopkg.in/urfave/cli.v1"
)

func fetchCredentials(ctx *cli.Context) lib.Credentials {
	return lib.Credentials{
		Server:   ctx.String("server"),
		User:     ctx.String("user"),
		Password: ctx.String("password"),
		Token:    ctx.String("token"),
	}
}

func run(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return fmt.Errorf("Exactly one argument is required, but %v was given", len(ctx.Args()))
	}

	cfg, err := lib.ParseConfiguration(
		ctx.Args().First(),
		fetchCredentials(ctx),
		lib.Options{SingleShot: ctx.Bool("single-shot")},
	)
	if err != nil {
		return err
	}
	err = lib.Sync(cfg)
	return err
}

func main() {
	app := cli.NewApp()
	app.Name = "openhub"
	app.Usage = "Glue service between OBS and DockerHub"
	app.UsageText = "openhub"
	app.HideHelp = true
	app.Version = versionString()

	app.CommandNotFound = func(context *cli.Context, cmd string) {
		fmt.Printf("Incorrect usage: there are no commands in openhub.\n\n")
		cli.ShowAppHelp(context)
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "server, s",
			Usage:  "The location of the Open Build Service server",
			Value:  "https://api.opensuse.org",
			EnvVar: "OPENHUB_OBS_SERVER",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "The password for the Open Build Service",
			EnvVar: "OPENHUB_OBS_PASSWORD",
		},
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "The authentication token provided from DockerHub",
			EnvVar: "OPENHUB_DOCKER_TOKEN",
		},
		cli.StringFlag{
			Name:   "user, u",
			Usage:  "The user to be used for the Open Build Service",
			EnvVar: "OPENHUB_OBS_USER",
		},
		cli.BoolFlag{
			Name:   "single-shot",
			Usage:  "Only run the execution cycle once",
			EnvVar: "OPENHUB_SINGLE_SHOT",
		},
	}

	app.Action = run
	app.RunAndExitOnError()
}
