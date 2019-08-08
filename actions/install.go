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

//InstallCommand to pull images from dockerhub
func InstallCommand(c *cli.Context) {
	tag := c.String("tag")
	jsonOutput := c.Bool("json")

	imageArr := [3]string{"docker.io/eclipse/codewind-pfe-amd64:",
		"docker.io/eclipse/codewind-performance-amd64:",
		"docker.io/eclipse/codewind-initialize-amd64:"}

	targetArr := [3]string{"codewind-pfe-amd64:",
		"codewind-performance-amd64:",
		"codewind-initialize-amd64:"}

	for i := 0; i < len(imageArr); i++ {
		utils.PullImage(imageArr[i]+tag, jsonOutput)
		utils.TagImage(imageArr[i]+tag, targetArr[i]+tag)
	}

	fmt.Println("Image Tagging Successful")
}
