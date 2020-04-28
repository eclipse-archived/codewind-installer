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

package apiroutes

import (
	"errors"
	"net/http"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/test"
	cwTest "github.com/eclipse/codewind-installer/pkg/test"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const numCodewindTemplates = 8

const numAppsodyTemplatesEnabled = 11

const numAppsodyTemplatesDisabled = 7

const numAppsodyTemplates = numAppsodyTemplatesEnabled + numAppsodyTemplatesDisabled

const numTemplatesEnabled = numCodewindTemplates + numAppsodyTemplatesEnabled

const numTemplates = numTemplatesEnabled + numAppsodyTemplatesDisabled

const URLOfExistingRepo = "https://raw.githubusercontent.com/codewind-resources/codewind-templates/master/devfiles/index.json"
const URLOfUnknownRepo = "https://raw.githubusercontent.com/UNKNOWN"
const URLOfUnknownRepo2 = "https://raw.githubusercontent.com/UNKNOWN_2"

func TestGetTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		inProjectStyle    string
		inShowEnabledOnly bool
		wantedType        []Template
		wantedLength      int
	}{
		"get templates of all styles": {
			inProjectStyle:    "",
			inShowEnabledOnly: false,
			wantedType:        []Template{},
			wantedLength:      numTemplates,
		},
		"filter templates by known style (Codewind)": {
			inProjectStyle:    "Codewind",
			inShowEnabledOnly: false,
			wantedType:        []Template{},
			wantedLength:      numCodewindTemplates,
		},
		"filter templates by known style (Appsody)": {
			inProjectStyle:    "Appsody",
			inShowEnabledOnly: false,
			wantedType:        []Template{},
			wantedLength:      numAppsodyTemplates,
		},
		"filter templates by unknown style": {
			inProjectStyle:    "unknownStyle",
			inShowEnabledOnly: false,
			wantedType:        []Template{},
			wantedLength:      0,
		},
		"filter templates by enabled templates": {
			inShowEnabledOnly: true,
			wantedType:        []Template{},
			wantedLength:      numTemplatesEnabled,
		},
		"filter templates by enabled templates of unknown style": {
			inProjectStyle:    "unknownStyle",
			inShowEnabledOnly: true,
			wantedType:        []Template{},
			wantedLength:      0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplates("local", test.inProjectStyle, test.inShowEnabledOnly)
			assert.IsType(t, test.wantedType, got)
			assert.True(t, len(got) >= test.wantedLength, "wanted len(got) >= %d but len(got) was %d", test.wantedLength, len(got))
			assert.Nil(t, err)
		})
	}
}

func TestGetTemplateStyles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		want      []string
		wantedErr error
	}{
		"success case": {
			want:      []string{"Appsody", "Codewind"},
			wantedErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateStyles("local")
			assert.Equal(t, test.want, got)
			assert.IsType(t, test.wantedErr, err)
		})
	}
}

func TestGetTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		wantedType   []utils.TemplateRepo
		wantedLength int
		wantedErr    error
	}{
		"success case": {
			wantedType:   []utils.TemplateRepo{},
			wantedLength: 3,
			wantedErr:    nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateRepos("local")
			assert.IsType(t, test.wantedType, got)
			assert.True(t, len(got) >= test.wantedLength, "wanted len(got) >= %d but len(got) was %d", test.wantedLength, len(got))
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestFailuresAddTemplateRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		inURL         string
		inDescription string
		wantedType    []utils.TemplateRepo
		wantedErr     error
	}{
		"fail case: add invalid URL": {
			inURL:         "invalidURL",
			inDescription: "invalidURL",
			wantedType:    nil,
			wantedErr:     errors.New("Error: 'invalidURL' is not a valid URL"),
		},
		"fail case: add duplicate URL": {
			inURL:         URLOfExistingRepo,
			inDescription: "example repository containing links to templates",
			wantedType:    nil,
			wantedErr:     errors.New("Error: Bad Request - " + URLOfExistingRepo + " is already a template repository"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := AddTemplateRepo("local", test.inURL, test.inDescription, "template-name", utils.GitCredentials{})
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestFailuresDeleteTemplateRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		inURL      string
		wantedType []utils.TemplateRepo
		wantedErr  error
	}{
		"fail case: remove invalid URL": {
			inURL:      "invalidURL",
			wantedType: nil,
			wantedErr:  errors.New("Error: 'invalidURL' is not a valid URL"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := DeleteTemplateRepo("local", test.inURL)
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestSuccessfulAddAndDeleteTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		skip             bool
		inURL            string
		inGitCredentials utils.GitCredentials
	}{
		"PublicGHDevFileURL": {
			inURL:            cwTest.PublicGHDevfileURL,
			inGitCredentials: utils.GitCredentials{},
		},
		"GHEDevfileURL": {
			skip:  !cwTest.UsingOwnGHECredentials,
			inURL: cwTest.GHEDevfileURL,
			inGitCredentials: utils.GitCredentials{
				Username: test.GHEUsername,
				Password: test.GHEPassword,
			},
		},
	}
	for name, test := range tests {
		if test.skip {
			t.Skip()
		}

		originalRepos, err := GetTemplateRepos("local")
		require.Nilf(t, err, "Error getting template repos: %s", err)
		originalNumRepos := len(originalRepos)

		t.Run(name, func(t *testing.T) {
			t.Run("Add template repo", func(t *testing.T) {
				wantedNumRepos := originalNumRepos + 1

				got, err := AddTemplateRepo("local", test.inURL, "description", "name", test.inGitCredentials)

				assert.IsType(t, []utils.TemplateRepo{}, got)
				assert.Equal(t, wantedNumRepos, len(got), "got: %v", got)
				require.Nil(t, err)
			})

			t.Run("Delete template repo", func(t *testing.T) {
				wantedNumRepos := originalNumRepos

				got, err := DeleteTemplateRepo("local", test.inURL)

				assert.IsType(t, []utils.TemplateRepo{}, got)
				assert.Equal(t, wantedNumRepos, len(got), "got: %v", got)
				assert.Nil(t, err)
			})
		})
		// This test block cleans up after itself
	}
}

func TestFailuresEnableTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		in         []string
		wantedType []utils.TemplateRepo
		wantedErr  error
	}{
		"nil repo URL": {
			in:         nil,
			wantedType: nil,
			wantedErr:  errors.New("Error: '[]' is not a valid URL"),
		},
		"invalid repo URL": {
			in:         []string{"invalidURL"},
			wantedType: nil,
			wantedErr:  errors.New("Error: 'invalidURL' is not a valid URL"),
		},
		"unknown repo URL": {
			in:         []string{URLOfUnknownRepo},
			wantedType: []utils.TemplateRepo{},
			wantedErr:  nil,
		},
		"multiple unknown repo URLs": {
			in:         []string{URLOfUnknownRepo, URLOfUnknownRepo2},
			wantedType: []utils.TemplateRepo{},
			wantedErr:  nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := EnableTemplateRepos("local", test.in)
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestFailuresDisableTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		in         []string
		wantedType []utils.TemplateRepo
		wantedErr  error
	}{
		"nil repo URL": {
			in:         nil,
			wantedType: nil,
			wantedErr:  errors.New("Error: '[]' is not a valid URL"),
		},
		"invalid repo URL": {
			in:         []string{"invalidURL"},
			wantedType: nil,
			wantedErr:  errors.New("Error: 'invalidURL' is not a valid URL"),
		},
		"unknown repo URL": {
			in:         []string{URLOfUnknownRepo},
			wantedType: []utils.TemplateRepo{},
			wantedErr:  nil,
		},
		"multiple unknown repo URLs": {
			in:         []string{URLOfUnknownRepo, URLOfUnknownRepo2},
			wantedType: []utils.TemplateRepo{},
			wantedErr:  nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := DisableTemplateRepos("local", test.in)
			assert.IsType(t, test.wantedType, got, "got: %v", got)
			assert.Equal(t, test.wantedErr, err)
		})
	}
}

func TestSuccessfulEnableAndDisableTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testRepoURL := URLOfExistingRepo

	t.Run("Successfully disable 1 template repo", func(t *testing.T) {
		got, err := DisableTemplateRepos("local", []string{testRepoURL})

		assert.IsType(t, []utils.TemplateRepo{}, got)
		assert.Nil(t, err)
		for _, repo := range got {
			if repo.URL == testRepoURL {
				assert.False(t, repo.Enabled)
			}
		}
	})

	t.Run("Successfully enable 1 template repo", func(t *testing.T) {
		got, err := EnableTemplateRepos("local", []string{testRepoURL})

		assert.IsType(t, []utils.TemplateRepo{}, got)
		assert.Nil(t, err)
		for _, repo := range got {
			if repo.URL == testRepoURL {
				assert.True(t, repo.Enabled)
			}
		}
	})

	// This test block cleans up after itself, assuming that the template repo tested was initially enabled. (This test block resets it to 'enabled')
}

func TestBatchPatchTemplateRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	tests := map[string]struct {
		in        []RepoOperation
		want      []SubResponseFromBatchOperation
		wantedErr error
	}{
		"enable 1 valid repo": {
			in: []RepoOperation{
				RepoOperation{
					Operation: "enable",
					URL:       URLOfExistingRepo,
					Value:     "true",
				},
			},
			want: []SubResponseFromBatchOperation{
				SubResponseFromBatchOperation{
					Status: 200,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfExistingRepo,
						Value:     "true",
					},
				},
			},
			wantedErr: nil,
		},
		"enable 1 unknown repo": {
			in: []RepoOperation{
				RepoOperation{
					Operation: "enable",
					URL:       URLOfUnknownRepo,
					Value:     "true",
				},
			},
			want: []SubResponseFromBatchOperation{
				SubResponseFromBatchOperation{
					Status: 404,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfUnknownRepo,
						Value:     "true",
					},
					Error: "Unknown repository URL",
				},
			},
			wantedErr: nil,
		},
		"disable 1 valid repo": {
			in: []RepoOperation{
				RepoOperation{
					Operation: "enable",
					URL:       URLOfExistingRepo,
					Value:     "false",
				},
			},
			want: []SubResponseFromBatchOperation{
				SubResponseFromBatchOperation{
					Status: 200,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfExistingRepo,
						Value:     "false",
					},
				},
			},
			wantedErr: nil,
		},
		"disable 1 unknown repo": {
			in: []RepoOperation{
				RepoOperation{
					Operation: "enable",
					URL:       URLOfUnknownRepo,
					Value:     "false",
				},
			},
			want: []SubResponseFromBatchOperation{
				SubResponseFromBatchOperation{
					Status: 404,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfUnknownRepo,
						Value:     "false",
					},
					Error: "Unknown repository URL",
				},
			},
			wantedErr: nil,
		},
		"enable/disable multiple repos": {
			in: []RepoOperation{
				RepoOperation{
					Operation: "enable",
					URL:       URLOfExistingRepo,
					Value:     "true",
				},
				RepoOperation{
					Operation: "enable",
					URL:       URLOfUnknownRepo,
					Value:     "false",
				},
			},
			want: []SubResponseFromBatchOperation{
				SubResponseFromBatchOperation{
					Status: 200,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfExistingRepo,
						Value:     "true",
					},
				},
				SubResponseFromBatchOperation{
					Status: 404,
					RequestedOperation: RepoOperation{
						Operation: "enable",
						URL:       URLOfUnknownRepo,
						Value:     "false",
					},
					Error: "Unknown repository URL",
				},
			},
			wantedErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := BatchPatchTemplateRepos("local", test.in)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantedErr, err)
		})
	}

	// This test block cleans up after itself, assuming that the template repo tested was initially enabled. (This test block resets it to 'enabled')
}

func TestHTTPRequestWithRetryOnLock(t *testing.T) {
	t.Run("Checks 423 is returned if the response StatusCode is always 423", func(t *testing.T) {
		mockClient := &MockResponse{StatusCode: http.StatusLocked}
		mockConnection := connections.Connection{ID: "local"}
		mockReq, _ := http.NewRequest("", "", nil)

		resp, httpSecError := HTTPRequestWithRetryOnLock(mockClient, mockReq, &mockConnection)
		expectedResp := &http.Response{
			StatusCode: http.StatusLocked,
		}
		assert.Equal(t, expectedResp, resp)
		assert.Nil(t, httpSecError)
	})
	t.Run("Checks that a non 423 StatusCode can be returned", func(t *testing.T) {
		mockClient := &MockResponse{StatusCode: http.StatusInternalServerError}
		mockConnection := connections.Connection{ID: "local"}
		mockReq, _ := http.NewRequest("", "", nil)

		resp, httpSecError := HTTPRequestWithRetryOnLock(mockClient, mockReq, &mockConnection)
		expectedResp := &http.Response{
			StatusCode: http.StatusInternalServerError,
		}
		assert.Equal(t, expectedResp, resp)
		assert.Nil(t, httpSecError)
	})
	t.Run("Checks secError is returned by not using a mocked client (URL doesn't exist)", func(t *testing.T) {
		mockConnection := connections.Connection{ID: "local"}
		req, _ := http.NewRequest("GET", "nonexistanturl", nil)

		resp, httpSecError := HTTPRequestWithRetryOnLock(http.DefaultClient, req, &mockConnection)
		assert.Nil(t, resp)
		assert.Equal(t, "tx_connection", httpSecError.Op)
	})
}
