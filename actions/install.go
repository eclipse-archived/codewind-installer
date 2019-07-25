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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

//InstallCommand to pull images from dockerhub
func InstallCommand(c *cli.Context) {
	tag := c.String("tag")
	jsonOutput := c.Bool("json")

	imageArr := [3]string{"docker.io/ibmcom/codewind-pfe-amd64:",
		"docker.io/ibmcom/codewind-performance-amd64:",
		"docker.io/ibmcom/codewind-initialize-amd64:"}

	targetArr := [3]string{"codewind-pfe-amd64:",
		"codewind-performance-amd64:",
		"codewind-initialize-amd64:"}

	for i := 0; i < len(imageArr); i++ {
		utils.PullImage(imageArr[i]+tag, "", jsonOutput)
		utils.TagImage(imageArr[i]+tag, targetArr[i]+tag)
	}

	fmt.Println("Image Tagging Successful")
}

//InstallDevCommand to pull images from artifactory
func InstallDevCommand() {
	authConfig := types.AuthConfig{
		Username: os.Getenv("USER"),
		Password: os.Getenv("PASS"),
	}
	encodedJSON, err := json.Marshal(authConfig)
	errors.CheckErr(err, 106, "")

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	imageArr := [3]string{"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe-amd64",
		"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance-amd64",
		"sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize-amd64"}

	targetArr := [3]string{"codewind-pfe-amd64:latest",
		"codewind-performance-amd64:latest",
		"codewind-initialize-amd64:latest"}

	for i := 0; i < len(imageArr); i++ {
		utils.PullImage(imageArr[i], authStr, false)
		utils.TagImage(imageArr[i], targetArr[i])
	}

	fmt.Println("Image Tagging Successful")
}
