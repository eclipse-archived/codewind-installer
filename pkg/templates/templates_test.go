package templates

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/project"
	"github.com/eclipse/codewind-installer/pkg/security"
	cwTest "github.com/eclipse/codewind-installer/pkg/test"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDir = "./testDir"

func TestSuccessfulAddAndDeleteTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		skip             bool
		inURL            string
		inGitCredentials *utils.GitCredentials
	}{
		"public GH devfile URL": {
			inURL: cwTest.PublicGHDevfileURL,
		},
		"GHE devfile URL with GHE basic credentials": {
			skip:  !cwTest.UsingOwnGHECredentials,
			inURL: cwTest.GHEDevfileURL,
			inGitCredentials: &utils.GitCredentials{
				Username: cwTest.GHEUsername,
				Password: cwTest.GHEPassword,
			},
		},
		"GHE devfile URL with GHE personal access token": {
			skip:  !cwTest.UsingOwnGHECredentials,
			inURL: cwTest.GHEDevfileURL,
			inGitCredentials: &utils.GitCredentials{
				PersonalAccessToken: cwTest.GHEPersonalAccessToken,
			},
		},
	}
	for name, test := range tests {
		if test.skip {
			t.Skip()
		}

		t.Run(name, func(t *testing.T) {
			var IDOfAddedRepo string

			os.RemoveAll(testDir)
			defer os.RemoveAll(testDir)

			t.Run("Add template repo", func(t *testing.T) {
				got, err := AddTemplateRepo(cwTest.ConID, test.inURL, "description", "name", test.inGitCredentials)

				assert.IsType(t, []utils.TemplateRepo{}, got)
				require.Nil(t, err)

				if test.inGitCredentials != nil {
					for _, repo := range got {
						if repo.ID != "" && repo.URL == test.inURL {
							IDOfAddedRepo = repo.ID
						}
					}
					gitCredsString, keyringErr := security.GetSecretFromKeyring(cwTest.ConID, "gitcredentials-"+IDOfAddedRepo)
					assert.Nil(t, keyringErr)

					var gitCredentials *utils.GitCredentials
					unmarshalErr := json.Unmarshal([]byte(gitCredsString), &gitCredentials)
					assert.Nil(t, unmarshalErr)
					assert.Equal(t, test.inGitCredentials, gitCredentials)
				}
			})

			t.Run("Create project from template from added repo", func(t *testing.T) {
				templates, err := apiroutes.GetTemplates(cwTest.ConID, "", false)
				require.Nilf(t, err, "Error getting template repos: %s", err)

				var URLOfAddedTemplate string
				for _, template := range templates {
					if template.SourceID == IDOfAddedRepo {
						URLOfAddedTemplate = template.URL
					}
				}
				gitCredentials, err := GetGitCredentialsFromKeychain(cwTest.ConID, URLOfAddedTemplate)
				assert.Nil(t, err)
				assert.Equal(t, test.inGitCredentials, gitCredentials)

				result, err := project.DownloadTemplate(testDir, URLOfAddedTemplate, gitCredentials)
				assert.Nil(t, err)
				if result != nil {
					assert.Equal(t, result.Status, "success")
				}
			})

			t.Run("Delete template repo", func(t *testing.T) {
				got, err := DeleteTemplateRepo(cwTest.ConID, test.inURL)

				assert.IsType(t, []utils.TemplateRepo{}, got)
				assert.Nil(t, err)

				if test.inGitCredentials != nil {
					gitCredsString, err := security.GetSecretFromKeyring(cwTest.ConID, "gitcredentials-"+IDOfAddedRepo)
					assert.Equal(t, "", gitCredsString)
					require.NotNil(t, err)
					assert.Equal(t, "sec_keyring_secret_not_found", err.Op)
					assert.Contains(t, err.Desc, "not found in keyring")
				}
			})
		})
	}
}
