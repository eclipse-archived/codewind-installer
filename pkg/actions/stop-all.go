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

// StopAllCommand to stop codewind and project containers
func StopAllCommand(c *cli.Context, dockerComposeFile string) {
	tag := c.String("tag")

	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	containers, err := docker.GetContainerList(dockerClient)
	if err != nil {
		HandleDockerError(err)
		os.Exit(1)
	}

	dockerErr = docker.DockerComposeStop(tag, dockerComposeFile)
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	fmt.Println("Stopping Project containers")
	containersToRemove := docker.GetContainersToRemove(containers)
	for _, container := range containersToRemove {
		fmt.Println("Stopping container ", container.Names[0], "... ")
		docker.StopContainer(dockerClient, container)
	}
}
