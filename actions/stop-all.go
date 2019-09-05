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

	"github.com/eclipse/codewind-installer/utils"
)

//StopAllCommand to stop codewind and project containers
func StopAllCommand() {
	containerArr := [2]string{}
	containerArr[0] = "codewind-pfe"
	containerArr[1] = "codewind-performance"
	containerArr[0] = "cw-"
	containerArr[1] = "appsody"

	containers := utils.GetContainerList()

	fmt.Println("Stopping Codewind and Project containers")
	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				fmt.Println(container)
				// stop appsody containers with "cw-" in their name, and all other cw containers
				if key != "appsody" || strings.Contains(container.Names[0], "cw-") {
					fmt.Println("Stopping container ", container.Names, "... ")
					utils.StopContainer(container)
				}
			}
		}
	}
}
