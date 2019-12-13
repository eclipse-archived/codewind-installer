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
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

//StopAllCommand to stop codewind and project containers
func StopAllCommand(c *cli.Context, dockerComposeFile string) {
	tag := c.String("tag")
	containers := utils.GetContainerList()
	utils.DockerComposeStop(tag, dockerComposeFile)

	fmt.Println("Stopping Project containers")
	containersToRemove := getContainersToRemove(containers)
	for _, container := range containersToRemove {
		fmt.Println("Stopping container ", container.Names[0], "... ")
		utils.StopContainer(container)
	}

}

func getContainersToRemove(containerList []types.Container) []types.Container {
	codewindContainerPrefixes := []string{
		"/cw-",
	}

	containersToRemove := []types.Container{}
	for _, container := range containerList {
		for _, prefix := range codewindContainerPrefixes {
			if strings.HasPrefix(container.Names[0], prefix) {
				containersToRemove = append(containersToRemove, container)
				break
			}
		}
	}
	return containersToRemove
}
