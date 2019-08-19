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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "r",
					Usage: "repository url",
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
			Usage:   "Remove Codewind/Project docker images and the codewind network",
			Action: func(c *cli.Context) error {
				RemoveCommand()
				return nil
			},
		},
	}

	// Start application
	err := app.Run(os.Args)
	errors.CheckErr(err, 300, "")
}
