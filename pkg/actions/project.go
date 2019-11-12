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
	"strings"

<<<<<<< HEAD
	"github.com/eclipse/codewind-installer/pkg/project"
=======
	"github.com/eclipse/codewind-installer/pkg/utils/project"
	logr "github.com/sirupsen/logrus"
>>>>>>> replace 'fmt.Print' with logrus #1
	"github.com/urfave/cli"
)

// ProjectValidate : Validate a project
func ProjectValidate(c *cli.Context) {
	err := project.ValidateProject(c)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// ProjectCreate : Downloads template and creates a new project
func ProjectCreate(c *cli.Context) {
	err := project.DownloadTemplate(c)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
}

// ProjectSync : Does a project Sync
func ProjectSync(c *cli.Context) {
	PrintAsJSON := c.GlobalBool("json")
	response, err := project.SyncProject(c)
	if err != nil {
		logr.Errorln(err.Err)
		os.Exit(1)
	} else {
		if PrintAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			logr.Infoln("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

// ProjectBind : Does a project bind
func ProjectBind(c *cli.Context) {
	PrintAsJSON := c.GlobalBool("json")
	response, err := project.BindProject(c)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	} else {
		if PrintAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			logr.Infoln("Project ID: " + response.ProjectID)
			logr.Infoln("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

// UpgradeProjects : Upgrades projects
func UpgradeProjects(c *cli.Context) {
	dir := strings.TrimSpace(c.String("workspace"))
	response, err := project.UpgradeProjects(dir)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
	PrettyPrintJSON(response)
	os.Exit(0)
}

// ProjectSetConnection : Set connection for a project
func ProjectSetConnection(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	err := project.SetConnection(conID, projectID)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
	response, _ := json.Marshal(project.Result{Status: "OK", StatusMessage: "Project target added successfully"})
	fmt.Println(string(response))
	os.Exit(0)
}

// ProjectGetConnection : List connection for a project
func ProjectGetConnection(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	connectionTargets, err := project.GetConnectionID(projectID)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
	logr.Infoln(connectionTargets)
	os.Exit(0)
}

// ProjectRemoveConnection : Remove Connection from  a project
func ProjectRemoveConnection(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	err := project.ResetConnectionFile(projectID)
	if err != nil {
		logr.Errorln(err.Error())
		os.Exit(1)
	}
	response, _ := json.Marshal(project.Result{Status: "OK", StatusMessage: "Project target removed successfully"})
	fmt.Println(string(response))
	os.Exit(0)
}
