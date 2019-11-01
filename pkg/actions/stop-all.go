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

	"github.com/eclipse/codewind-installer/pkg/utils"
)

//StopAllCommand to stop codewind and project containers
func StopAllCommand() {
	containerArr := []string{
		"codewind-pfe",
		"codewind-performance",
		"cw-",
		"appsody",
	}

	containers := utils.GetContainerList()

	fmt.Println("Stopping Codewind and Project containers")
	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				if key != "appsody" || strings.Contains(container.Names[0], "cw-") {
					fmt.Println("Stopping container ", container.Names[0], "... ")
					utils.StopContainer(container)
					break
				}
			}
		}
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
