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
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/urfave/cli"
)

// RemoveProject : Unbind a project from Codewind and delete json connection file
func RemoveProject(c *cli.Context) *ProjectError {
	projectID := strings.TrimSpace(c.String("id"))
	deleteFiles := c.Bool("delete")
	projectPath := ""

	// Get the connection for this project
	conID, conErr := GetConnectionID(projectID)
	if conErr != nil {
		return conErr
	}

	conInfo, conInfoErr := connections.GetConnectionByID(conID)
	if conInfoErr != nil {
		return &ProjectError{conInfoErr.Op, conInfoErr.Err, conInfoErr.Desc}
	}

	conURL, configErr := config.PFEOriginFromConnection(conInfo)
	if configErr != nil {
		return &ProjectError{configErr.Op, configErr.Err, configErr.Desc}
	}

	// If we are deleting the source, retrieve project to find out the path
	if deleteFiles {
		project, projErr := GetProjectFromID(http.DefaultClient, conInfo, conURL, projectID)
		if projErr != nil {
			return projErr
		}
		projectPath = project.LocationOnDisk
	}

	// Unbind the project from codewind
	projError := Unbind(http.DefaultClient, conInfo, conURL, projectID)
	if projError != nil {
		return projError
	}

	// Delete the associated connection file
	projError = RemoveConnectionFile(projectID)
	if projError != nil {
		return projError
	}

	// Delete the source if the flag is set
	if deleteFiles {
		var err = os.RemoveAll(projectPath)
		if err != nil {
			return &ProjectError{errOpFileDelete, err, err.Error()}
		}
	}
	return nil
}
