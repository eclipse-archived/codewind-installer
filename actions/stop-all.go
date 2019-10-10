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

//StopAllCommand to stop codewind and project containers
func StopAllCommand() {
	containerArr := []string{
		"codewind-pfe",
		"codewind-performance",
		"cw-",
		"appsody",
	}

	containers, err := utils.GetContainerList()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

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
}
