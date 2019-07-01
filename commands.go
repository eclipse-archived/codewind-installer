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
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"

	"github.com/docker/docker/api/types"

	"github.com/urfave/cli"
)

var tempFilePath = "installer-docker-compose.yaml"
var debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

const versionNum = "0.2.0"
const healthEndpoint = "http://localhost:9090/api/v1/environment"

func commands() {
	app := cli.NewApp()
	app.Name = "Codewind Installer"
	app.Version = versionNum
	app.Usage = "Start, Stop and Remove Codewind"

	// create commands
	app.Commands = []cli.Command{

		{
			Name:    "install-dev",
			Aliases: []string{"in-dev"},
			Usage:   "Pull pfe, performance & intialize images from artifactory",
			Action: func(c *cli.Context) error {

				authConfig := types.AuthConfig{
					Username: os.Getenv("USER"),
					Password: os.Getenv("PASS"),
				}
				encodedJSON, err := json.Marshal(authConfig)
				errors.CheckErr(err, 106, "")

				authStr := base64.URLEncoding.EncodeToString(encodedJSON)

				imageArr := [3]string{"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe-amd64",
					"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance-amd64",
					"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize-amd64"}

				targetArr := [3]string{"codewind-pfe-amd64:latest",
					"codewind-performance-amd64:latest",
					"codewind-initialize-amd64:latest"}

				for i := 0; i < len(imageArr); i++ {
					utils.PullImage(imageArr[i], authStr)
					utils.TagImage(imageArr[i], targetArr[i])
				}

				fmt.Println("Image Tagging Successful")

				return nil
			},
		},

		{
			Name:    "install",
			Aliases: []string{"in"},
			Usage:   "Pull pfe, performance & intialize images from dockerhub",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "latest",
					Usage: "dockerhub image tag",
				},
			},
			Action: func(c *cli.Context) error {
				tag := c.String("tag")

				imageArr := [3]string{"docker.io/ibmcom/codewind-pfe-amd64:",
					"docker.io/ibmcom/codewind-performance-amd64:",
					"docker.io/ibmcom/codewind-initialize-amd64:"}

				targetArr := [3]string{"codewind-pfe-amd64:",
					"codewind-performance-amd64:",
					"codewind-initialize-amd64:"}

				for i := 0; i < len(imageArr); i++ {
					utils.PullImage(imageArr[i]+tag, "")
					utils.TagImage(imageArr[i]+tag, targetArr[i]+tag)
				}

				fmt.Println("Image Tagging Successful")

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
			},
			Action: func(c *cli.Context) error {
				status := utils.CheckContainerStatus()

				if status {
					fmt.Println("Codewind is already running!")
				} else {
					tag := c.String("tag")
					utils.CreateTempFile(tempFilePath)
					utils.WriteToComposeFile(tempFilePath)
					utils.DockerCompose(tag)
					utils.DeleteTempFile(tempFilePath) // Remove installer-docker-compose.yaml
					utils.PingHealth(healthEndpoint)
				}
				return nil
			},
		},

		{
			Name:  "status",
			Usage: "Print the installation status of Codewind",
			Action: func(c *cli.Context) error {

				if utils.CheckContainerStatus() {
					fmt.Println("Codewind is installed and running")
					os.Exit(2)
				}

				if utils.CheckImageStatus() {
					fmt.Println("Codewind is installed but not running")
					os.Exit(1)
				} else {
					fmt.Println("Codewind is not installed")
					os.Exit(0)
				}
				return nil
			},
		},

		{
			Name:  "stop",
			Usage: "Stop the running Codewind containers",
			Action: func(c *cli.Context) error {
				containerArr := [2]string{}
				containerArr[0] = "codewind-pfe"
				containerArr[1] = "codewind-performance"

				containers := utils.GetContainerList()

				fmt.Println("Only stopping Codewind containers. To stop project containers, please use 'stop-all'")

				for _, container := range containers {
					for _, key := range containerArr {
						if strings.HasPrefix(container.Image, key) {
							fmt.Println("Stopping container ", container.Names, "... ")
							utils.StopContainer(container)
						}
					}
				}
				return nil
			},
		},

		{
			Name:  "stop-all",
			Usage: "Stop all of the Codewind and project containers",
			Action: func(c *cli.Context) error {
				containerArr := [3]string{}
				containerArr[0] = "codewind-pfe"
				containerArr[1] = "codewind-performance"
				containerArr[2] = "cw-"

				containers := utils.GetContainerList()

				fmt.Println("Stopping Codewind and Project containers")
				for _, container := range containers {
					for _, key := range containerArr {
						if strings.HasPrefix(container.Image, key) {
							fmt.Println("Stopping container ", container.Names, "... ")
							utils.StopContainer(container)
						}
					}
				}
				return nil
			},
		},

		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "Remove Codewind/Project docker images and the codewind network",
			Action: func(c *cli.Context) error {
				imageArr := [7]string{}
				imageArr[0] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe"
				imageArr[1] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance"
				imageArr[2] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize"
				imageArr[3] = "ibmcom/codewind-pfe"
				imageArr[4] = "ibmcom/codewind-performance"
				imageArr[5] = "ibmcom/codewind-initialize"
				imageArr[6] = "cw-"
				networkName := "codewind"

				images := utils.GetImageList()

				fmt.Println("Removing Codewind docker images..")

				for _, image := range images {
					imageRepo := strings.Join(image.RepoDigests, " ")
					imageTags := strings.Join(image.RepoTags, " ")
					for _, key := range imageArr {
						if strings.HasPrefix(imageRepo, key) || strings.HasPrefix(imageTags, key) {
							if len(image.RepoTags) > 0 {
								fmt.Println("Deleting Image ", image.RepoTags[0], "... ")
							} else {
								fmt.Println("Deleting Image ", image.ID, "... ")
							}
							utils.RemoveImage(image.ID)
						}
					}
				}

				networks := utils.GetNetworkList()

				for _, network := range networks {
					if strings.Contains(network.Name, networkName) {
						fmt.Print("Removing docker network: ", network.Name, "... ")
						utils.RemoveNetwork(network)
					}
				}
				return nil
			},
		},
	}

	// Start application
	err := app.Run(os.Args)
	errors.CheckErr(err, 300, "")
}
