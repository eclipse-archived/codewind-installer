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

	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/urfave/cli"
)

// StartCommand : start the codewind containers
func StartCommand(c *cli.Context, dockerComposeFile string, healthEndpoint string) {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	status, err := docker.CheckContainerStatus(dockerClient)
	if err != nil {
		HandleDockerError(err)
		os.Exit(1)
	}

	if status {
		fmt.Println("Codewind is already running!")
	} else {
		tag := c.String("tag")
		debug := c.Bool("debug")
		loglevel := c.GlobalString("loglevel")
		fmt.Println("Debug:", debug)

		writeToComposeFileErr := docker.WriteToComposeFile(dockerComposeFile, debug)
		if writeToComposeFileErr != nil {
			HandleDockerError(writeToComposeFileErr)
			os.Exit(1)
		}

		err := docker.DockerCompose(dockerComposeFile, tag, loglevel)
		if err != nil {
			HandleDockerError(err)
			os.Exit(1)
		}

		_, pingHealthErr := docker.PingHealth(healthEndpoint)
		if pingHealthErr != nil {
			HandleDockerError(pingHealthErr)
			os.Exit(1)
		}
	}
}
