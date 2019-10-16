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
	"path"
	"regexp"
	"strings"

	"github.com/eclipse/codewind-installer/apiroutes"
	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/eclipse/codewind-installer/utils/project"
	"github.com/urfave/cli"
)

// DownloadTemplate using the url/link provided
func DownloadTemplate(c *cli.Context) {
	destination := c.Args().Get(0)

	if destination == "" {
		log.Fatal("destination not set")
	}

	projectDir := path.Base(destination)

	// Remove invalid characters from the string we will use
	// as the project name in the template.
	r := regexp.MustCompile("[^a-zA-Z0-9._-]")
	projectName := r.ReplaceAllString(projectDir, "")
	if len(projectName) == 0 {
		projectName = "PROJ_NAME_PLACEHOLDER"
	}

	url := c.String("u")

	err := utils.DownloadFromURLThenExtract(url, destination)
	if err != nil {
		log.Fatal(err)
	}
	err = utils.ReplaceInFiles(destination, "[PROJ_NAME_PLACEHOLDER]", projectName)
	if err != nil {
		log.Fatal(err)
	}
}

// checkIsExtension checks if a project is an extension project and run associated commands as necessary
func checkIsExtension(projectPath string, c *cli.Context) (string, error) {

	extensions, err := apiroutes.GetExtensions()
	if err != nil {
		log.Println("There was a problem retrieving extensions data")
		return "unknown", err
	}

	params := make(map[string]string)
	commandName := "postProjectValidate"

	// determine if type:subtype hint was given
	// but only if url was not given
	if c.String("u") == "" && c.String("t") != "" {
		parts := strings.Split(c.String("t"), ":")
		params["type"] = parts[0]
		if len(parts) > 1 {
			params["subtype"] = parts[1]
		}
		commandName = "postProjectValidateWithType"
	}

	for _, extension := range extensions {

		var isMatch bool

		if len(params) > 0 {
			// check if extension project type matched the hinted type
			isMatch = extension.ProjectType == params["type"]
		} else {
			// check if project contains the detection file an extension defines
			isMatch = extension.Detection != "" && utils.PathExists(path.Join(projectPath, extension.Detection))
		}

		if isMatch {

			var cmdErr error

			// check if there are any commands to run
			for _, command := range extension.Commands {
				if command.Name == commandName {
					cmdErr = utils.RunCommand(projectPath, command, params)
					break
				}
			}

			return extension.ProjectType, cmdErr
		}
	}

	return "", nil
}

// ValidateProject returns the language and buildType for a project at given filesystem path,
// and writes a default .cw-settings file to that project
func ValidateProject(c *cli.Context) {
	projectPath := c.Args().Get(0)
	utils.CheckProjectPath(projectPath)
	validationStatus := "success"
	// result could be ProjectType or string, so define as an interface
	var validationResult interface{}
	language, buildType := utils.DetermineProjectInfo(projectPath)
	validationResult = ProjectType{
		Language:  language,
		BuildType: buildType,
	}
	extensionType, err := checkIsExtension(projectPath, c)
	if extensionType != "" {
		if err == nil {
			validationResult = ProjectType{
				Language:  language,
				BuildType: extensionType,
			}
		} else {
			validationStatus = "failed"
			validationResult = err.Error()
		}
	}

	response := ValidationResponse{
		Status: validationStatus,
		Path:   projectPath,
		Result: validationResult,
	}
	projectInfo, err := json.Marshal(response)

	errors.CheckErr(err, 203, "")
	writeCwSettingsIfNotInProject(projectPath, buildType)
	fmt.Println(string(projectInfo))
}

func writeCwSettingsIfNotInProject(projectPath string, BuildType string) {
	pathToCwSettings := path.Join(projectPath, ".cw-settings")
	pathToLegacySettings := path.Join(projectPath, ".mc-settings")

	if _, err := os.Stat(pathToLegacySettings); os.IsExist(err) {
		utils.RenameLegacySettings(pathToLegacySettings, pathToCwSettings)
	} else if _, err := os.Stat(pathToCwSettings); os.IsNotExist(err) {
		utils.WriteNewCwSettings(pathToCwSettings, BuildType)
	}
}

func ProjectSync(c *cli.Context) {
	err := project.SyncProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}

func ProjectBind(c *cli.Context) {
	err := project.BindProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}
