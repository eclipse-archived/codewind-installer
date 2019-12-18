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

	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//StopCommand to stop only the codewind containers
func StopCommand(c *cli.Context, dockerComposeFile string) {
	tag := c.String("tag")
	printAsJSON := c.GlobalBool("json")
	fmt.Println("Only stopping Codewind containers. To stop project containers, please use 'stop-all'")
	err := utils.DockerComposeStop(tag, dockerComposeFile)
	if err != nil {
		if printAsJSON {
			fmt.Println(err.Error())
		} else {
			logr.Println(err.Desc)
		}
		os.Exit(1)
	}
}
