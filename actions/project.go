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

	"github.com/eclipse/codewind-installer/utils/project"
	"github.com/urfave/cli"
)

func ProjectValidate(c *cli.Context) {
	err := project.ValidateProject(c)
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(0)
}

func ProjectCreate(c *cli.Context) {
	err := project.DownloadTemplate(c)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func ProjectSync(c *cli.Context) {
	PrintAsJSON := c.GlobalBool("json")
	response, err := project.SyncProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if PrintAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			fmt.Println("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

func ProjectBind(c *cli.Context) {
	PrintAsJSON := c.GlobalBool("json")
	response, err := project.BindProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if PrintAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			fmt.Println("Project ID: " + response.ProjectID)
			fmt.Println("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

func UpgradeProjects(c *cli.Context) {
	err := project.UpgradeProjects(c)
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(0)
}

// ProjectAddTargetDeployment : Add project to a deployment
func ProjectAddTargetDeployment(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	depID := strings.TrimSpace(strings.ToLower(c.String("depid")))
	err := project.AddDeploymentTarget(projectID, depID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(project.Result{Status: "OK", StatusMessage: "Project target added successfully"})
	fmt.Println(string(response))
	os.Exit(0)
}

// ProjectTargetList : List deployment targets for a project
func ProjectTargetList(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	deploymentTargets, err := project.ListTargetDeployments(projectID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(deploymentTargets)
	fmt.Println(string(response))
	os.Exit(0)
}

// ProjectRemoveTargetDeployment : Remove a project from a deployment
func ProjectRemoveTargetDeployment(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	depID := strings.TrimSpace(strings.ToLower(c.String("depid")))
	err := project.RemoveDeploymentTarget(projectID, depID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(project.Result{Status: "OK", StatusMessage: "Project target removed successfully"})
	fmt.Println(string(response))
	os.Exit(0)
}
