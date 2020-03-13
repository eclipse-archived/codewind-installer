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

	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/remote"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//RemoveCommand to remove all codewind and project images
func RemoveCommand(c *cli.Context, dockerComposeFile string) {
	tag := c.String("tag")
	if tag == "" {
		tag = "latest"
	}
	imageArr := []string{
		"cw-",
	}

	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	images, err := docker.GetImageList(dockerClient)
	if err != nil {
		HandleDockerError(err)
		os.Exit(1)
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
				docker.RemoveImage(image.ID)
			}
		}
	}

	dockerErr = docker.DockerComposeRemove(dockerComposeFile, tag)
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
}

// DoRemoteRemove : Delete a remote Codewind deployment
func DoRemoteRemove(c *cli.Context) {
	removeOptions := remote.RemoveDeploymentOptions{
		Namespace:   c.String("namespace"),
		WorkspaceID: c.String("workspace"),
	}

	_, remInstError := remote.RemoveRemote(&removeOptions)
	if remInstError != nil {
		if printAsJSON {
			fmt.Println(remInstError.Error())
		} else {
			logr.Errorf("Error: %v - %v\n", remInstError.Op, remInstError.Desc)
		}
		os.Exit(1)
	}

	os.Exit(0)
}

// DoRemoteKeycloakRemove : Delete a remote Keycloak deployment
func DoRemoteKeycloakRemove(c *cli.Context) {
	printAsJSON := c.GlobalBool("json")
	removeOptions := remote.RemoveDeploymentOptions{
		Namespace:   c.String("namespace"),
		WorkspaceID: c.String("workspace"),
	}

	_, remInstError := remote.RemoveRemoteKeycloak(&removeOptions)
	if remInstError != nil {
		if printAsJSON {
			fmt.Println(remInstError.Error())
		} else {
			logr.Errorf("Error: %v - %v\n", remInstError.Op, remInstError.Desc)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
