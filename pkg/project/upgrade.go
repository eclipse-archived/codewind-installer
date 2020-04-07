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
	"io/ioutil"
	"os"
	"path/filepath"
)

// UpgradeProjects : Upgrades Projects
func UpgradeProjects(oldDir string) (*map[string]interface{}, *ProjectError) {
	// Check to see if the workspace exists
	_, err := os.Stat(oldDir)
	if err != nil {
		return nil, &ProjectError{errBadPath, err, err.Error()}
	}
	projectDir := oldDir + "/.projects/"
	// Check to see if the .projects dir exists
	_, fileerr := os.Stat(projectDir)
	if fileerr != nil {
		return nil, &ProjectError{textNoProjects, fileerr, fileerr.Error()}
	}

	migrationStatus := make(map[string]interface{})
	migrationStatus["migrated"] = make([]string, 0)
	migrationStatus["failed"] = make([]interface{}, 0)

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

			// only read .inf files
			extension := filepath.Ext(path)
			if extension != ".inf" {
				return nil
			}

			var result map[string]string
			err = json.Unmarshal([]byte(file), &result)
			if err != nil {
				return nil
			}

			language := result["language"]
			projectType := result["projectType"]
			name := result["name"]
			location := oldDir + "/" + name

			if language != "" && projectType != "" && name != "" && location != "" {
				_, bindErr := Bind(location, name, language, projectType, "local")
				if bindErr != nil {
					errResponse := make(map[string]string)
					errResponse["projectName"] = name
					errResponse["error"] = bindErr.Desc
					migrationStatus["failed"] = append(migrationStatus["failed"].([]interface{}), &errResponse)
				} else {
					// attempt to delete the inf file
					os.Remove(path)
					migrationStatus["migrated"] = append(migrationStatus["migrated"].([]string), name)
				}
			} else {
				errResponse := make(map[string]string)
				errResponse["projectName"] = name
				errResponse["error"] = "Unable to upgrade project, failed to determine project details"
				migrationStatus["failed"] = append(migrationStatus["failed"].([]interface{}), &errResponse)
			}
		}
		return nil
	})
	return &migrationStatus, nil
}
