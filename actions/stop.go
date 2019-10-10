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
	"strings"

	"github.com/eclipse/codewind-installer/utils"
)

//StopCommand to stop only the codewind containers
func StopCommand() {
	containerArr := [2]string{}
	containerArr[0] = "codewind-pfe"
	containerArr[1] = "codewind-performance"

	containers, err := utils.GetContainerList()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Println("Only stopping Codewind containers. To stop project containers, please use 'stop-all'")

	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				fmt.Println("Stopping container ", container.Names, "... ")
				utils.StopContainer(container)
			}
		}
	}
}
