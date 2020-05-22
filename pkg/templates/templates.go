package templates

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

// AddTemplateRepo adds the provided template repo to PFE and
// stores provided gitCredentials to the keyring
func AddTemplateRepo(conID, URL, description, name string, gitCredentials *utils.GitCredentials) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}

	repos, addErr := apiroutes.AddTemplateRepoToPFE(conID, URL, description, name, gitCredentials)
	if addErr != nil {
		return nil, addErr
	}

	keyringErr := storeGitCredentialsInKeyring(URL, repos, conID, gitCredentials)
	if keyringErr != nil {
		return nil, keyringErr
	}
	return repos, nil
}

func storeGitCredentialsInKeyring(
	URL string,
	repos []utils.TemplateRepo,
	conID string,
	gitCredentials *utils.GitCredentials,
) error {
	var IDOfAddedRepo string
	for _, repo := range repos {
		if repo.ID != "" && repo.URL == URL {
			IDOfAddedRepo = repo.ID
		}
	}
	marshalledGitCredentials, marshallErr := json.Marshal(gitCredentials)
	if marshallErr != nil {
		return marshallErr
	}
	keyringErr := security.StoreSecretInKeyring(
		conID,
		"gitcredentials-"+IDOfAddedRepo,
		string(marshalledGitCredentials),
	)
	if keyringErr != nil {
		return keyringErr
	}
	return nil
}

// DeleteTemplateRepo deletes a template repo from PFE and
// deletes any matching credentials from the keyring and
// returns the new list of existing repos
func DeleteTemplateRepo(conID, URL string) ([]utils.TemplateRepo, error) {
	if _, err := url.ParseRequestURI(URL); err != nil {
		return nil, fmt.Errorf("Error: '%s' is not a valid URL", URL)
	}
	sourceID, findErr := findTemplateRepoSourceID(conID, URL)
	if findErr != nil {
		return nil, findErr
	}
	if sourceID != "" {
		keyringErr := security.DeleteSecretFromKeyring(conID, "gitcredentials-"+sourceID)
		if keyringErr != nil {
			return nil, keyringErr
		}
	}
	return apiroutes.DeleteTemplateRepoFromPFE(conID, URL)
}

// findTemplateSourceID matches a template URL to its SourceID if it has one
func findTemplateSourceID(conID, templateURL string) (string, error) {
	templates, err := apiroutes.GetTemplates(conID, "", false)
	if err != nil {
		return "", err
	}
	for _, template := range templates {
		if template.SourceID != "" && template.URL == templateURL {
			return template.SourceID, nil
		}
	}
	return "", nil
}

// findTemplateRepoSourceID matches a template repo URL to its SourceID if it has one
func findTemplateRepoSourceID(conID, repoURL string) (string, error) {
	repos, err := apiroutes.GetTemplateRepos(conID)
	if err != nil {
		return "", err
	}
	for _, repo := range repos {
		if repo.ID != "" && repo.URL == repoURL {
			return repo.ID, nil
		}
	}
	return "", nil
}

// GetGitCredentialsFromKeychain gets GitHub credentials for a template from the keychain
func GetGitCredentialsFromKeychain(conID, templateURL string) (utils.GitCredentials, error) {
	gitCredentials := utils.GitCredentials{}
	sourceID, findErr := findTemplateSourceID(conID, templateURL)
	if findErr != nil {
		return utils.GitCredentials{}, findErr
	}
	if sourceID != "" {
		savedGitCredentials, err := security.GetSecretFromKeyring(conID, "gitcredentials-"+sourceID)
		if err != nil {
			return utils.GitCredentials{}, err
		}

		unmarshalErr := json.Unmarshal([]byte(savedGitCredentials), &gitCredentials)
		if unmarshalErr != nil {
			return utils.GitCredentials{}, err
		}
	}
	return gitCredentials, nil
}
