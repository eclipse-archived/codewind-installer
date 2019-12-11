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
	"fmt"
	"os"

	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// StartCommand : start the codewind containers
func StartCommand(c *cli.Context, tempFilePath string, healthEndpoint string) {
	tag := c.String("tag")

	imagesAreInstalled, err := utils.CheckImageStatus()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if !imagesAreInstalled {
		fmt.Println("Cannot find Codewind images, try running install to pull them")
		os.Exit(0)
	}

	taggedImagesAreInstalled, err := utils.CheckImageTag(tag)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if !taggedImagesAreInstalled {
		fmt.Println("Cannot find Codewind images with given tag, try running install to pull them")
		os.Exit(0)
	}

	containersAreRunning, err := utils.CheckContainerStatus()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if containersAreRunning {
		fmt.Println("Codewind is already running!")
	} else {
		debug := c.Bool("debug")
		tag := c.String("tag")
		fmt.Println("Debug:", debug)

		// Stop all running project containers and remove codewind networks
		StopAllCommand()

		utils.CreateTempFile(tempFilePath)
		utils.WriteToComposeFile(tempFilePath, debug)
		utils.DockerCompose(tempFilePath, tag)
		utils.DeleteTempFile(tempFilePath) // Remove installer-docker-compose.yaml
		utils.PingHealth(healthEndpoint)
	}
	os.Exit(0)
}
