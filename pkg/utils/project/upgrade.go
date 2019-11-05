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

package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

func UpgradeProjects(c *cli.Context) *ProjectError {
	fmt.Println("About to upgrade projects")

	oldDir := strings.TrimSpace(c.String("workspace"))

	// Check to see if the workspace exists
	_, err := os.Stat(oldDir)
	if err != nil {
		return &ProjectError{errBadPath, err, err.Error()}
	}

	projectDir := oldDir + "/.projects/"
	// Check to see if the .projects dir exists
	_, fileerr := os.Stat(projectDir)
	if fileerr != nil {
		return &ProjectError{textNoProjects, fileerr, fileerr.Error()}
	}

	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			err = errors.New(textUpgradeError)
			return &ProjectError{errOpFileParse, err, textUpgradeError}
		}

		if !info.IsDir() {
			file, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}
			var result map[string]string
			json.Unmarshal([]byte(file), &result)

			language := result["language"]
			projectType := result["projectType"]
			name := result["name"]
			location := result["workspace"] + name
			fmt.Println("Calling bind for project " + name + "," + projectType + "," + language)

			response, binderr := Bind(location, name, language, projectType, "local")
			PrintAsJSON := c.GlobalBool("json")
			if binderr != nil {
				fmt.Println(err)
			} else {
				if PrintAsJSON {
					jsonResponse, _ := json.Marshal(response)
					fmt.Println(string(jsonResponse))
				} else {
					fmt.Println("Project ID: " + response.ProjectID)
					fmt.Println("Status: " + response.Status)
				}
			}
		}
		return nil
	})
	return nil

}
