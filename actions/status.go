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
	"os"

	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

//StatusCommand to show the status
func StatusCommand(c *cli.Context) {
	jsonOutput := c.Bool("json")
	if utils.CheckContainerStatus() {
		// STARTED
		hostname, port := utils.GetPFEHostAndPort()
		if jsonOutput {

			type status struct {
				Status   string   `json:"status"`
				URL      string   `json:"url"`
				Versions []string `json:"installed-versions"`
				Started  []string `json:"started"`
			}

			imageTagArr := utils.GetImageTag()
			containerTagArr := utils.GetContainerTag()
			resp := &status{
				Status:   "started",
				URL:      "http://" + hostname + ":" + port,
				Versions: imageTagArr,
				Started:  containerTagArr,
			}

			//output, _ := json.Marshal(map[string]string{"status": "started", "url": "http://" + hostname + ":" + port})
			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is installed and running on http://" + hostname + ":" + port)
		}
		os.Exit(0)
	}

	if utils.CheckImageStatus() {
		// INSTALLED NOT STARTED
		if jsonOutput {

			type status struct {
				Status   string   `json:"status"`
				Versions []string `json:"installed-versions"`
				Started  []string `json:"started"`
			}

			tagArr := utils.GetImageTag()
			resp := &status{
				Status:   "stopped",
				Versions: tagArr,
				Started:  []string{},
			}

			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
		} else {
			fmt.Println("Codewind is installed but not running")
		}
		os.Exit(0)
	} else {
		// NOT INSTALLED
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
