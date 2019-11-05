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

	"github.com/stretchr/testify/assert"
)

var exampleGitURL = "https://github.com/microclimate-dev2ops/nodeExpressTemplate"
var exampleZipURL = "https://codeload.github.com/microclimate-dev2ops/nodeExpressTemplate/legacy.zip/master"
var exampleTarGzURL = "https://github.com/appsody/stacks/releases/download/nodejs-v0.2.3/incubator.nodejs.templates.simple.tar.gz"
var testDir = "./testDir"

func TestDownloadFromURLThenExtract(t *testing.T) {
	tests := map[string]struct {
		inURL          string
		inDestination  string
		wantedType     error
		wantedNumFiles int
	}{
		"success case: input good Git URL": {
			inURL:          exampleGitURL,
			inDestination:  filepath.Join(testDir, "git"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"success case: input good zip URL": {
			inURL:          exampleZipURL,
			inDestination:  filepath.Join(testDir, "zip"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"success case: input good tar.gz URL": {
			inURL:          exampleTarGzURL,
			inDestination:  filepath.Join(testDir, "targz"),
			wantedType:     nil,
			wantedNumFiles: 6,
		},
		"fail case: input bad URL": {
			inURL:          "bad URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedNumFiles: 0,
		},
		"fail case: input URL that doesn't return 200": {
			inURL:          "/bad/URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedNumFiles: 0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {
			got := DownloadFromURLThenExtract(test.inURL, test.inDestination)
			assert.IsType(t, test.wantedType, got, "Got: %s", got)

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))

		})
		os.RemoveAll(testDir)
		fmt.Println()
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
		inURL          string
		inDestination  string
		wantedType     error
		wantedNumFiles int
	}{
		"success case: input good path": {
			inURL:          exampleGitURL,
			inDestination:  filepath.Join(testDir, "git"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"fail case: input URL that doesn't return 200": {
			inURL:          "/bad/URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedNumFiles: 0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {
			got := DownloadFromRepoURL(test.inURL, test.inDestination)

			assert.IsType(t, test.wantedType, got, "Got: %s", got)

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))

		})
		os.RemoveAll(testDir)
		fmt.Println()
	}
}

func TestDownloadAndExtractZip(t *testing.T) {
	tests := map[string]struct {
		inURL          string
		inDestination  string
		wantedType     error
		wantedNumFiles int
	}{
		"success case: input good path": {
			inURL:          exampleZipURL,
			inDestination:  filepath.Join(testDir, "zip"),
			wantedType:     nil,
			wantedNumFiles: 17,
		},
		"fail case: input bad URL": {
			inURL:          "/bad/URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedNumFiles: 0,
		},
		"fail case: input URL that doesn't provide JSON": {
			inURL:          "https://www.google.com/",
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
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))

		})
		os.RemoveAll(testDir)
		fmt.Println()
	}
}

func TestDownloadFromTarGzURL(t *testing.T) {
	tests := map[string]struct {
		inURL          string
		inDestination  string
		wantedType     error
		wantedNumFiles int
	}{
		"success case: input good path": {
			inURL:          exampleTarGzURL,
			inDestination:  "./testDir",
			wantedType:     nil,
			wantedNumFiles: 6,
		},
		"fail case: input bad URL": {
			inURL:          "/bad/URL",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     new(url.Error),
			wantedNumFiles: 0,
		},
		"fail case: input URL that doesn't provide JSON": {
			inURL:          "https://www.google.com/",
			inDestination:  filepath.Join(testDir, "badURL"),
			wantedType:     errors.New(""),
			wantedNumFiles: 0,
		},
	}
	for name, test := range tests {
		os.RemoveAll(testDir)
		t.Run(name, func(t *testing.T) {

			got := DownloadFromTarGzURL(test.inURL, test.inDestination)

			assert.IsType(t, test.wantedType, got, "Got: %s", got)

			createdFiles, _ := ioutil.ReadDir(test.inDestination)
			assert.Truef(t, len(createdFiles) == test.wantedNumFiles, "len(createdFiles) was %d but should have been %d. createdFiles: %s", len(createdFiles), test.wantedNumFiles, getFilenames(createdFiles))

		})
		os.RemoveAll(testDir)
		fmt.Println()
	}
}

func TestIsTarGzURL(t *testing.T) {
	tests := map[string]struct {
		in   string
		want bool
	}{
		"success case": {
			in:   exampleTarGzURL,
			want: true,
		},
		"fail case: git repo URL": {
			in:   exampleGitURL,
			want: false,
		},
		"fail case: zip URL": {
			in:   exampleZipURL,
			want: false,
		},
		"fail case: other string": {
			in:   "not a targz",
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
