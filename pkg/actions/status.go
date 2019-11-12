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
	"net/http"
	"os"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/utils"
<<<<<<< HEAD
=======
	"github.com/eclipse/codewind-installer/pkg/utils/connections"
	logr "github.com/sirupsen/logrus"
>>>>>>> replace 'fmt.Print' with logrus #1
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
	jsonOutput := c.Bool("json") || c.GlobalBool("json")
	conID := c.String("conid")
	connection, conErr := connections.GetConnectionByID(conID)
	if conErr != nil {
		logr.Errorln(conErr)
		os.Exit(1)
	}

	PFEReady, err := apiroutes.IsPFEReady(http.DefaultClient, connection.URL)
	if err != nil || PFEReady == false {
		if jsonOutput {
			type status struct {
				Status string `json:"status"`
			}
			resp := &status{
				Status: "stopped",
			}
			if err != nil {
				logr.Errorln(err)
				os.Exit(1)
			}
			output, _ := json.Marshal(resp)
			fmt.Println(string(output))
			os.Exit(1)
		} else {
			logr.Errorln("Codewind did not respond on remote connection", conID)
			logr.Errorln(err)
		}
	}

	// Codewind responded
	if jsonOutput {
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
		logr.Infoln("Remote Codewind is installed and running")
	}
	os.Exit(0)
}

// StatusCommandLocalConnection : Output local connection details
func StatusCommandLocalConnection(c *cli.Context) {
	jsonOutput := c.Bool("json") || c.GlobalBool("json")
	if utils.CheckContainerStatus() {
		// Started
		hostname, port := utils.GetPFEHostAndPort()
		if jsonOutput {

			imageTagArr := utils.GetImageTags()
			containerTagArr := utils.GetContainerTags()

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
			logr.Infoln("Codewind is installed and running on http://" + hostname + ":" + port)
		}
		os.Exit(0)
	}

	if utils.CheckImageStatus() {
		// Installed but not started
		if jsonOutput {

			imageTagArr := utils.GetImageTags()

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
			logr.Infoln("Codewind is installed but not running")
		}
		os.Exit(0)
	} else {
		// Not installed
		if jsonOutput {
			output, _ := json.Marshal(map[string]string{"status": "uninstalled"})
			fmt.Println(string(output))
		} else {
			logr.Infoln("Codewind is not installed")
		}
		os.Exit(0)
	}
	return
}
