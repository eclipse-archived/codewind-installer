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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/urfave/cli"
)

// StatusCommand : to show the status
func StatusCommand(c *cli.Context) {
	conID := c.String("conid")
	if conID != "" && conID != "local" {
		StatusCommandRemoteConnection(c)
	} else {
		StatusCommandLocalConnection(c)
	}
}

// StatusCommandRemoteConnection : Output remote connection details
func StatusCommandRemoteConnection(c *cli.Context) {
	conID := c.String("conid")
	connection, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		fmt.Println(conErr)
		os.Exit(1)
	}

	PFEReady, err := apiroutes.IsPFEReady(http.DefaultClient, connection.URL)
	if err != nil || PFEReady == false {
		if printAsJSON {
			type status struct {
				Status string `json:"status"`
			}
			resp := &status{
				Status: "stopped",
			}
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
			os.Exit(1)
		} else {
			fmt.Println("Codewind did not respond on remote connection", conID)
			log.Println(err)
		}
	}

	// Codewind responded
	if printAsJSON {
		type status struct {
			Status   string   `json:"status"`
			URL      string   `json:"url"`
			Versions []string `json:"installed-versions"`
			Started  []string `json:"started"`
		}
		resp := &status{
			Status: "started",
		}
		output, _ := json.Marshal(resp)
		fmt.Println(string(output))
	} else {
		fmt.Println("Remote Codewind is installed and running")
	}
	os.Exit(0)
}

// StatusCommandLocalConnection : Output local connection details
func StatusCommandLocalConnection(c *cli.Context) {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	containersAreRunning, err := docker.CheckContainerStatus(dockerClient)
	if err != nil {
		HandleDockerError(err)
		os.Exit(1)
	}

	if containersAreRunning {
		// Started
		hostname, port, err := docker.GetPFEHostAndPort(dockerClient)
		if err != nil {
			HandleDockerError(err)
			os.Exit(1)
		}
		if printAsJSON {
			imageTagArr, err := docker.GetImageTags(dockerClient)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			containerTagArr, err := docker.GetContainerTags(dockerClient)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			type status struct {
				Status   string   `json:"status"`
				URL      string   `json:"url"`
				Versions []string `json:"installed-versions"`
				Started  []string `json:"started"`
			}

			resp := &status{
				Status:   "started",
				URL:      "http://" + hostname + ":" + port,
				Versions: imageTagArr,
				Started:  containerTagArr,
			}

			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is installed and running on http://" + hostname + ":" + port)
		}
		os.Exit(0)
	}

	imagesAreInstalled, err := docker.CheckImageStatus(dockerClient)
	if err != nil {
		HandleDockerError(err)
		os.Exit(1)
	}

	if imagesAreInstalled {
		// Installed but not started
		if printAsJSON {

			imageTagArr, err := docker.GetImageTags(dockerClient)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			type status struct {
				Status   string   `json:"status"`
				Versions []string `json:"installed-versions"`
			}

			resp := &status{
				Status:   "stopped",
				Versions: imageTagArr,
			}

			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is installed but not running")
		}
		os.Exit(0)
	} else {
		// Not installed
		if printAsJSON {
			output, _ := json.Marshal(map[string]string{"status": "uninstalled"})
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is not installed")
		}
		os.Exit(0)
	}
	return
}
