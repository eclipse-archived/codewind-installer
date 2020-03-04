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
	"strings"
	"text/tabwriter"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/project"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ProjectValidate : Detects the project type, and adds .cw-settings if it does not already exist
func ProjectValidate(c *cli.Context) {
	response, projectErr := project.ValidateProject(c)
	if projectErr != nil {
		fmt.Println(projectErr.Error())
		os.Exit(1)
	}
	projectInfo, _ := json.Marshal(response)
	fmt.Println(string(projectInfo))
	os.Exit(0)
}

// ProjectCreate : Downloads template, create a new project then validate it
func ProjectCreate(c *cli.Context) {
	destination := c.String("path")
	url := c.String("url")
	result, err := project.DownloadTemplate(destination, url)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	}
	if printAsJSON {
		jsonResponse, _ := json.Marshal(result)
		logr.Tracef(string(jsonResponse))
	} else {
		logr.Tracef("Project downloaded to %v", destination)
	}
	ProjectValidate(c)
}

// ProjectSync : Does a project Sync
func ProjectSync(c *cli.Context) {
	response, err := project.SyncProject(c)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	} else {
		if printAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			fmt.Println("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

// ProjectBind : Does a project bind
func ProjectBind(c *cli.Context) {
	response, err := project.BindProject(c)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	} else {
		if printAsJSON {
			jsonResponse, _ := json.Marshal(response)
			fmt.Println(string(jsonResponse))
		} else {
			fmt.Println("Project ID: " + response.ProjectID)
			fmt.Println("Status: " + response.Status)
		}
	}
	os.Exit(0)
}

// ProjectRemove : Does a project remove
func ProjectRemove(c *cli.Context) {
	err := project.RemoveProject(c)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	}
	os.Exit(0)
}

// UpgradeProjects : Upgrades projects
func UpgradeProjects(c *cli.Context) {
	dir := strings.TrimSpace(c.String("workspace"))
	response, err := project.UpgradeProjects(dir)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	}
	utils.PrettyPrintJSON(response)
	os.Exit(0)
}

// ProjectSetConnection : Set connection for a project
func ProjectSetConnection(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	err := project.SetConnection(conID, projectID)
	if err != nil {
		HandleProjectError(err)
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
		HandleProjectError(err)
		os.Exit(1)
	}
	fmt.Println(connectionTargets)
	os.Exit(0)
}

// ProjectRemoveConnection : Remove Connection from a project
func ProjectRemoveConnection(c *cli.Context) {
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	err := project.ResetConnectionFile(projectID)
	if err != nil {
		HandleProjectError(err)
		os.Exit(1)
	}
	response, _ := json.Marshal(project.Result{Status: "OK", StatusMessage: "Project target removed successfully"})
	fmt.Println(string(response))
	os.Exit(0)
}

// ProjectList : Print the list of projects to the terminal
func ProjectList(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		HandleConnectionError(conInfoErr)
		os.Exit(1)
	}

	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		HandleConfigError(conErr)
		os.Exit(1)
	}

	projects, getAllErr := project.GetAll(http.DefaultClient, conInfo, conURL)
	if getAllErr != nil {
		HandleProjectError(getAllErr)
		os.Exit(1)
	}

	if printAsJSON {
		json, _ := json.Marshal(projects)
		fmt.Println(string(json))
	} else {
		if len(projects) == 0 {
			fmt.Println("No projects bound to Codewind")
		} else {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', 0)
			fmt.Fprintln(w, "PROJECT ID \tNAME \tLANGUAGE \tAPP STATUS \tLOCATION ON DISK")
			for _, project := range projects {
				appStatus := strings.Title(project.AppStatus)
				fmt.Fprintln(w, project.ProjectID+"\t"+project.Name+"\t"+project.Language+"\t"+appStatus+"\t"+project.LocationOnDisk)
			}
			fmt.Fprintln(w)
			w.Flush()
		}
	}
	os.Exit(0)
}

// ProjectGet : Prints information about a given project using its ID
func ProjectGet(c *cli.Context) {
	conID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	projectID := strings.TrimSpace(strings.ToLower(c.String("id")))
	projectName := c.String("name")

	if projectID == "" && projectName == "" {
		logr.Errorln("Must specify either project ID (--id) or project name (--name)")
		os.Exit(1)
	}

	if projectID != "" && (conID == "local" || conID == "") {
		newConID, conIDErr := project.GetConnectionID(projectID)
		if conIDErr != nil {
			HandleProjectError(conIDErr)
			os.Exit(1)
		}
		conID = newConID
	}

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		HandleConnectionError(conInfoErr)
		os.Exit(1)
	}

	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		HandleConfigError(conErr)
		os.Exit(1)
	}

	var projectObj *project.Project
	var projectErr *project.ProjectError
	if projectID == "" && projectName != "" {
		projectObj, projectErr = project.GetProjectFromName(http.DefaultClient, conInfo, conURL, projectName)
	} else {
		projectObj, projectErr = project.GetProjectFromID(http.DefaultClient, conInfo, conURL, projectID)
	}

	if projectErr != nil {
		HandleProjectError(projectErr)
		os.Exit(1)
	}

	if printAsJSON {
		json, _ := json.Marshal(projectObj)
		fmt.Println(string(json))
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintln(w, "PROJECT ID \tNAME \tLANGUAGE \tAPP STATUS \tLOCATION ON DISK")
		appStatus := strings.Title(projectObj.AppStatus)
		fmt.Fprintln(w, projectObj.ProjectID+"\t"+projectObj.Name+"\t"+projectObj.Language+"\t"+appStatus+"\t"+projectObj.LocationOnDisk)
		fmt.Fprintln(w)
		w.Flush()
	}
	os.Exit(0)
}
