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

const versionNum = "0.3.1"
const healthEndpoint = "/api/v1/environment"

//Commands for the installer
func Commands() {
	app := cli.NewApp()
	app.Name = "codewind-installer"
	app.Version = versionNum
	app.Usage = "Start, Stop and Remove Codewind"

	// Global Flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "disable certificate checking",
		},
	}

	// create commands
	app.Commands = []cli.Command{

		{
			Name:  "project",
			Usage: "Manage Codewind projects",
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
			Name:  "templates",
			Usage: "Manage project templates",
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage: "List available templates",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "projectStyle",
							Value: "Codewind",
							Usage: "Filter by project style",
						},
						cli.StringFlag{
							Name:  "showEnabledOnly",
							Value: "false",
							Usage: "Filter by whether a template is enabled or not",
						},
					},
					Action: func(c *cli.Context) error {
						ListTemplates(c)
						return nil
					},
				},
				{
					Name:  "styles",
					Usage: "List available template styles",
					Action: func(c *cli.Context) error {
						ListTemplateStyles()
						return nil
					},
				},
				{
					Name:  "repos",
					Usage: "List available template repos",
					Action: func(c *cli.Context) error {
						ListTemplateRepos()
						return nil
					},
				},
			},
		},

		//  Deployment maintenance //
		{
			Name:    "deployments",
			Aliases: []string{"dep"},
			Usage:   "Maintain local deployments list",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "Add a new deployment to the list",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "name", Usage: "A reference name", Required: true},
						cli.StringFlag{Name: "label", Usage: "A displayable name", Required: false},
						cli.StringFlag{Name: "url", Usage: "The ingress url of the PFE instance", Required: true},
					},
					Action: func(c *cli.Context) error {
						AddDeploymentToList(c)
						return nil
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm"},
					Usage:   "Remove a deployment from the list",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "name", Usage: "The reference name of the deployment to be removed", Required: true},
					},
					Action: func(c *cli.Context) error {
						if c.NumFlags() != 1 {
						} else {
							RemoveDeploymentFromList(c)
						}
						return nil
					},
				},
				{
					Name:    "target",
					Aliases: []string{"t"},
					Usage:   "Show/Change the current target deployment",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "name", Usage: "The deployment name of a new target"},
					},
					Action: func(c *cli.Context) error {
						if c.NumFlags() == 0 {
							ListTargetDeployment()
						} else {
							SetTargetDeployment(c)
						}
						return nil
					},
				},
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "List known deployments",
					Action: func(c *cli.Context) error {
						ListDeployments()
						return nil
					},
				},
				{
					Name:  "reset",
					Usage: "Resets the deployments list",
					Action: func(c *cli.Context) error {
						ResetDeploymentsFile()
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
