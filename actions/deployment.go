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
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	codewindErrors "github.com/eclipse/codewind-installer/errors"
	"github.com/eclipse/codewind-installer/utils"
	"github.com/eclipse/codewind-installer/utils/deployments"
	"github.com/urfave/cli"
)

// deploymentsSchemaVersion must be incremented when changing the Deployment Config or Deployment Entry
const deploymentsSchemaVersion = 1

// DeploymentConfig state and possible deployments
type DeploymentConfig struct {
	SchemaVersion int          `json:"schemaversion"`
	Active        string       `json:"active"`
	Deployments   []Deployment `json:"deployments"`
}

// Deployment entry
type Deployment struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	AuthURL  string `json:"auth"`
	Realm    string `json:"realm"`
	ClientID string `json:"clientid"`
}

// InitDeploymentConfigIfRequired : Check the config file exist, if it does not then create a new default configuration
func InitDeploymentConfigIfRequired() {
	_, err := os.Stat(getDeploymentConfigFilename())
	if os.IsNotExist(err) {
		os.MkdirAll(getDeploymentConfigPath(), 0777)
		ResetDeploymentsFile()
	} else {
		err = applySchemaUpdates()
	}
}

// ResetDeploymentsFile : Creates a new / overwrites deployment config file with a default single local Codewind deployment
func ResetDeploymentsFile() {
	// create the default local deployment
	initialConfig := DeploymentConfig{
		Active:        "local",
		SchemaVersion: deploymentsSchemaVersion,
		Deployments: []Deployment{
			Deployment{
				ID:       "local",
				Label:    "Codewind local deployment",
				URL:      "",
				AuthURL:  "",
				Realm:    "",
				ClientID: "",
			},
		},
	}
	body, err := json.MarshalIndent(initialConfig, "", "\t")
	codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

// FindTargetDeployment : Returns the single active deployment
func FindTargetDeployment() (*Deployment, error) {
	data, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to process the deployments config file")
	activeID := data.Active
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(activeID, data.Deployments[i].ID) {
			targetDeployment := data.Deployments[i]
			targetDeployment.URL = strings.TrimSuffix(targetDeployment.URL, "/")
			targetDeployment.AuthURL = strings.TrimSuffix(targetDeployment.AuthURL, "/")
			return &targetDeployment, nil
		}
	}
	depError := errors.New(deployments.TextTargetNotFound)
	return nil, &deployments.DepError{deployments.ErrOpSchema, depError, depError.Error()}
}

// GetDeploymentsConfig : Retrieves and returns the entire Deployment configuration contents
func GetDeploymentsConfig() *DeploymentConfig {
	data, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to process the deployments config file")
	return data
}

// SetTargetDeployment : If the deployment is unknown the command will return an error message
func SetTargetDeployment(c *cli.Context) {
	newTargetName := c.String("id")
	data, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to process the deployments config file")
	foundID := ""

	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(newTargetName, data.Deployments[i].ID) {
			foundID = data.Deployments[i].ID
			break
		}
	}
	if foundID == "" {
		log.Fatal("Unable to change deployment. '" + newTargetName + "' has no matching configuration")
	}

	data.Active = foundID
	body, err := json.MarshalIndent(data, "", "\t")
	codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

// AddDeploymentToList : adds a new deployment to the deployment config
func AddDeploymentToList(c *cli.Context) {
	id := strings.TrimSpace(strings.ToLower(c.String("id")))
	label := strings.TrimSpace(c.String("label"))
	url := c.String("url")
	if url != "" && len(strings.TrimSpace(url)) > 0 {
		url = strings.TrimSuffix(url, "/")
	}
	auth := c.String("auth")
	if auth != "" && len(strings.TrimSpace(auth)) > 0 {
		auth = strings.TrimSuffix(auth, "/")
	}

	realm := strings.TrimSpace(c.String("realm"))
	clientID := strings.TrimSpace(c.String("clientid"))

	data, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to process the deployments config file")

	// check the name is not already in use
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(id, data.Deployments[i].ID) {
			log.Fatal("Deployment '" + id + "' already exists, to update:  first remove, then add")
		}
	}

	// create the new deployment
	newDeployment := Deployment{
		ID:       id,
		Label:    label,
		URL:      url,
		AuthURL:  auth,
		Realm:    realm,
		ClientID: clientID,
	}

	// append it to the list
	data.Deployments = append(data.Deployments, newDeployment)
	body, err := json.MarshalIndent(data, "", "\t")
	codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

// RemoveDeploymentFromList : Removes the stored entry
func RemoveDeploymentFromList(c *cli.Context) {
	id := c.String("id")
	if strings.EqualFold(id, "local") {
		log.Fatal("Local is a required deployment and can not be removed")
	}
	data, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to process the deployments config file")
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(id, data.Deployments[i].ID) {
			copy(data.Deployments[i:], data.Deployments[i+1:])
			data.Deployments = data.Deployments[:len(data.Deployments)-1]
		}
	}
	data.Active = "local"
	body, err := json.MarshalIndent(data, "", "\t")
	codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
}

// ListTargetDeployment : Display the deployment details for the current target deployment
func ListTargetDeployment() {
	targetDeployment, _ := FindTargetDeployment()
	if targetDeployment != nil {
		utils.PrettyPrintJSON(targetDeployment)
	} else {
		log.Fatal("Unable to find a matching target - set one now using the target command")
	}
}

// ListDeployments : Output all saved deployments
func ListDeployments() {
	deploymentConfig := GetDeploymentsConfig()
	if deploymentConfig != nil && deploymentConfig.Deployments != nil && len(deploymentConfig.Deployments) > 0 {
		utils.PrettyPrintJSON(deploymentConfig)
	} else {
		log.Fatal("Unable to any deployments - please run the reset command")
	}
}

// loadDeploymentsConfigFile : Load the deployments configuration file from disk
// and returns the contents of the file or an error
func loadDeploymentsConfigFile() (*DeploymentConfig, int, error) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return nil, 207, err
	}
	data := DeploymentConfig{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return nil, 208, err
	}
	return &data, 0, nil
}

// saveDeploymentsConfigFile : Save the deployments configuration file to disk
// returns an error, and error code
func saveDeploymentsConfigFile(deploymentConfig *DeploymentConfig) (int, error) {
	body, err := json.MarshalIndent(deploymentConfig, "", "\t")
	codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
	saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
	return 0, nil
}

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

func loadRawDeploymentsFile() ([]byte, int, error) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return nil, 207, err
	}
	return file, 0, nil
}

// update any existing entries to use the new schema design
func applySchemaUpdates() error {

	loadedFile, errCode, err := loadDeploymentsConfigFile()
	codewindErrors.CheckErr(err, errCode, "Unable to load the deployments config file")
	savedSchemaVersion := loadedFile.SchemaVersion

	// upgrade the schema if needed
	if savedSchemaVersion < deploymentsSchemaVersion {
		file, errCode, err := loadRawDeploymentsFile()
		codewindErrors.CheckErr(err, errCode, "Unable to read the deployments config file")

		// apply schama updates from version 0 to version 1
		if savedSchemaVersion == 0 {

			// current config file
			deploymentConfig := deployments.DeploymentConfigV0{}

			// create new config structure
			newDeploymentConfig := deployments.DeploymentConfigV1{}

			err = json.Unmarshal([]byte(file), &deploymentConfig)
			if err != nil {
				return err
			}

			newDeploymentConfig.Active = deploymentConfig.Active
			newDeploymentConfig.SchemaVersion = 1

			// copy deployments from old to new config
			originalDeploymentsV0 := deploymentConfig.Deployments
			for i := 0; i < len(originalDeploymentsV0); i++ {
				originalDeployment := originalDeploymentsV0[i]
				deploymentJSON, _ := json.Marshal(originalDeployment)
				var upgradedDeployment deployments.DeploymentV1
				json.Unmarshal(deploymentJSON, &upgradedDeployment)

				// rename 'name' field to 'id'
				upgradedDeployment.ID = originalDeployment.Name
				newDeploymentConfig.Deployments = append(newDeploymentConfig.Deployments, upgradedDeployment)
			}

			// schema has been updated
			body, err := json.MarshalIndent(newDeploymentConfig, "", "\t")
			codewindErrors.CheckErr(err, 208, "Unable to format deployments file")
			saveErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
			codewindErrors.CheckErr(saveErr, 203, "Unable to save the deployments config file")
		}

	}
	return nil
}
