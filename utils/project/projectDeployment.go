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
	"path"
	"runtime"
	"strings"

	"github.com/eclipse/codewind-installer/utils/deployments"
)

// Target : A Deployment target
type Target struct {
	DeploymentID string `json:"id"`
}

// DeploymentTargets : Structure of the project-deployments file
type DeploymentTargets struct {
	SchemaVersion     int      `json:"schemaVersion"`
	DeploymentTargets []Target `json:"deploymentTargets"`
}

const deploymentTargetSchemaVersion = 1

// AddDeploymentTarget : Add a deployment target
func AddDeploymentTarget(projectID string, depID string) *ProjectError {

	deployment, depErr := deployments.GetDeploymentByID(depID)
	if depErr != nil || deployment == nil {
		projError := errors.New("Deployment unknown")
		return &ProjectError{"dep_not_found", projError, projError.Error()}
	}

	// Check if projectID is supplied in correct format
	if !IsProjectIDValid(projectID) {
		projError := errors.New(textInvalidProjectID)
		return &ProjectError{errOpInvalidID, projError, projError.Error()}
	}

	// Load the project-deployment.json
	deploymentTargets, projError := loadTargets(projectID)

	if projError != nil && deploymentTargets == nil {
		_, err := os.Stat(getProjectDeploymentsFilename(projectID))
		if os.IsNotExist(err) {
			os.MkdirAll(getProjectDeploymentConfigDir(), 0777)
			projErr := ResetTargetFile(projectID)
			if projErr != nil {
				return projErr
			}
		}
		deploymentTargets, projErr := loadTargets(projectID)
		if projErr != nil {
			return projErr
		}
		target := Target{
			DeploymentID: depID,
		}
		deploymentTargets.DeploymentTargets = append(deploymentTargets.DeploymentTargets, target)
		projError := saveDeploymentTargets(projectID, deploymentTargets)
		if projError != nil {
			return projError
		}
		return nil
	}

	// Check if the deployment has been added to the project already
	for i := 0; i < len(deploymentTargets.DeploymentTargets); i++ {
		if strings.EqualFold(depID, deploymentTargets.DeploymentTargets[i].DeploymentID) {
			projError := errors.New(textDeploymentExists)
			return &ProjectError{errOpConflict, projError, projError.Error()}
		}
	}

	// Add the deployment to the project-deployments file
	target := Target{
		DeploymentID: depID,
	}
	deploymentTargets.DeploymentTargets = append(deploymentTargets.DeploymentTargets, target)

	// Save the project-deployments file
	projError = saveDeploymentTargets(projectID, deploymentTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// ResetTargetFile : Reset target file
func ResetTargetFile(projectID string) *ProjectError {
	deploymentTargets := DeploymentTargets{
		SchemaVersion: deploymentTargetSchemaVersion,
	}
	projError := saveDeploymentTargets(projectID, &deploymentTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// RemoveDeploymentTarget : Remove deployment target from project-deployments file
func RemoveDeploymentTarget(projectID string, depID string) *ProjectError {

	deployment, depErr := deployments.GetDeploymentByID(depID)
	if depErr != nil || deployment == nil {
		projError := errors.New("Deployment unknown")
		return &ProjectError{"dep_not_found", projError, projError.Error()}
	}

	// Check if projectID is supplied in correct format
	if !IsProjectIDValid(projectID) {
		projError := errors.New(textInvalidProjectID)
		return &ProjectError{errOpInvalidID, projError, projError.Error()}
	}

	// Load the deployments
	deploymentTargets, projErr := loadTargets(projectID)
	if projErr != nil {
		return projErr
	}

	deploymentFound := false

	// Remove the deployment
	for i := 0; i < len(deploymentTargets.DeploymentTargets); i++ {
		if strings.EqualFold(depID, deploymentTargets.DeploymentTargets[i].DeploymentID) {
			copy(deploymentTargets.DeploymentTargets[i:], deploymentTargets.DeploymentTargets[i+1:])
			deploymentFound = true
			deploymentTargets.DeploymentTargets = deploymentTargets.DeploymentTargets[:len(deploymentTargets.DeploymentTargets)-1]
		}
	}

	if !deploymentFound {
		projErr := errors.New(textDepMissing)
		return &ProjectError{errOpNotFound, projErr, projErr.Error()}
	}

	// Save the project-deployments file
	projError := saveDeploymentTargets(projectID, deploymentTargets)
	if projError != nil {
		return projError
	}
	return nil
}

// ListTargetDeployments : List the target deployments for this project
func ListTargetDeployments(projectID string) (*DeploymentTargets, *ProjectError) {
	deploymentTargets, projErr := loadTargets(projectID)
	if projErr != nil {
		return nil, projErr
	}
	return deploymentTargets, nil
}

// getProjectDeploymentConfigDir : get directory path to the deployments file
func getProjectDeploymentConfigDir() string {
	const GOOS string = runtime.GOOS
	homeDir := ""
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return path.Join(homeDir, ".codewind", "config", "deployments")
}

// getProjectDeploymentsFilename  : get full file path of deployments file
func getProjectDeploymentsFilename(projectID string) string {
	return path.Join(getProjectDeploymentConfigDir(), projectID+".json")
}

// saveDeploymentTargets : write the targets file in JSON format
func saveDeploymentTargets(projectID string, deploymentTargets *DeploymentTargets) *ProjectError {
	body, err := json.MarshalIndent(deploymentTargets, "", "\t")
	if err != nil {
		return &ProjectError{errOpFileParse, err, err.Error()}
	}
	projError := ioutil.WriteFile(getProjectDeploymentsFilename(projectID), body, 0644)
	if projError != nil {
		return &ProjectError{errOpFileWrite, projError, projError.Error()}
	}
	return nil
}

// loadTargets :  Loads the config file for a project
func loadTargets(projectID string) (*DeploymentTargets, *ProjectError) {
	projectID = strings.ToLower(projectID)
	file, err := ioutil.ReadFile(getProjectDeploymentsFilename(projectID))
	if err != nil {
		return nil, &ProjectError{errOpFileLoad, err, err.Error()}
	}

	// parse the file
	projectDeploymentTargets := DeploymentTargets{}
	err = json.Unmarshal([]byte(file), &projectDeploymentTargets)
	if err != nil {
		return nil, &ProjectError{errOpFileParse, err, err.Error()}
	}
	return &projectDeploymentTargets, nil
}
