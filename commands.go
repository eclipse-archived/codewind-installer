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

	//myFlags := []cli.Flag{} - No need to use seperate global flags yet

	//app.Flags = myFlags

	// create commands
	app.Commands = []cli.Command{

		{
			Name:  "install-dev",
			Usage: "Pull pfe, performance & intialize images from artifactory",
			Action: func(c *cli.Context) error {

				authConfig := types.AuthConfig{
					Username: os.Getenv("USER"),
					Password: os.Getenv("PASS"),
				}
				encodedJSON, err := json.Marshal(authConfig)
				errors.CheckErr(err, 106, "")

				authStr := base64.URLEncoding.EncodeToString(encodedJSON)

				codewindImage := "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe-amd64"
				codewindImageTarget := "codewind-pfe-amd64:latest"

				performanceImage := "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance-amd64"
				performanceImageTarget := "codewind-performance-amd64:latest"

				initializeImage := "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize-amd64"
				initializeImageTarget := "codewind-initialize-amd64:latest"

				utils.PullImage(codewindImage, authStr)
				utils.PullImage(performanceImage, authStr)
				utils.PullImage(initializeImage, authStr)

				utils.TagImage(codewindImage, codewindImageTarget)
				utils.TagImage(performanceImage, performanceImageTarget)
				utils.TagImage(initializeImage, initializeImageTarget)

				fmt.Println("Image Tagging Successful")

				return nil
			},
		},

		{
			Name:  "install",
			Usage: "Pull pfe, performance & intialize images from dockerhub",
			Action: func(c *cli.Context) error {

				codewindImage := "docker.io/ibmcom/codewind-pfe-amd64"
				codewindImageTarget := "codewind-pfe-amd64:latest"

				performanceImage := "docker.io/ibmcom/codewind-performance-amd64"
				performanceImageTarget := "codewind-performance-amd64:latest"

				initializeImage := "docker.io/ibmcom/codewind-initialize-amd64"
				initializeImageTarget := "codewind-initialize-amd64:latest"

				utils.PullImage(codewindImage, "")
				utils.PullImage(performanceImage, "")
				utils.PullImage(initializeImage, "")

				utils.TagImage(codewindImage, codewindImageTarget)
				utils.TagImage(performanceImage, performanceImageTarget)
				utils.TagImage(initializeImage, initializeImageTarget)

				fmt.Println("Image Tagging Successful")

				return nil
			},
		},

		{
			Name:  "start",
			Usage: "Start the Codewind containers",
			Action: func(c *cli.Context) error {
				status := utils.CheckContainerStatus()

				if status {
					fmt.Println("Codewind is already running!")
				} else {
					utils.CreateTempFile(tempFilePath)
					utils.WriteToComposeFile(tempFilePath)
					utils.DockerCompose()
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
				containerArr[2] = "mc-"

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
			Name:  "remove",
			Usage: "Remove Codewind/Project docker images and the microclimate network",
			Action: func(c *cli.Context) error {
				imageArr := [7]string{}
				imageArr[0] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe"
				imageArr[1] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance"
				imageArr[2] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize"
				imageArr[3] = "ibmcom/codewind-pfe"
				imageArr[4] = "ibmcom/codewind-performance"
				imageArr[5] = "ibmcom/codewind-initialize"
				imageArr[6] = "mc-"
				networkName := "microclimate"

				images := utils.GetImageList()

				fmt.Println("Removing Codewind docker images..")

				for _, image := range images {
					imageRepo := strings.Join(image.RepoDigests, " ")
					imageTags := strings.Join(image.RepoTags, " ")
					for _, key := range imageArr {
						if strings.HasPrefix(imageRepo, key) || strings.HasPrefix(imageTags, key) {
							fmt.Println("Deleting Image ", image.RepoTags[0], "... ")
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
