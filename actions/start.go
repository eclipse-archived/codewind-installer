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

	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

//StartCommand to start the codewind conainers
func StartCommand(c *cli.Context, tempFilePath string, healthEndpoint string) {
	status := utils.CheckContainerStatus()

	if status {
		fmt.Println("Codewind is already running!")
	} else {
		tag := c.String("tag")
		debug := c.Bool("debug")
		fmt.Println("Debug:", debug)
		utils.CreateTempFile(tempFilePath)
		utils.WriteToComposeFile(tempFilePath, debug)
		utils.DockerCompose(tempFilePath, tag)
		utils.DeleteTempFile(tempFilePath) // Remove installer-docker-compose.yaml
		utils.PingHealth(healthEndpoint)
	}
}
