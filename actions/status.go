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
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/apiroutes"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/eclipse/codewind-installer/utils/deployments"
	"github.com/urfave/cli"
)

// StatusCommand : to show the status
func StatusCommand(c *cli.Context) {
	targetDeployment, err := deployments.GetTargetDeployment()
	if err != nil {
		log.Fatal(err.Error())
	}
	if strings.EqualFold(targetDeployment.ID, "local") {
		StatusCommandLocalDeployment(c)
	} else {
		StatusCommandRemoteDeployment(c, targetDeployment)
	}
}

// StatusCommandRemoteDeployment : Output remote deployment details
func StatusCommandRemoteDeployment(c *cli.Context, d *deployments.Deployment) {
	jsonOutput := c.Bool("json")
	apiResponse, err := apiroutes.GetAPIEnvironment(c, d.URL)
	if err != nil {
		if jsonOutput {
			type status struct {
				Status   string   `json:"status"`
				Versions []string `json:"installed-versions"`
			}
			respStatus := &status{
				Status:   "stopped",
				Versions: []string{},
			}
			output, err := json.Marshal(respStatus)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind remote deployment did not respond on " + d.URL)
			log.Println(err)
		}
		os.Exit(0)
	}

	// Codewind responded
	if jsonOutput {
		type status struct {
			Status   string   `json:"status"`
			URL      string   `json:"url"`
			Versions []string `json:"installed-versions"`
			Started  []string `json:"started"`
		}
		versions := []string{}
		resp := &status{
			Status:   "started",
			Versions: append(versions, apiResponse.Version),
			Started:  append(versions, apiResponse.Version),
			URL:      d.URL,
		}
		output, _ := json.Marshal(resp)
		fmt.Println(string(output))
	} else {
		fmt.Println("Codewind is installed and running on: " + d.URL)
	}
	os.Exit(0)
}

// StatusCommandLocalDeployment : Output local deployment details
func StatusCommandLocalDeployment(c *cli.Context) {
	jsonOutput := c.Bool("json")
	containersAreRunning, err := utils.CheckContainerStatus()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if containersAreRunning {
		// Started
		hostname, port, err := utils.GetPFEHostAndPort()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}

		if jsonOutput {

			imageTagArr, err := utils.GetImageTags()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
			}

			containerTagArr, err := utils.GetContainerTags()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
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

	imagesAreInstalled, err := utils.CheckImageStatus()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if imagesAreInstalled {
		// Installed but not started
		if jsonOutput {

			imageTagArr, err := utils.GetImageTags()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
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
		if jsonOutput {
			output, _ := json.Marshal(map[string]string{"status": "uninstalled"})
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is not installed")
		}
		os.Exit(0)
	}
	return
}
