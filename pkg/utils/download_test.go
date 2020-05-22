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

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDir = "./testDir"
const publicGHZipURL = "https://codeload.github.com/microclimate-dev2ops/nodeExpressTemplate/legacy.zip/master"
const publicGHTarGzURL = "https://github.com/appsody/stacks/releases/download/nodejs-v0.2.3/incubator.nodejs.templates.simple.tar.gz"

const privateGHRepoURL = "https://github.com/rwalle61/samplePrivateRepo"
const privateGHTarGzURL = "https://github.com/rwalle61/samplePrivateRepo/releases/download/v0.1.0/incubator.java-spring-boot2.v0.3.29.templates.default.tar.gz"
const privateGHUsername = "INSERT YOUR OWN: e.g. foo.bar@foobar.com"
const privateGHPassword = "INSERT YOUR OWN: e.g. 1234kljfdsjfaleru29348spodkfj445"
const usingOwnPrivateGHCredentials = false

const GHETarGzURL = "https://github.ibm.com/Richard-Waller/sampleGHERepo/releases/download/v0.1.0/incubator.nodejs.templates.simple.tar.gz"
const GHETarGzURLToNonExistentRepo = "https://github.ibm.com/Richard-Waller/nonExistentRepo/releases/download/v0.1.0/incubator.nodejs.templates.simple.tar.gz"
const GHETarGzURLToNonExistentRelease = "https://github.ibm.com/Richard-Waller/sampleGHERepo/releases/download/v0.nonExistentRelease.0/incubator.nodejs.templates.simple.tar.gz"
const GHETarGzURLToNonExistentReleaseAsset = "https://github.ibm.com/Richard-Waller/sampleGHERepo/releases/download/v0.1.0/nonExistentReleaseAsset.tar.gz"

func TestDownloadFromURLThenExtract(t *testing.T) {
	tests := map[string]struct {
		skip             bool
		inURL            string
		inDestination    string
		inGitCredentials *GitCredentials
		wantedType       error
		wantedErrMsg     string
		wantedNumFiles   int
	}{
		"success case: input good insecure Git URL": {
			inURL:          test.PublicGHRepoURL,
			inDestination:  filepath.Join(testDir, "git"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"success case: input good GHE URL and username-password": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            test.GHERepoURL,
			inDestination:    filepath.Join(testDir, "git"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       nil,
			wantedNumFiles:   7,
		},
		"success case: input good private Git URL and credentials": {
			skip:             !usingOwnPrivateGHCredentials,
			inURL:            privateGHRepoURL,
			inDestination:    filepath.Join(testDir, "git"),
			inGitCredentials: &GitCredentials{Username: privateGHUsername, Password: privateGHPassword},
			wantedType:       nil,
			wantedNumFiles:   15,
		},
		"success case: input good zip URL": {
			inURL:          publicGHZipURL,
			inDestination:  filepath.Join(testDir, "zip"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"success case: input good tar.gz URL": {
			inURL:          publicGHTarGzURL,
			inDestination:  filepath.Join(testDir, "targz"),
			wantedType:     nil,
			wantedNumFiles: 6,
		},
		"success case: input good private tar.gz URL and username-password": {
			skip:             !usingOwnPrivateGHCredentials,
			inURL:            privateGHTarGzURL,
			inDestination:    filepath.Join(testDir, "privateTarGz"),
			inGitCredentials: &GitCredentials{Username: privateGHUsername, Password: privateGHPassword},
			wantedType:       nil,
			wantedNumFiles:   5,
		},
		"success case: input good GHE tar.gz URL and username-password": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            GHETarGzURL,
			inDestination:    filepath.Join(testDir, "GHETarGz"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       nil,
			wantedNumFiles:   6,
		},
		"fail case: input bad URL": {
			inURL:          "bad URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedErrMsg:   "invalid URI",
			wantedNumFiles: 0,
		},
		"fail case: input relative URL": {
			inURL:          "/relative/URL",
			inDestination:  filepath.Join(testDir, "relativeURL"),
			wantedType:     errors.New(""),
			wantedErrMsg:   "URL must be absolute, but received relative URL /relative/URL",
			wantedNumFiles: 0,
		},
		"fail case: input good GHE repo URL but bad password": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            test.GHERepoURL,
			inDestination:    filepath.Join(testDir, "failCase"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: "bad password"},
			wantedType:       errors.New(""),
			wantedErrMsg:     "401 Unauthorized",
			wantedNumFiles:   0,
		},
		"fail case: input good GHE tar.gz URL and credentials but no matching repo found": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            GHETarGzURLToNonExistentRepo,
			inDestination:    filepath.Join(testDir, "failCase"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       errors.New(""),
			wantedErrMsg:     "GitHub responded with status code 404",
			wantedNumFiles:   0,
		},
		"fail case: input good GHE tar.gz URL and credentials but no matching releases found": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            GHETarGzURLToNonExistentRelease,
			inDestination:    filepath.Join(testDir, "failCase"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       errors.New(""),
			wantedErrMsg:     "Cannot find release ",
			wantedNumFiles:   0,
		},
		"fail case: input good GHE tar.gz URL and credentials but no matching assets found for release": {
			skip:             !test.UsingOwnGHECredentials,
			inURL:            GHETarGzURLToNonExistentReleaseAsset,
			inDestination:    filepath.Join(testDir, "failCase"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       errors.New(""),
			wantedErrMsg:     "Cannot find matching assets for release ",
			wantedNumFiles:   0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
			}

			os.RemoveAll(testDir)
			defer os.RemoveAll(testDir)

			got := DownloadFromURLThenExtract(test.inURL, test.inDestination, test.inGitCredentials)
			require.IsType(t, test.wantedType, got, "Got: %s", got)
			if test.wantedErrMsg != "" {
				assert.Contains(t, got.Error(), test.wantedErrMsg)
			}

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))
		})
	}
}

func getFilenames(files []os.FileInfo) []string {
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name()+",")
	}
	return filenames
}

func TestDownloadFromRepoURL(t *testing.T) {
	tests := map[string]struct {
		inURL            *url.URL
		inDestination    string
		inGitCredentials *GitCredentials
		wantedType       error
		wantedErrMsg     string
		wantedNumFiles   int
	}{
		"success case: input good path": {
			inURL:          toURL(test.PublicGHRepoURL),
			inDestination:  filepath.Join(testDir, "git"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"fail case: input URL that isn't to GH": {
			inURL:          toURL("https://www.google.com"),
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedErrMsg:   "URL must point to a GitHub repository",
			wantedNumFiles: 0,
		},
		"fail case: input GH URL that isn't to a repo": {
			inURL:          toURL("https://github.com/eclipse"),
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedErrMsg:   "URL must point to a GitHub repository",
			wantedNumFiles: 0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {
			got := DownloadFromRepoURL(test.inURL, test.inDestination, test.inGitCredentials)

			require.IsType(t, test.wantedType, got, "Got: %s", got)
			if test.wantedErrMsg != "" {
				assert.Contains(t, got.Error(), test.wantedErrMsg)
			}

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))

		})
		os.RemoveAll(testDir)
	}
}

func TestDownloadAndExtractZip(t *testing.T) {
	tests := map[string]struct {
		inURL          *url.URL
		inDestination  string
		wantedType     error
		wantedNumFiles int
	}{
		"success case: input good path": {
			inURL:          toURL(publicGHZipURL),
			inDestination:  filepath.Join(testDir, "zip"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"fail case: input bad URL": {
			inURL:          toURL("/bad/URL"),
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedNumFiles: 0,
		},
		"fail case: input URL that doesn't provide JSON": {
			inURL:          toURL("https://www.google.com/"),
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedNumFiles: 0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {
			got := DownloadAndExtractZip(test.inURL, test.inDestination)

			assert.IsType(t, test.wantedType, got, "Got: %s", got)

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(
				t,
				len(createdFiles) == test.wantedNumFiles,
				"len(createdFiles) was %d but should have been %d. createdFiles: %s",
				len(createdFiles),
				test.wantedNumFiles,
				getFilenames(createdFiles),
			)
		})
		os.RemoveAll(testDir)
		fmt.Println()
	}
}

func TestDownloadFromTarGzURL(t *testing.T) {
	tests := map[string]struct {
		inURL            *url.URL
		inDestination    string
		inGitCredentials *GitCredentials
		wantedType       error
		wantedErrMsg     string
		wantedNumFiles   int
	}{
		"success case: input good public GH URL": {
			inURL:          toURL(publicGHTarGzURL),
			inDestination:  "./testDir",
			wantedType:     nil,
			wantedNumFiles: 6,
		},
		"fail case: input bad URL": {
			inURL:          toURL("/bad/URL"),
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedErrMsg:   "unsupported protocol scheme",
			wantedNumFiles: 0,
		},
		"fail case: input Git credentials with URL that isn't GitHub": {
			inURL:            toURL("https://www.google.com/"),
			inDestination:    filepath.Join(testDir, "badURL"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       errors.New(""),
			wantedErrMsg:     "URL must point to a GitHub repository release asset",
			wantedNumFiles:   0,
		},
		"fail case: input Git credentials with GHE URL that isn't to a release asset": {
			inURL:            toURL(test.GHERepoURL),
			inDestination:    filepath.Join(testDir, "badURL"),
			inGitCredentials: &GitCredentials{Username: test.GHEUsername, Password: test.GHEPassword},
			wantedType:       errors.New(""),
			wantedErrMsg:     "URL must point to a GitHub repository release asset",
			wantedNumFiles:   0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {

			got := DownloadFromTarGzURL(test.inURL, test.inDestination, test.inGitCredentials)

			require.IsType(t, test.wantedType, got, "Got: %s", got)
			if test.wantedErrMsg != "" {
				assert.Contains(t, got.Error(), test.wantedErrMsg)
			}

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(
				t,
				len(createdFiles) == test.wantedNumFiles,
				"len(createdFiles) was %d but should have been %d. createdFiles: %s",
				len(createdFiles),
				test.wantedNumFiles,
				getFilenames(createdFiles),
			)
		})
		os.RemoveAll(testDir)
		fmt.Println()
	}
}

func TestDownloadFile(t *testing.T) {
	t.Run("fail case - response status code is not 200", func(t *testing.T) {
		testURL := "https://github.com/nonexistentrepo"
		err := DownloadFile(toURL(testURL), testDir)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "File download failed for "+testURL+", status code ")
	})
}

func TestIsTarGzURL(t *testing.T) {
	tests := map[string]struct {
		in   *url.URL
		want bool
	}{
		"success case": {
			in:   toURL(publicGHTarGzURL),
			want: true,
		},
		"fail case: git repo URL": {
			in:   toURL(test.PublicGHRepoURL),
			want: false,
		},
		"fail case: zip URL": {
			in:   toURL(publicGHZipURL),
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsTarGzURL(test.in)
			assert.Equal(t, got, test.want)
		})
	}
}

func toURL(inURL string) *url.URL {
	URL, _ := url.ParseRequestURI(inURL)
	return URL
}
