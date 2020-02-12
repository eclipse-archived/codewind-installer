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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/stretchr/testify/assert"
)

type (
	noIgnoredPaths struct {
		Field1 []string `json:"field1"`
		Field2 string   `json:"field2"`
	}

	testDirPaths struct {
		cwSettingsPopulated      string
		cwSettingsEmpty          string
		cwSettingsNoIgnoredPaths string
	}
)

func TestCompleteUpload(t *testing.T) {
	tests := map[string]struct {
		responseStatus int
	}{
		"200 status": {
			responseStatus: http.StatusOK,
		},
		"400 status": {
			responseStatus: http.StatusBadRequest,
		},
	}
	mockRequest := CompleteRequest{
		FileList: []string{"mock-file"},
	}
	// create empty res body, if nil resp.Body.close() will nil pointer panic
	var resBody interface{}
	r, _ := json.Marshal(resBody)
	body := ioutil.NopCloser(bytes.NewReader([]byte(r)))

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockClient := &security.ClientMockAuthenticate{StatusCode: test.responseStatus, Body: body}
			mockConnection := connections.Connection{ID: "local"}
			_, got := completeUpload(mockClient, "mockid", mockRequest, &mockConnection, "dummyURL")
			assert.Equal(t, got, test.responseStatus)
		})
	}
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
		"success case: file called file.iml should be ignored (IntelliJ metadata file, *.iml)": {
			name:             "file.iml",
			isDir:            false,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
		},
		"success case: path containing .idea should be ignored (IntelliJ metadata directory, .idea)": {
			name:             ".idea",
			isDir:            true,
			shouldBeIgnored:  true,
			ignoredPathsList: []string{},
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

func createTestPathsForIgnoredPathsTests(t *testing.T, testFolder string) testDirPaths {
	t.Helper()
	return testDirPaths{
		cwSettingsPopulated:      path.Join(testFolder, "cwSettingsPopulated"),
		cwSettingsEmpty:          path.Join(testFolder, "cwSettingsEmpty"),
		cwSettingsNoIgnoredPaths: path.Join(testFolder, "cwSettingsNoIgnoredPathsField"),
	}
}

func TestSyncFiles(t *testing.T) {
	testDir := "sync_test_folder_delete_me"
	os.Mkdir(testDir, 0777)
	body := ioutil.NopCloser(bytes.NewReader([]byte{}))
	mockClient := &security.ClientMockAuthenticate{StatusCode: http.StatusOK, Body: body}
	mockConnection := connections.Connection{ID: "local"}

	populatedIgnoredPaths := CWSettings{
		IgnoredPaths: []string{
			"testfile",
			"anothertestfile",
		},
	}
	cwSettingsFile, _ := json.Marshal(populatedIgnoredPaths)

	t.Run("success case - sync new empty file", func(t *testing.T) {
		mockProjectPath := path.Join(testDir, "empty-file")

		os.Mkdir(mockProjectPath, 0777)

		ioutil.WriteFile(path.Join(mockProjectPath, "test"), []byte{}, 0644)
		ioutil.WriteFile(path.Join(mockProjectPath, ".cw-settings"), cwSettingsFile, 0644)

		got, err := syncFiles(mockClient, mockProjectPath, "mockID", "dummyURL", 0, &mockConnection)
		if err != nil {
			t.Errorf("syncFiles() failed with error: %s", err)
		}

		expectedFileList := []string{".cw-settings", "test"}
		assert.Equal(t, expectedFileList, got.fileList)
	})

	t.Run("success case - sync new empty file, ignore IgnoredFile", func(t *testing.T) {
		mockProjectPath := path.Join(testDir, "empty-file")

		os.Mkdir(mockProjectPath, 0777)

		ioutil.WriteFile(path.Join(mockProjectPath, "test"), []byte{}, 0644)
		ioutil.WriteFile(path.Join(mockProjectPath, "testfile"), []byte{}, 0644)
		ioutil.WriteFile(path.Join(mockProjectPath, ".cw-settings"), cwSettingsFile, 0644)

		got, err := syncFiles(mockClient, mockProjectPath, "mockID", "dummyURL", 0, &mockConnection)
		if err != nil {
			t.Errorf("syncFiles() failed with error: %s", err)
		}

		expectedFileList := []string{".cw-settings", "test"}
		assert.Equal(t, expectedFileList, got.fileList)
	})

	t.Run("success case - sync new empty dir", func(t *testing.T) {
		mockProjectPath := path.Join(testDir, "new-dir")
		newDirPath := path.Join(mockProjectPath, "nested-dir")

		os.Mkdir(mockProjectPath, 0777)
		os.Mkdir(newDirPath, 0777)

		ioutil.WriteFile(path.Join(newDirPath, "test"), []byte{}, 0644)
		ioutil.WriteFile(path.Join(mockProjectPath, ".cw-settings"), cwSettingsFile, 0644)

		got, err := syncFiles(mockClient, mockProjectPath, "mockID", "dummyURL", 0, &mockConnection)
		if err != nil {
			t.Errorf("syncFiles() failed with error: %s", err)
		}

		expectedFileList := []string{".cw-settings", "nested-dir/test"}
		expectedDirList := []string{"nested-dir"}
		assert.Equal(t, got.fileList, expectedFileList)
		assert.Equal(t, got.directoryList, expectedDirList)
	})

	t.Run("success case - create 2 files, modify 1, only 1 added to modified list", func(t *testing.T) {
		mockProjectPath := path.Join(testDir, "modified-file")
		newDirPath := path.Join(mockProjectPath, "nested-dir")

		os.Mkdir(mockProjectPath, 0777)
		os.Mkdir(newDirPath, 0777)

		modTestPath := path.Join(newDirPath, "testmod")
		noModTestPath := path.Join(newDirPath, "testnomod")

		ioutil.WriteFile(path.Join(mockProjectPath, ".cw-settings"), cwSettingsFile, 0644)
		ioutil.WriteFile(modTestPath, []byte{}, 0644)
		ioutil.WriteFile(noModTestPath, []byte{}, 0644)

		file, _ := os.Stat(noModTestPath)
		modifiedTime := file.ModTime().UnixNano() / 1000000
		newContent := []byte("I have changed!")

		// wait for a second, so file modification time is greater than the lastSync time in syncFiles parameters
		time.Sleep(1 * time.Second)
		ioutil.WriteFile(modTestPath, newContent, 0644)

		got, _ := syncFiles(mockClient, mockProjectPath, "mockID", "dummyURL", modifiedTime, &mockConnection)

		expectedFileList := []string{".cw-settings", "nested-dir/testmod", "nested-dir/testnomod"}
		expectedDirList := []string{"nested-dir"}
		expectedModList := []string{"nested-dir/testmod"}
		assert.Equal(t, got.fileList, expectedFileList)
		assert.Equal(t, got.directoryList, expectedDirList)
		assert.Equal(t, got.modifiedList, expectedModList)
	})

	cleanupTestFolder(t, testDir)
}
func TestRetrieveIgnoredPathsList(t *testing.T) {
	testFolder := "sync_test_folder_delete_me"
	createTestDirPaths := createTestPathsForIgnoredPathsTests(t, testFolder)
	setupIgnoredPathsTests(t, testFolder, createTestDirPaths)

	tests := map[string]struct {
		projectPath           string
		shouldBeIgnored       []string
		shouldBeIgnoredLength int
	}{
		"success case: the returned ignoredPaths list should contain testfile and anothertestfile": {
			projectPath:           createTestDirPaths.cwSettingsPopulated,
			shouldBeIgnored:       []string{"testfile", "anothertestfile"},
			shouldBeIgnoredLength: 2,
		},
		"success case: the returned ignoredPaths list should be empty": {
			projectPath:           createTestDirPaths.cwSettingsEmpty,
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
			projectPath:           createTestDirPaths.cwSettingsNoIgnoredPaths,
			shouldBeIgnored:       nil,
			shouldBeIgnoredLength: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ignoredPathsList := retrieveIgnoredPathsList(test.projectPath)

			assert.Equal(t, test.shouldBeIgnoredLength, len(ignoredPathsList), "Length of ignoredPathsList was %b but should have been %b", len(ignoredPathsList), test.shouldBeIgnoredLength)

			assert.Equal(t, test.shouldBeIgnored, ignoredPathsList, "ignoredPathsList was %b but should have been %b", ignoredPathsList, test.shouldBeIgnored)
		})
	}
	cleanupTestFolder(t, testFolder)
}

func setupIgnoredPathsTests(t *testing.T, testFolder string, createPaths testDirPaths) {
	t.Helper()
	os.Mkdir(testFolder, 0777)

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

	os.Mkdir(createPaths.cwSettingsPopulated, 0777)
	file, _ := json.Marshal(populatedIgnoredPaths)
	ioutil.WriteFile(path.Join(createPaths.cwSettingsPopulated, ".cw-settings"), file, 0644)

	os.Mkdir(createPaths.cwSettingsEmpty, 0777)
	emptyFile, _ := json.Marshal(emptyIgnoredPaths)
	ioutil.WriteFile(path.Join(createPaths.cwSettingsEmpty, ".cw-settings"), emptyFile, 0644)

	os.Mkdir(createPaths.cwSettingsNoIgnoredPaths, 0777)
	NoIgnoredPathsfile, _ := json.Marshal(noIgnoredPaths)
	ioutil.WriteFile(path.Join(createPaths.cwSettingsNoIgnoredPaths, ".cw-settings"), NoIgnoredPathsfile, 0644)
}

func cleanupTestFolder(t *testing.T, testFolder string) {
	t.Helper()
	err := os.RemoveAll(testFolder)
	if err != nil {
		fmt.Println("Error removing test dir, you may need to remove manually")
	}
}
