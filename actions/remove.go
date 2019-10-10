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
	"github.com/urfave/cli"
)

//RemoveCommand to remove all codewind and project images
func RemoveCommand(c *cli.Context) {
	tag := c.String("tag")
	imageArr := []string{
		"eclipse/codewind-pfe-amd64:" + tag,
		"eclipse/codewind-performance-amd64:" + tag,
		"eclipse/codewind-initialize-amd64:" + tag,
		"cw-",
	}
	networkName := "codewind"

	images, err := utils.GetImageList()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Println("Removing Codewind docker images..")

	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		imageTags := strings.Join(image.RepoTags, " ")
		for _, key := range imageArr {
			if strings.HasPrefix(imageRepo, key) || strings.HasPrefix(imageTags, key) {
				if len(image.RepoTags) > 0 {
					fmt.Println("Deleting Image ", image.RepoTags[0], "... ")
				} else {
					fmt.Println("Deleting Image ", image.ID, "... ")
				}
				utils.RemoveImage(image.ID)
			}
		}
	}

	networks, err := utils.GetNetworkList()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	for _, network := range networks {
		if strings.Contains(network.Name, networkName) {
			fmt.Print("Removing docker network: ", network.Name, "... ")
			utils.RemoveNetwork(network)
		}
	}
}
