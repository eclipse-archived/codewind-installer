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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/moby/moby/client"
	"github.ibm.com/codewind-installer/utils"

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
	app.Usage = "Start and Stop Codewind"

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
				if err != nil {
					log.Fatal(err)
				}
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

				ctx := context.Background()
				cli, err := client.NewEnvClient()
				if err != nil {
					log.Fatal(err)
				}

				containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Only stopping Codewind containers")
				for _, container := range containers {

					if strings.Contains(container.Image, "codewind") {
						fmt.Println("Stopping container ", container.Names, "... ")

						// Stop the running container
						if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
							log.Fatal(err)
						}

						// Remove the container so it isnt lingering in the background
						if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
							log.Fatal(err)
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

				ctx := context.Background()
				cli, err := client.NewEnvClient()
				if err != nil {
					log.Fatal(err)
				}

				containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Stopping Codewind and application containers")
				for _, container := range containers {

					if strings.Contains(container.Image, "codewind") || strings.Contains(container.Image, "mc-") {
						fmt.Println("Stopping container ", container.Names, "... ")

						// Stop the running container
						if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
							log.Fatal(err)
						}

						// Remove the container so it isnt lingering in the background
						if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
							log.Fatal(err)
						}
					}

				}

				return nil
			},
		},

		{
			Name:  "remove",
			Usage: "Remove Codewind docker images, Application docker images and the microclimate network",
			Action: func(c *cli.Context) error {

				ctx := context.Background()
				cli, err := client.NewEnvClient()
				if err != nil {
					log.Fatal(err)
				}

				images, err := cli.ImageList(ctx, types.ImageListOptions{})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Removing Codewind docker images..")
				for _, image := range images {

					imageRepo := strings.Join(image.RepoDigests, " ")

					if strings.Contains(imageRepo, "sys-mcs-docker-local.") || strings.Contains(imageRepo, "ibmcom/codewind") {
						fmt.Println("Deleting Image ", image.ID[:16], "... ")
						if _, err := cli.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{Force: true}); err != nil {
							log.Fatal(err)
						}
					}
				}

				networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
				if err != nil {
					log.Fatal(err)
				}

				for _, network := range networks {

					if strings.Contains(network.Name, "microclimate") {
						fmt.Print("Removing docker network: ", network.Name, "... ")

						// Remove the network
						if err := cli.NetworkRemove(ctx, network.ID); err != nil {
							fmt.Println("Cannot remove " + network.Name + ". Use 'stop-all' flag to ensure all containers have been terminated")
							log.Fatal(err)
						}
					}
				}

				return nil
			},
		},
	}

	// Start application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
