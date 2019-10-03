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
	"crypto/tls"
	"net/http"
	"os"

	"github.com/eclipse/codewind-installer/errors"

	"github.com/urfave/cli"
)

var tempFilePath = "codewind-docker-compose.yaml"

const versionNum = "x.x.dev"

const healthEndpoint = "/api/v1/environment"

//Commands for the controller
func Commands() {
	app := cli.NewApp()
	app.Name = "cwctl"
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
				cli.StringFlag{
					Name:  "type, t",
					Usage: "Known type and subtype of project (`type:subtype`). Ignored when URL is given",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("u") != "" {
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
					Usage:   "List available templates",
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
					Usage: "Manage template repos",
					Subcommands: []cli.Command{
						{
							Name:    "list",
							Aliases: []string{"ls"},
							Usage:   "List available template repos",
							Action: func(c *cli.Context) error {
								ListTemplateRepos()
								return nil
							},
						},
						{
							Name:  "add",
							Usage: "Add a template repo",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "URL",
									Usage: "URL of the template repo",
								},
								cli.StringFlag{
									Name:  "description",
									Value: "",
									Usage: "Description of the template repo",
								},
								cli.StringFlag{
									Name:  "name",
									Value: "",
									Usage: "Name of the template repo",
								},
							},
							Action: func(c *cli.Context) error {
								AddTemplateRepo(c)
								return nil
							},
						},
						{
							Name:    "remove",
							Aliases: []string{"rm"},
							Usage:   "Remove a template repo",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "URL",
									Usage: "URL of the template repo",
								},
							},
							Action: func(c *cli.Context) error {
								DeleteTemplateRepo(c)
								return nil
							},
						},
					},
				},
			},
		},

		//  Security //
		{
			Name:    "sectoken",
			Aliases: []string{"st"},
			Usage:   "Authenticate and obtain an access_token",
			Subcommands: []cli.Command{
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "Login and retrieve access_token",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: true},
						cli.StringFlag{Name: "realm,r", Usage: "Application realm", Required: true},
						cli.StringFlag{Name: "username,u", Usage: "Account Username", Required: true},
						cli.StringFlag{Name: "password,p", Usage: "Account Password", Required: true},
						cli.StringFlag{Name: "client,c", Usage: "Client", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityTokenGet(c)
						return nil
					},
				},
			},
		},
		{
			Name:    "seckeyring",
			Aliases: []string{"sk"},
			Usage:   "Manage the desktop keyring",
			Subcommands: []cli.Command{
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "Add new or update existing Codewind credentials in the keyring",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "depid,d", Usage: "Deployment ID (see the deployments cmd)", Required: true},
						cli.StringFlag{Name: "username,u", Usage: "Username", Required: true},
						cli.StringFlag{Name: "password,p", Usage: "New password", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityKeyUpdate(c)
						return nil
					},
				}, {
					Name:    "validate",
					Aliases: []string{"v"},
					Usage:   "Checks if Codewind credentials exist in the keyring",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "depid,d", Usage: "Keycloak login ID", Required: true},
						cli.StringFlag{Name: "username,u", Usage: "Username", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityKeyValidate(c)
						return nil
					},
				},
			},
		},
		{
			Name:    "secrealm",
			Aliases: []string{"sr"},
			Usage:   "Manage Realm configuration",
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new realm (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: true},
						cli.StringFlag{Name: "realm,r", Usage: "Existing realm name", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityCreateRealm(c)
						return nil
					},
				},
			},
		}, {
			Name:    "secclient",
			Aliases: []string{"sc"},
			Usage:   "Manage client access configuration",
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new client in a Keycloak realm (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: true},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "clientid,c", Usage: "New client ID to create", Required: true},
						cli.StringFlag{Name: "redirect,l", Usage: "Redirect URL", Required: false},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityClientCreate(c)
						return nil
					},
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "Get client id (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "clientid,c", Usage: "New client ID to create", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityClientGet(c)
						return nil
					},
				},
				{
					Name:    "secret",
					Aliases: []string{"s"},
					Usage:   "Get client secret (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "clientid,c", Usage: "Client id", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityClientGetSecret(c)
						return nil
					},
				},
			},
		},
		{
			Name:    "secuser",
			Aliases: []string{"su"},
			Usage:   "Manage keycloak user account",
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new user (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "admintoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
						cli.StringFlag{Name: "name,n", Usage: "Username to add", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityUserCreate(c)
						return nil
					},
				}, {
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "Get details of a user (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "admintoken,t", Usage: "Admin access_token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
						cli.StringFlag{Name: "name,n", Usage: "Username to retrieve", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityUserGet(c)
						return nil
					},
				}, {
					Name:    "setpw",
					Aliases: []string{"p"},
					Usage:   "Sets the password of an existing user (requires either admin_token or username/password)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin Access Token", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Admin Username", Required: false},
						cli.StringFlag{Name: "password,p", Usage: "Admin Password", Required: false},
						cli.StringFlag{Name: "name,n", Usage: "Existing user account name to process", Required: true},
						cli.StringFlag{Name: "newpw,w", Usage: "New password", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityUserSetPassword(c)
						return nil
					},
				},
			},
		},
		//  Deployment maintenance //
		{
			Name:    "deployments",
			Aliases: []string{"dep"},
			Usage:   "Manage deployments list",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "Add a new deployment to the configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "id", Usage: "A reference name", Required: true},
						cli.StringFlag{Name: "label", Usage: "A displayable name", Required: false},
						cli.StringFlag{Name: "url", Usage: "The ingress URL of the PFE instance", Required: true},
						cli.StringFlag{Name: "auth", Usage: "URL of Keycloak service eg: https://mykeycloak.test:8443", Required: false},
						cli.StringFlag{Name: "realm", Usage: "Security realm eg: codewind or che", Required: false},
						cli.StringFlag{Name: "clientid", Usage: "Security client_id to connect as eg: codewind_ctl or che-public", Required: false},
					},
					Action: func(c *cli.Context) error {
						DeploymentAddToList(c)
						return nil
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm"},
					Usage:   "Remove a deployment from the configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "id", Usage: "The reference ID of the deployment to be removed", Required: true},
					},
					Action: func(c *cli.Context) error {
						if c.NumFlags() != 1 {
						} else {
							DeploymentRemoveFromList(c)
						}
						return nil
					},
				},
				{
					Name:    "target",
					Aliases: []string{"t"},
					Usage:   "Show/Change the current target deployment",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "id", Usage: "The deployment id of the target to switch to"},
					},
					Action: func(c *cli.Context) error {
						if c.NumFlags() == 0 {
							DeploymentGetTarget()
						} else {
							DeploymentSetTarget(c)
						}
						return nil
					},
				},
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "List known deployments",
					Action: func(c *cli.Context) error {
						DeploymentListAll()
						return nil
					},
				},
				{
					Name:  "reset",
					Usage: "Resets the deployments list",
					Action: func(c *cli.Context) error {
						DeploymentResetList()
						return nil
					},
				},
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		// Handle Global flag to disable certificate checking
		if c.GlobalBool("insecure") {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		return nil
	}

	// Start application
	err := app.Run(os.Args)
	errors.CheckErr(err, 300, "")
}
