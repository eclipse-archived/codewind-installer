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
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

type noIgnoredPaths struct {
	Field1 []string `json:"field1"`
	Field2 string   `json:"field2"`
}

var testFolder, cwSettingsPopulatedPath,
	cwSettingsEmptyPath, cwSettingsNoIgnoredPathsObject string

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.RemoveAll(testFolder)
	os.Exit(code)
}

func setup() {
	// Create directories to write .cw-settings files to
	testFolder = "sync_test_folder_delete_me"
	cwSettingsPopulatedPath = path.Join(testFolder, "cwSettingsPopulated")
	cwSettingsEmptyPath = path.Join(testFolder, "cwSettingsEmpty")
	cwSettingsNoIgnoredPathsObject = path.Join(testFolder, "cwSettingsNoIgnoredPathsObject")
	populatedIgnoredPaths := CWSettings{
		IgnoredPaths: []string{
			"testfile",
			"anothertestfile",
		},
	}
	emptyIgnoredPaths := CWSettings{
		IgnoredPaths: []string{},
	}
	noIgnoredPaths := noIgnoredPaths{
		Field1: []string{"Something", "Else"},
		Field2: "Something Else",
	}

	// Create test directory to store .cw-settings files in
	os.Mkdir(testFolder, 0777)
	os.Mkdir(cwSettingsPopulatedPath, 0777)
	os.Mkdir(cwSettingsEmptyPath, 0777)
	os.Mkdir(cwSettingsNoIgnoredPathsObject, 0777)

	// Create .cw-settings files
	file, _ := json.Marshal(populatedIgnoredPaths)
	ioutil.WriteFile(path.Join(cwSettingsPopulatedPath, ".cw-settings"), file, 0644)
	file, _ = json.Marshal(emptyIgnoredPaths)
	ioutil.WriteFile(path.Join(cwSettingsEmptyPath, ".cw-settings"), file, 0644)
	file, _ = json.Marshal(noIgnoredPaths)
	ioutil.WriteFile(path.Join(cwSettingsNoIgnoredPathsObject, ".cw-settings"), file, 0644)
}

func TestIgnoreFileOrDirectory(t *testing.T) {
	tests := map[string]struct {
		name             string
		isDir            bool
		shouldBeIgnored  bool
		ignoredPathsList []string
	}{
		"success case: directory called node_modules should be ignored": {
			name:             "node_modules",
			isDir:            true,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
		},
		"success case: directory called load-test-23498729 should be ignored": {
			name:             "load-test-23498729",
			isDir:            true,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
		},
		"success case: directory called not-a-load-test-23498729 should not be ignored": {
			name:             "not-a-load-test-23498729",
			isDir:            true,
			shouldBeIgnored:  false,
			ignoredPathsList: []string{},
		},
		"success case: directory called noddy_modules should not be ignored": {
			name:             "noddy_modules",
			isDir:            true,
			shouldBeIgnored:  false,
			ignoredPathsList: []string{},
		},
		"success case: file called .DS_Store should be ignored": {
			name:             ".DS_Store",
			isDir:            false,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
		},
		"success case: file called something.swp should be ignored": {
			name:             "something.swp",
			isDir:            false,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
		},
		"success case: file called something.swpnot should not be ignored": {
			name:             "something.swpnot",
			isDir:            false,
			shouldBeIgnored:  false,
			ignoredPathsList: []string{},
		},
		"success case: file called node_modules should not be ignored": {
			name:             "node_modules",
			isDir:            false,
			shouldBeIgnored:  false,
			ignoredPathsList: []string{},
		},
		"success case: directory called .DS_Store should not be ignored": {
			name:             ".DS_Store",
			isDir:            true,
			shouldBeIgnored:  false,
			ignoredPathsList: []string{},
		},
		"success case: path containing noddy_modules should be ignored as it is in .cw-settings": {
			name:             "noddy_modules",
			isDir:            true,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{"noddy_modules"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fileIsIgnored := ignoreFileOrDirectory(test.name, test.isDir, test.ignoredPathsList)

			assert.IsType(t, test.shouldBeIgnored, fileIsIgnored, "Got: %s", fileIsIgnored)

			assert.Equal(t, test.shouldBeIgnored, fileIsIgnored, "fileIsIgnored was %b but should have been %b", fileIsIgnored, test.shouldBeIgnored)
		})
	}
}

func TestRetrieveIgnoredPathsList(t *testing.T) {
	tests := map[string]struct {
		projectPath           string
		shouldBeIgnored       []string
		shouldBeIgnoredLength int
	}{
		"success case: the returned ignoredPaths list should contain testfile and anothertestfile": {
			projectPath:           cwSettingsPopulatedPath,
			shouldBeIgnored:       []string{"testfile", "anothertestfile"},
			shouldBeIgnoredLength: 2,
		},
		"success case: the returned ignoredPaths list should be empty": {
			projectPath:           cwSettingsEmptyPath,
			shouldBeIgnored:       []string{},
			shouldBeIgnoredLength: 0,
		},
		"success case: calling on a path that doesn't exist should return nil with a length 0": {
			projectPath:           "pathdoesntexist",
			shouldBeIgnored:       nil,
			shouldBeIgnoredLength: 0,
		},
		"success case: calling on a path that does exist but isn't valid JSON": {
			projectPath:           "sync.go",
			shouldBeIgnored:       nil,
			shouldBeIgnoredLength: 0,
		},
		"success case: calling on a path that does exist and is valid JSON but doesn't contain ignoredPaths": {
			projectPath:           cwSettingsNoIgnoredPathsObject,
			shouldBeIgnored:       nil,
			shouldBeIgnoredLength: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cwSettings := retrieveCWSettings(test.projectPath)
			ignoredPathsList := cwSettings.IgnoredPaths

			assert.Equal(t, test.shouldBeIgnoredLength, len(ignoredPathsList), "Length of ignoredPathsList was %b but should have been %b", len(ignoredPathsList), test.shouldBeIgnoredLength)

			assert.Equal(t, test.shouldBeIgnored, ignoredPathsList, "ignoredPathsList was %b but should have been %b", ignoredPathsList, test.shouldBeIgnored)
		})
	}
}
