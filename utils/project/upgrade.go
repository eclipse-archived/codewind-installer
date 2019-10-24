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

	_, err := os.Stat(oldDir)
	if err != nil {
		return &ProjectError{errBadPath, err, err.Error()}
	}

	projectDir := oldDir + "/.projects/"
	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			panic(err)
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
			projectID := result["projectID"]

			depID, err := GetDeploymentURL(projectID)
			if err != nil {
				return err
			}
			Bind(location, name, language, projectType, depID)
		}
		return nil
	})
	return nil

}
