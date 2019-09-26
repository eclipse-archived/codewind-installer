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
	"log"

	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// StartCommand to start the codewind conainers
func StartCommand(c *cli.Context, tempFilePath string, healthEndpoint string) {
	tag := c.String("tag")

	if !utils.CheckImageStatus() {
		log.Fatal("Error: Cannot find Codewind images, try running install to pull them")
	}
	if !utils.CheckImageTags(tag) {
		log.Fatal(fmt.Sprintf("Cannot find Codewind images with tag %s, try running install with this tag", tag))
	}

	status := utils.CheckContainerStatus()
	if status {
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
}
