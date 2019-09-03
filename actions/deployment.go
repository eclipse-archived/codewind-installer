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
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

func getDeploymentConfigPath() string {
	const GOOS string = runtime.GOOS
	homeDir := ""
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return path.Join(homeDir, ".codewind", "config")
}

func getDeploymentConfigFilename() string {
	return path.Join(getDeploymentConfigPath(), "deployments.json")
}

type DeploymentConfig struct {
	Active      string       `json:"active"`
	Deployments []Deployment `json:"deployments"`
}

type Deployment struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Url   string `json:"url"`
}

/**
* Check the config file exist, if it does not then create a new default configuration
 */
func InitDeploymentConfigIfRequired() {
	_, err := os.Stat(getDeploymentConfigFilename())
	if os.IsNotExist(err) {
		os.MkdirAll(getDeploymentConfigPath(), 0777)
		ResetDeploymentsFile()
	}
}

/***
 * ResetDeploymentsFile
 * Creates a new / overwrites deployment config file with a default single local Codewind deployment
 */
func ResetDeploymentsFile() {
	// create the default local deployment
	initialConfig := DeploymentConfig{
		Active: "local",
		Deployments: []Deployment{
			Deployment{
				Name:  "local",
				Label: "Codewind local deployment",
				Url:   "tbd",
			},
		},
	}
	body, err := json.MarshalIndent(initialConfig, "", "\t")
	errors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	errors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

/***
 * Load the deployments configuration file from disk
 * returns:  the contents of the file, an error, an error code
 */
func loadDeploymentsConfigFile() (*DeploymentConfig, error, int) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return nil, err, 207
	}
	data := DeploymentConfig{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return nil, err, 208
	}
	return &data, nil, 0
}

/***
 * Save the deployments configuration file to disk
 * returns:  an error,  and error code
 */
func saveDeploymentsConfigFile() (error, int) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return err, 207
	}
	data := DeploymentConfig{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return err, 208
	}
	return nil, 0
}

/***
 * FindTargetDeployment
 * returns:  The single active deployment
 */
func FindTargetDeployment() *Deployment {
	data, err, errCode := loadDeploymentsConfigFile()
	errors.CheckErr(err, errCode, "Unable to process the deployments config file")
	activeID := data.Active
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(activeID, data.Deployments[i].Name) {
			targetDeployment := data.Deployments[i]
			targetDeployment.Url = strings.TrimSuffix(targetDeployment.Url, "/")
			return &targetDeployment
		}
	}
	return nil
}

/***
 * GetDeploymentConfig
 * returns:  The entire Deployment configuration contents
 */
func GetDeploymentsConfig() *DeploymentConfig {
	data, err, errCode := loadDeploymentsConfigFile()
	errors.CheckErr(err, errCode, "Unable to process the deployments config file")
	return data
}

/**
* Set active deployment
* If the deployment is unknown the command will fail with an error message
 */
func SetTargetDeployment(c *cli.Context) {
	newTargetName := c.String("name")
	data, err, errCode := loadDeploymentsConfigFile()
	errors.CheckErr(err, errCode, "Unable to process the deployments config file")
	foundName := ""

	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(newTargetName, data.Deployments[i].Name) {
			foundName = data.Deployments[i].Name
			break
		}
	}
	if foundName == "" {
		log.Fatal("Unable to change deployment. '" + newTargetName + "' has no matching configuration")
	}

	data.Active = foundName
	body, err := json.MarshalIndent(data, "", "\t")
	errors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	errors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

/**
 * AddDeploymentToList
 * Adds a new deployment to the deployment config
 */
func AddDeploymentToList(c *cli.Context) {
	name := strings.TrimSpace(strings.ToLower(c.String("name")))
	label := strings.TrimSpace(c.String("label"))
	url := c.String("url")
	if url != "" && len(strings.TrimSpace(url)) > 0 {
		url = strings.TrimSuffix(url, "/")
	}

	data, err, errCode := loadDeploymentsConfigFile()
	errors.CheckErr(err, errCode, "Unable to process the deployments config file")

	// check the name is not already in use
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(name, data.Deployments[i].Name) {
			log.Fatal("Deployment '" + name + "' already exists, to update:  first remove, then add")
		}
	}

	// create the new deployment
	newDeployment := Deployment{
		Name:  name,
		Label: label,
		Url:   url,
	}

	// append it to the list
	data.Deployments = append(data.Deployments, newDeployment)
	body, err := json.MarshalIndent(data, "", "\t")
	errors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	errors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

/**
 * RemoveDeploymentFromList
 * Removes a deployment from the list
 */
func RemoveDeploymentFromList(c *cli.Context) {
	name := c.String("name")
	if strings.EqualFold(name, "local") {
		log.Fatal("Local is a required deployment and can not be removed")
	}
	data, err, errCode := loadDeploymentsConfigFile()
	errors.CheckErr(err, errCode, "Unable to process the deployments config file")
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(name, data.Deployments[i].Name) {
			copy(data.Deployments[i:], data.Deployments[i+1:])
			data.Deployments = data.Deployments[:len(data.Deployments)-1]
		}
	}
	data.Active = "local"
	body, err := json.MarshalIndent(data, "", "\t")
	errors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	errors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

// Display the deployment details for the current target deployment
func ListTargetDeployment() {
	targetDeployment := FindTargetDeployment()
	if targetDeployment != nil {
		utils.PrettyPrintJSON(targetDeployment)
	} else {
		log.Fatal("Unable to find a matching target - set one now using the target command")
	}
}

// Display saved deployments
func ListDeployments() {
	deploymentConfig := GetDeploymentsConfig()
	if deploymentConfig != nil && deploymentConfig.Deployments != nil && len(deploymentConfig.Deployments) > 0 {
		utils.PrettyPrintJSON(deploymentConfig)
	} else {
		log.Fatal("Unable to any deployments - please run the reset command")
	}
}
