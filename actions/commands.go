/*******************************************************************************
 * Copyright (c) 2019 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

package actions

import (
	"os"
	"log"

	"github.com/eclipse/codewind-installer/errors"

	"github.com/urfave/cli"
)

var tempFilePath = "installer-docker-compose.yaml"

const versionNum = "0.2.0"
const healthEndpoint = "/api/v1/environment"

//Commands for the installer
func Commands() {
	app := cli.NewApp()
	app.Name = "Codewind Installer"
	app.Version = versionNum
	app.Usage = "Start, Stop and Remove Codewind"

	// create commands
	app.Commands = []cli.Command{

		{
			Name:  "project",
			Usage: "Manage Codewind projects",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Aliases: []string{""},
					Usage: "create a project on disk",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "url, u",
							Usage: "URL of project to download",
						},
					},
					Action: func(c *cli.Context) error {
						if c.NumFlags() != 0 {
							DownloadTemplate(c)
						}
						ValidateProject(c)
						return nil
					},
				},
				{
					Name:  "validate",
					Aliases: []string{""},
					Usage: "validate an existing project on disk",
					Action: func(c *cli.Context) error {
						ValidateProject(c)
						return nil
					},
				},
				{
					Name:  "bind",
					Aliases: []string{""},
					Usage: "bind a project to codewind for building and running",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "the name of the project",
						},
						cli.StringFlag{
							Name:  "language, l",
							Usage: "the project language",
						},
						cli.StringFlag{
							Name:  "type, t",
							Usage: "the type of the project",
						},
					},
					Action: func(c *cli.Context) error {
						if c.NArg() == 0 {
							log.Fatal("path to source not set")
						}
						if c.NArg() > 1 {
							log.Fatal("too many arguments")
						}
						BindProject(c.Args().Get(0), c.String("name"), c.String("language"), c.String("type"))
						return nil
					},
				},
				{
					Name:  "sync",
					Aliases: []string{""},
					Usage: "synchronize a project to codewind for building",
					Action: func(c *cli.Context) error {
						// BindProject(c)
						return nil
					},
				},
			},
		},

		{
			Name:    "install",
			Aliases: []string{"in"},
			Usage:   "Pull pfe, performance & initialize images from dockerhub",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "latest",
					Usage: "dockerhub image tag",
				},
				cli.BoolFlag{
					Name:  "json, j",
					Usage: "specify terminal output",
				},
			},
			Action: func(c *cli.Context) error {
				InstallCommand(c)
				return nil
			},
		},

		{
			Name:  "start",
			Usage: "Start the Codewind containers",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "latest",
					Usage: "dockerhub image tag",
				},
				cli.BoolFlag{
					Name:  "debug, d",
					Usage: "add debug output",
				},
			},
			Action: func(c *cli.Context) error {
				StartCommand(c, tempFilePath, healthEndpoint)
				return nil
			},
		},

		{
			Name:  "status",
			Usage: "Print the installation status of Codewind",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json, j",
					Usage: "specify terminal output",
				},
			},
			Action: func(c *cli.Context) error {
				StatusCommand(c)
				return nil
			},
		},

		{
			Name:  "stop",
			Usage: "Stop the running Codewind containers",
			Action: func(c *cli.Context) error {
				StopCommand()
				return nil
			},
		},

		{
			Name:  "stop-all",
			Usage: "Stop all of the Codewind and project containers",
			Action: func(c *cli.Context) error {
				StopAllCommand()
				return nil
			},
		},

		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Usage: "dockerhub image tag",
				},
			},
			Usage: "Remove Codewind/Project docker images and the codewind network",
			Action: func(c *cli.Context) error {
				RemoveCommand(c)
				return nil
			},
		},

		{
			Name:    "templates",
			Usage:   "Manage project templates",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Aliases: []string{"ls"},
					Usage: "list available templates",
					Action: func(c *cli.Context) error {
						ListTemplates()
						return nil
					},
				},
				{
					Name:  "styles",
					Usage: "list available template styles",
					Action: func(c *cli.Context) error {
						ListTemplateStyles()
						return nil
					},
				},
				{
					Name:  "repos",
					Usage: "list available template repos",
					Action: func(c *cli.Context) error {
						ListTemplateRepos()
						return nil
					},
				},
			},
		},
	}

	// Start application
	err := app.Run(os.Args)
	errors.CheckErr(err, 300, "")
}
