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
	"runtime"

	"github.com/eclipse/codewind-installer/pkg/appconstants"
	"github.com/eclipse/codewind-installer/pkg/errors"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Get home directory
func getHomeDir() string {
	homeDir := ""
	const GOOS string = runtime.GOOS
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}

var homeDir = getHomeDir()
var dockerComposeFile = homeDir + "/.codewind/docker-compose.yaml"

const healthEndpoint = "/api/v1/environment"

//Commands for the controller
func Commands() {
	app := cli.NewApp()
	app.Name = "cwctl"
	app.Version = appconstants.VersionNum
	app.Usage = "Start, Stop and Remove Codewind"

	// Global Flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "disable certificate checking",
		},
		cli.BoolFlag{
			Name:  "json, j",
			Usage: "output as JSON",
		},
		cli.StringFlag{
			Name:  "loglevel",
			Value: "info",
			Usage: "log level {trace,debug,info,fatal,error}",
		},
	}

	// create commands
	app.Commands = []cli.Command{

		{
			Name:  "project",
			Usage: "Manage Codewind projects",

			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{""},
					Usage:   "create a project on disk",

					Flags: []cli.Flag{
						cli.StringFlag{Name: "url, u", Usage: "URL of project to download"},
						cli.StringFlag{Name: "type, t", Usage: "Known type and subtype of project (`type:subtype`). Ignored when URL is given"},
						cli.StringFlag{Name: "conid", Value: "local", Usage: "The connection id for the project", Required: false},
					},
					Action: func(c *cli.Context) error {
						if c.String("u") != "" {
							ProjectCreate(c)
						}
						ProjectValidate(c)
						return nil
					},
				},
				{
					Name:    "bind",
					Aliases: []string{""},
					Usage:   "bind a project to codewind for building and running",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "name, n", Usage: "The name of the project", Required: true},
						cli.StringFlag{Name: "language, l", Usage: "The project language", Required: true},
						cli.StringFlag{Name: "type, t", Usage: "The type of the project", Required: true},
						cli.StringFlag{Name: "path, p", Usage: "The path to the project", Required: true},
						cli.StringFlag{Name: "conid", Value: "local", Usage: "The connection id for the project", Required: false},
					},
					Action: func(c *cli.Context) error {
						ProjectBind(c)
						return nil
					},
				},
				{
					Name:    "remove",
					Aliases: []string{""},
					Usage:   "remove a project from codewind",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "id, i", Usage: "the project id", Required: true},
						cli.BoolFlag{Name: "delete, d", Usage: "delete local project files"},
					},
					Action: func(c *cli.Context) error {
						ProjectRemove(c)
						return nil
					},
				},
				{
					Name:    "sync",
					Aliases: []string{""},
					Usage:   "synchronize a project to codewind for building and running",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "path, p", Usage: "the path to the project", Required: true},
						cli.StringFlag{Name: "id, i", Usage: "the project id", Required: true},
						cli.StringFlag{Name: "time, t", Usage: "time of the last sync for the given project", Required: true},
					},
					Action: func(c *cli.Context) error {
						ProjectSync(c)
						return nil
					},
				},
				{
					Name:    "connection",
					Aliases: []string{"con"},
					Usage:   "Manage project connections",
					Subcommands: []cli.Command{
						{
							Name:    "set",
							Aliases: []string{"s"},
							Usage:   "Set connectionID for a project",
							Flags: []cli.Flag{
								cli.StringFlag{Name: "id,i", Usage: "Project ID", Required: true},
								cli.StringFlag{Name: "conid", Usage: "Connection ID", Required: true},
							},
							Action: func(c *cli.Context) error {
								ProjectSetConnection(c)
								return nil
							},
						},
						{
							Name:    "get",
							Aliases: []string{"g"},
							Usage:   "Get connectionID for a project",
							Flags: []cli.Flag{
								cli.StringFlag{Name: "id,i", Usage: "Project ID", Required: true},
							},
							Action: func(c *cli.Context) error {
								ProjectGetConnection(c)
								return nil
							},
						}, {
							Name:    "remove",
							Aliases: []string{"r"},
							Usage:   "Remove connection from a project",

							Flags: []cli.Flag{
								cli.StringFlag{Name: "id,i", Usage: "Project ID", Required: true},
							},
							Action: func(c *cli.Context) error {
								ProjectRemoveConnection(c)
								return nil
							},
						},
					},
				},
			},
		},

		{
			Name:    "install",
			Aliases: []string{"in"},
			Usage:   "Pull pfe and performance images from dockerhub",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "latest",
					Usage: "dockerhub image tag",
				},
				cli.BoolFlag{
					Name:  "json, j",
					Usage: "ouput as JSON",
				},
			},
			Action: func(c *cli.Context) error {
				InstallCommand(c)
				return nil
			},

			Subcommands: []cli.Command{
				{
					Name:    "remote",
					Aliases: []string{"r"},
					Usage:   "Install a remote deployment of Codewind",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "namespace,n", Usage: "Kubernetes namespace", Required: true},
						cli.StringFlag{Name: "session,ses", Usage: "Codewind session secret", Required: false},
						cli.StringFlag{Name: "ingress,i", Usage: "Ingress Domain eg: 10.22.33.44.nip.io", Required: false},
						cli.StringFlag{Name: "kadminuser,au", Usage: "Keycloak admin user", Required: false},
						cli.StringFlag{Name: "kadminpass,ap", Usage: "Keycloak admin password", Required: false},
						cli.StringFlag{Name: "kdevuser,du", Usage: "Keycloak developer username to add", Required: false},
						cli.StringFlag{Name: "kdevpass,dp", Usage: "Keycloak developer username initial password", Required: false},
						cli.StringFlag{Name: "krealm,r", Usage: "Keycloak realm to setup", Required: false},
						cli.StringFlag{Name: "kclient,c", Usage: "Keycloak client to setup", Required: false},
						cli.IntFlag{Name: "pvcsize,p", Usage: "Codewind PVC size (integer between 1 and 999 Gigabytes)", Required: false, Value: 1},
						cli.StringFlag{Name: "kurl", Usage: "Don't deploy a new Keycloak pod, use this existing one instead", Required: false},
						cli.BoolFlag{Name: "konly", Usage: "Install a deployment of Keycloak only", Required: false},
					},
					Action: func(c *cli.Context) error {
						DoRemoteInstall(c)
						return nil
					},
				},
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
				StartCommand(c, dockerComposeFile, healthEndpoint)
				return nil
			},
		},

		{
			Name:  "status",
			Usage: "Print the installation status of Codewind",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json, j",
					Usage: "output as JSON",
				},
				cli.StringFlag{
					Name:  "conid",
					Usage: "ConnectionID to check",
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
				StopCommand(c, dockerComposeFile)
				return nil
			},
		},

		{
			Name:  "stop-all",
			Usage: "Stop all of the Codewind and project containers",
			Action: func(c *cli.Context) error {
				StopAllCommand(c, dockerComposeFile)
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
				RemoveCommand(c, dockerComposeFile)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:    "remote",
					Aliases: []string{"r"},
					Usage:   "Removes and deletes a Codewind remote deployment from Kubernetes",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "namespace,n", Usage: "Kubernetes namespace", Required: true},
						cli.StringFlag{Name: "workspace,w", Usage: "Codewind workspace ID", Required: true},
					},
					Action: func(c *cli.Context) error {
						DoRemoteRemove(c)
						return nil
					},
				},
				{
					Name:    "keycloak",
					Aliases: []string{"k"},
					Usage:   "Removes and deletes a Keycloak deployment from Kubernetes",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "namespace,n", Usage: "Kubernetes namespace", Required: true},
						cli.StringFlag{Name: "workspace,w", Usage: "Keycloak workspace ID", Required: true},
					},
					Action: func(c *cli.Context) error {
						DoRemoteKeycloakRemove(c)
						return nil
					},
				},
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
							Usage: "Filter by project style",
						},
						cli.BoolFlag{
							Name:  "showEnabledOnly",
							Usage: "Filter by whether a template is enabled or not",
						},
						cli.StringFlag{
							Name:     "conid",
							Value:    "local",
							Usage:    "Connection ID",
							Required: false,
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
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:     "conid",
							Value:    "local",
							Usage:    "Connection ID",
							Required: false,
						},
					},
					Action: func(c *cli.Context) error {
						ListTemplateStyles(c)
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
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:     "conid",
									Value:    "local",
									Usage:    "Connection ID",
									Required: false,
								},
							},
							Action: func(c *cli.Context) error {
								ListTemplateRepos(c)
								return nil
							},
						},
						{
							Name:  "add",
							Usage: "Add a template repo",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "url",
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
								cli.StringFlag{
									Name:     "conid",
									Value:    "local",
									Usage:    "Connection ID",
									Required: false,
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
									Name:  "url",
									Usage: "URL of the template repo",
								},
								cli.StringFlag{
									Name:     "conid",
									Value:    "local",
									Usage:    "Connection ID",
									Required: false,
								},
							},
							Action: func(c *cli.Context) error {
								DeleteTemplateRepo(c)
								return nil
							},
						},
						{
							Name:  "enable",
							Usage: "Enable template repos with the given URLs",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:     "conid",
									Value:    "local",
									Usage:    "Connection ID",
									Required: false,
								},
							},
							Action: func(c *cli.Context) error {
								EnableTemplateRepos(c)
								return nil
							},
						},
						{
							Name:  "disable",
							Usage: "Disable template repos with the given URLs",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:     "conid",
									Value:    "local",
									Usage:    "Connection ID",
									Required: false,
								},
							},
							Action: func(c *cli.Context) error {
								DisableTemplateRepos(c)
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
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Application realm", Required: false},
						cli.StringFlag{Name: "username,u", Usage: "Account Username", Required: true},
						cli.StringFlag{Name: "password,p", Usage: "Account Password", Required: false},
						cli.StringFlag{Name: "client,c", Usage: "Client", Required: false},
						cli.StringFlag{Name: "conid", Usage: "Connection ID", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityTokenGet(c)
						return nil
					},
				}, {
					Name:    "refresh",
					Aliases: []string{"r"},
					Usage:   "Obtain an access token using a refresh_token",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "conid", Usage: "Connection ID", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityTokenRefresh(c)
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
						cli.StringFlag{Name: "conid", Usage: "Connection ID (see the connections cmd)", Required: true},
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
						cli.StringFlag{Name: "conid,d", Usage: "Keycloak login ID", Required: true},
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
						cli.StringFlag{Name: "newrealm,r", Usage: "New realm name", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
					},
					Action: func(c *cli.Context) error {
						SecurityCreateRealm(c)
						return nil
					},
				},
			},
		},
		{
			Name:    "secrole",
			Aliases: []string{"sl"},
			Usage:   "Manage access role configuration",
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new role in an existing realm (requires admin_token)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: true},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "role,l", Usage: "New Role name", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityCreateRole(c)
						return nil
					},
				},
			},
		},
		{
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
						cli.StringFlag{Name: "realm,r", Usage: "Realm where client should be created", Required: true},
						cli.StringFlag{Name: "newclient,c", Usage: "New client ID to create", Required: true},
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
						cli.StringFlag{Name: "clientid,c", Usage: "Client ID to retrieve", Required: true},
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
						cli.StringFlag{Name: "clientid,c", Usage: "Client ID to retrieve", Required: true},
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
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
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
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin access_token", Required: false},
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
				}, {
					Name:  "addrole",
					Usage: "Adds an existing role to an existing user (requires admin_token)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "host", Usage: "URL or ingress to Keycloak service", Required: false},
						cli.StringFlag{Name: "realm,r", Usage: "Realm name", Required: true},
						cli.StringFlag{Name: "accesstoken,t", Usage: "Admin Access Token", Required: false},
						cli.StringFlag{Name: "name,n", Usage: "Existing user account name to process", Required: true},
						cli.StringFlag{Name: "role,rl", Usage: "Existing user role name to add to the user account", Required: true},
					},
					Action: func(c *cli.Context) error {
						SecurityUserAddRole(c)
						return nil
					},
				},
			},
		},
		//  Connection maintenance //
		{
			Name:    "connections",
			Aliases: []string{"con"},
			Usage:   "Manage connections list",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "Add a new connection to the configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "label", Usage: "A displayable name", Required: true},
						cli.StringFlag{Name: "url", Usage: "The ingress URL of Codewind gatekeeper", Required: true},
						cli.StringFlag{Name: "username,u", Usage: "Username", Required: true},
					},
					Action: func(c *cli.Context) error {
						ConnectionAddToList(c)
						return nil
					},
				},
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "Update an existing connection",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "conid", Usage: "Connection ID to update", Required: true},
						cli.StringFlag{Name: "label", Usage: "A displayable name", Required: true},
						cli.StringFlag{Name: "url", Usage: "The ingress URL of Codewind gatekeeper", Required: true},
						cli.StringFlag{Name: "username,u", Usage: "Username", Required: true},
					},
					Action: func(c *cli.Context) error {
						ConnectionUpdate(c)
						return nil
					},
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "Get a connection config by id",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "conid", Usage: "Connection ID to retrieve", Required: true},
					},
					Action: func(c *cli.Context) error {
						ConnectionGetByID(c)
						return nil
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm"},
					Usage:   "Remove a connection from the configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "conid", Usage: "The reference ID of the connection to be removed", Required: true},
					},
					Action: func(c *cli.Context) error {
						ConnectionRemoveFromList(c)
						return nil
					},
				},
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "List known connections",
					Action: func(c *cli.Context) error {
						ConnectionListAll(c)
						return nil
					},
				},
				{
					Name:  "reset",
					Usage: "Resets the connections list",
					Action: func(c *cli.Context) error {
						ConnectionResetList(c)
						return nil
					},
				},
			},
		},
		{
			Name:    "upgrade",
			Aliases: []string{"up"},
			Usage:   "Upgrade projects",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "workspace, ws", Usage: "the workspace directory to upgrade, location of projects", Required: true},
			},
			Action: func(c *cli.Context) error {
				UpgradeProjects(c)
				return nil
			},
		},
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Get versions of remotely deployed Codewind containers",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "conid", Usage: "The connection ID", Required: true},
			},
			Action: func(c *cli.Context) error {
				GetVersions(c)
				return nil
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		// Handle Global flag to disable certificate checking
		if c.GlobalBool("insecure") {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		// Handle Global log level flag

		switch loglevel := c.GlobalString("loglevel"); {
		case loglevel == "trace":
			logr.SetLevel(logr.TraceLevel)
			break
		case loglevel == "debug":
			logr.SetLevel(logr.DebugLevel)
			break
		case loglevel == "fatal":
			logr.SetLevel(logr.FatalLevel)
			break
		case loglevel == "error":
			logr.SetLevel(logr.ErrorLevel)
			break
		case loglevel == "warn":
			logr.SetLevel(logr.WarnLevel)
			break
		default:
			logr.SetLevel(logr.InfoLevel)
		}

		return nil
	}

	// Start application
	err := app.Run(os.Args)
	errors.CheckErr(err, 300, "")
}
