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
)

//StopAllCommand to stop codewind and project containers
func StopAllCommand() {
	containers := utils.GetContainerList()

	fmt.Println("Stopping Codewind and Project containers")
	containersToRemove := getContainersToRemove(containers)
	for _, container := range containersToRemove {
		fmt.Println("Stopping container ", container.Names[0], "... ")
		utils.StopContainer(container)
	}

	networkName := "codewind"
	networks := utils.GetNetworkList()
	fmt.Println("Removing Codewind docker networks..")
	for _, network := range networks {
		if strings.Contains(network.Name, networkName) {
			fmt.Print("Removing docker network: ", network.Name, "... ")
			utils.RemoveNetwork(network)
		}
	}
}

func getContainersToRemove(containerList []types.Container) []types.Container {
	codewindContainerNames := []string{
		"codewind-pfe",
		"codewind-performance",
	}

	// Docker returns all the names with a "/" on the front
	projectContainerPrefix := "/cw-"

	containersToRemove := []types.Container{}
	for _, container := range containerList {
		for _, key := range codewindContainerNames {
			if strings.Contains(container.Names[0], key) || strings.HasPrefix(container.Names[0], projectContainerPrefix) {
				containersToRemove = append(containersToRemove, container)
				break
			}
		}
	}
	return containersToRemove
}
