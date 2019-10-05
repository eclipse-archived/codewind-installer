package deployments

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

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

// InitConfigFileIfRequired : Check the config file exist, if it does not then create a new default configuration
func InitConfigFileIfRequired() *DepError {
	_, err := os.Stat(getDeploymentConfigFilename())
	if os.IsNotExist(err) {
		os.MkdirAll(getDeploymentConfigDir(), 0777)
		return ResetDeploymentsFile()
	}
	return applySchemaUpdates()
}

// ResetDeploymentsFile : Creates a new / overwrites deployment config file with a default single local Codewind deployment
func ResetDeploymentsFile() *DepError {
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
	if err != nil {
		return &DepError{errOpFileParse, err, err.Error()}
	}

	err = ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	if err != nil {
		return &DepError{errOpFileWrite, err, err.Error()}
	}
	return nil
}

// FindTargetDeployment : Returns the single active deployment
func FindTargetDeployment() (*Deployment, *DepError) {
	data, depError := loadDeploymentsConfigFile()
	if depError != nil {
		return nil, depError
	}

	activeID := data.Active
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(activeID, data.Deployments[i].ID) {
			targetDeployment := data.Deployments[i]
			targetDeployment.URL = strings.TrimSuffix(targetDeployment.URL, "/")
			targetDeployment.AuthURL = strings.TrimSuffix(targetDeployment.AuthURL, "/")
			return &targetDeployment, nil
		}
	}

	err := errors.New(errTargetNotFound)
	return nil, &DepError{errOpNotFound, err, err.Error()}
}

// GetDeploymentsConfig : Retrieves and returns the entire Deployment configuration contents
func GetDeploymentsConfig() (*DeploymentConfig, *DepError) {
	data, depErr := loadDeploymentsConfigFile()
	if depErr != nil {
		return nil, depErr
	}
	return data, nil
}

// SetTargetDeployment : If the deployment is unknown return an error
func SetTargetDeployment(c *cli.Context) *DepError {
	newTargetName := c.String("id")
	data, depErr := loadDeploymentsConfigFile()
	if depErr != nil {
		return depErr
	}
	foundID := ""
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(newTargetName, data.Deployments[i].ID) {
			foundID = data.Deployments[i].ID
			break
		}
	}
	if foundID == "" {
		err := errors.New(errTargetNotFound)
		return &DepError{errOpNotFound, err, err.Error()}
	}
	data.Active = foundID
	body, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return &DepError{errOpFileParse, err, err.Error()}
	}
	err = ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	if err != nil {
		return &DepError{errOpFileWrite, err, err.Error()}
	}
	return nil
}

// AddDeploymentToList : adds a new deployment to the deployment config
func AddDeploymentToList(c *cli.Context) *DepError {
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

	data, depErr := loadDeploymentsConfigFile()
	if depErr != nil {
		return depErr
	}

	// check the name is not already in use
	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(id, data.Deployments[i].ID) {
			depError := errors.New("Deployment '" + id + "' already exists, to update:  first remove, then add")
			return &DepError{errOpConflict, depError, depError.Error()}
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
	if err != nil {
		return &DepError{errOpFileParse, err, err.Error()}
	}

	err = ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	if err != nil {
		return &DepError{errOpFileWrite, err, err.Error()}
	}
	return nil
}

// RemoveDeploymentFromList : Removes the stored entry
func RemoveDeploymentFromList(c *cli.Context) *DepError {
	id := c.String("id")
	if strings.EqualFold(id, "local") {
		depError := errors.New("Local is a required deployment and must not be removed")
		return &DepError{errOpProtected, depError, depError.Error()}
	}
	data, depErr := loadDeploymentsConfigFile()
	if depErr != nil {
		return depErr
	}

	for i := 0; i < len(data.Deployments); i++ {
		if strings.EqualFold(id, data.Deployments[i].ID) {
			copy(data.Deployments[i:], data.Deployments[i+1:])
			data.Deployments = data.Deployments[:len(data.Deployments)-1]
		}
	}
	data.Active = "local"
	body, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return &DepError{errOpFileParse, err, err.Error()}
	}

	err = ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	if err != nil {
		return &DepError{errOpFileWrite, err, err.Error()}
	}
	return nil
}

// GetTargetDeployment : Retrieve the deployment details for the current target deployment
func GetTargetDeployment() (*Deployment, *DepError) {
	targetDeployment, depErr := FindTargetDeployment()
	if depErr != nil {
		return nil, depErr
	}
	if targetDeployment != nil {
		return targetDeployment, nil
	}
	depError := errors.New("Unable to find a matching target - set one now using the target command")
	return nil, &DepError{errOpNotFound, depError, depError.Error()}
}

// GetAllDeployments : Retrieve all saved deployments
func GetAllDeployments() (*DeploymentConfig, *DepError) {
	deploymentConfig, depErr := GetDeploymentsConfig()
	if depErr != nil {
		return nil, depErr
	}
	if deploymentConfig != nil && deploymentConfig.Deployments != nil && len(deploymentConfig.Deployments) > 0 {
		return deploymentConfig, nil
	}
	depError := errors.New("No Deployments found")
	return nil, &DepError{errOpNotFound, depError, depError.Error()}
}

// loadDeploymentsConfigFile : Load the deployments configuration file from disk
// and returns the contents of the file or an error
func loadDeploymentsConfigFile() (*DeploymentConfig, *DepError) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return nil, &DepError{errOpFileLoad, err, err.Error()}
	}
	data := DeploymentConfig{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return nil, &DepError{errOpFileParse, err, err.Error()}
	}
	return &data, nil
}

// saveDeploymentsConfigFile : Save the deployments configuration file to disk
// returns an error, and error code
func saveDeploymentsConfigFile(deploymentConfig *DeploymentConfig) *DepError {
	body, err := json.MarshalIndent(deploymentConfig, "", "\t")
	if err != nil {
		return &DepError{errOpFileParse, err, err.Error()}
	}
	depErr := ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
	if depErr != nil {
		return &DepError{errOpFileWrite, depErr, depErr.Error()}
	}
	return nil
}

// getDeploymentConfigDir : get directory path to the deployments file
func getDeploymentConfigDir() string {
	const GOOS string = runtime.GOOS
	homeDir := ""
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return path.Join(homeDir, ".codewind", "config")
}

// getDeploymentConfigFilename  : get full file path of deployments file
func getDeploymentConfigFilename() string {
	return path.Join(getDeploymentConfigDir(), "deployments.json")
}

func loadRawDeploymentsFile() ([]byte, *DepError) {
	file, err := ioutil.ReadFile(getDeploymentConfigFilename())
	if err != nil {
		return nil, &DepError{errOpFileLoad, err, err.Error()}
	}
	return file, nil
}

// applySchemaUpdates : update any existing entries to use the new schema design
func applySchemaUpdates() *DepError {

	loadedFile, depErr := loadDeploymentsConfigFile()
	if depErr != nil {
		return depErr
	}
	savedSchemaVersion := loadedFile.SchemaVersion

	// upgrade the schema if needed
	if savedSchemaVersion < deploymentsSchemaVersion {
		file, depErr := loadRawDeploymentsFile()
		if depErr != nil {
			return depErr
		}

		// apply schama updates from version 0 to version 1
		if savedSchemaVersion == 0 {

			// current config file
			deploymentConfig := DeploymentConfigV0{}

			// create new config structure
			newDeploymentConfig := DeploymentConfigV1{}

			err := json.Unmarshal([]byte(file), &deploymentConfig)
			if err != nil {
				return &DepError{errOpFileParse, err, err.Error()}
			}

			newDeploymentConfig.Active = deploymentConfig.Active
			newDeploymentConfig.SchemaVersion = 1

			// copy deployments from old to new config
			originalDeploymentsV0 := deploymentConfig.Deployments
			for i := 0; i < len(originalDeploymentsV0); i++ {
				originalDeployment := originalDeploymentsV0[i]
				deploymentJSON, _ := json.Marshal(originalDeployment)
				var upgradedDeployment DeploymentV1
				json.Unmarshal(deploymentJSON, &upgradedDeployment)

				// rename 'name' field to 'id'
				upgradedDeployment.ID = originalDeployment.Name
				newDeploymentConfig.Deployments = append(newDeploymentConfig.Deployments, upgradedDeployment)
			}

			// schema has been updated
			body, err := json.MarshalIndent(newDeploymentConfig, "", "\t")
			if err != nil {
				return &DepError{errOpFileParse, err, err.Error()}
			}
			err = ioutil.WriteFile(getDeploymentConfigFilename(), body, 0644)
			if err != nil {
				return &DepError{errOpFileWrite, err, err.Error()}
			}
		}
	}
	return nil
}
