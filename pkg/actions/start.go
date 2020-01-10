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

	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// StartCommand : start the codewind containers
func StartCommand(c *cli.Context, dockerComposeFile string, healthEndpoint string) {
	status, err := utils.CheckContainerStatus()
	if err != nil {
		errors.PrintError(err, printAsJSON)
		os.Exit(1)
	}

	if status {
		fmt.Println("Codewind is already running!")
	} else {
		tag := c.String("tag")
		debug := c.Bool("debug")
		fmt.Println("Debug:", debug)

		utils.CreateTempFile(dockerComposeFile)
		utils.WriteToComposeFile(dockerComposeFile, debug)

		err := utils.DockerCompose(dockerComposeFile, tag)
		if err != nil {
			errors.PrintError(err, printAsJSON)
			os.Exit(1)
		}

		_, err = utils.PingHealth(healthEndpoint)
		if err != nil {
			errors.PrintError(err, printAsJSON)
			os.Exit(1)
		}
	}
}
