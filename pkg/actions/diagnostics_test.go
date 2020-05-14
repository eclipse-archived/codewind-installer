/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
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
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testDir = "./testDir/diagnostics"

type testStruct struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func clearAllDiagnostics() {
	app := cli.NewApp()
	flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
	flagSet.Bool("clean", true, "")
	context := cli.NewContext(app, flagSet, nil)
	DiagnosticsCommand(context)
}

func getMockDockerClient() (docker.DockerClient, *docker.DockerError) {
	return &docker.MockDockerClientWithCw{}, nil
}

// clearAllDiagnostics()
// app := cli.NewApp()
// flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
// flagSet.String("conid", "local", "")
// context := cli.NewContext(app, flagSet, nil)
// t.Run("local success case - no arguments specified", func(t *testing.T) {
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w
// 	DiagnosticsCommand(context)
// 	w.Close()
// 	out, _ := ioutil.ReadAll(r)
// 	os.Stdout = originalStdout
// 	fmt.Println("Spitting out output")
// 	fmt.Println(string(out))
// 	assert.DirExists(t, filepath.Join(homeDir, ".codewind", "diagnostics"))
//  })

func Test_warnDG(t *testing.T) {
	warning := "test_warn"
	description := "test warning description"
	expectedConsoleOutput := warning + ": " + description + "\n"
	expectedJSONOutput := dgWarning{WarningType: warning, WarningDesc: description}
	t.Run("warnDG - console", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		warnDG(warning, description)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))
	})
	t.Run("warnDG - json", func(t *testing.T) {
		dgWarningArray = []dgWarning{}
		printAsJSON = true
		warnDG(warning, description)
		assert.Equal(t, expectedJSONOutput, dgWarningArray[0])
	})
}

func Test_writeStreamToFile(t *testing.T) {
	testString := "Testing writeStreamToFile"
	testFileName := "twstf.txt"
	os.MkdirAll(testDir, 0755)
	diagnosticsDirName = testDir
	t.Run("writeStreamToFile - success", func(t *testing.T) {
		readcloser := ioutil.NopCloser(strings.NewReader(testString))
		wSTFerr := writeStreamToFile(readcloser, testFileName)
		if wSTFerr != nil {
			t.Error("Error encountered - " + wSTFerr.Error())
		}
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, testFileName))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, testString, string(contents))
	})
}

func Test_writeJSONStructToFile(t *testing.T) {
	testStructure := testStruct{Status: "OK", Message: "testing writeJSONStructToFile"}
	testFileName := "twjstf.txt"
	os.MkdirAll(testDir, 0755)
	diagnosticsDirName = testDir
	t.Run("writeJSONStructToFile - success", func(t *testing.T) {
		wJSTFerr := writeJSONStructToFile(testStructure, testFileName)
		if wJSTFerr != nil {
			t.Error("Error encountered - " + wJSTFerr.Error())
		}
		file, fileErr := ioutil.ReadFile(filepath.Join(testDir, testFileName))
		if fileErr != nil {
			t.Error("Error encountered - " + fileErr.Error())
		}
		readStructure := testStruct{}
		_ = json.Unmarshal([]byte(file), &readStructure)
		assert.Equal(t, readStructure, testStructure)
	})
}

func Test_copyCodewindWorkspace(t *testing.T) {
	t.Run("copyCodewindWorkspace - empty containerID", func(t *testing.T) {
		expectedErrorString := "Unable to find Codewind PFE container - could not get container ID"
		cCWerr := copyCodewindWorkspace("", getMockDockerClient)
		if cCWerr == nil {
			t.Error("Did not receive expected error")
		}
		assert.Equal(t, expectedErrorString, cCWerr.Error())
	})
	t.Run("copyCodewindWorkspace - success", func(t *testing.T) {
		diagnosticsLocalDirName = testDir
		cCWerr := copyCodewindWorkspace("test", getMockDockerClient)
		if cCWerr != nil {
			t.Error("Error encountered - " + cCWerr.Error())
		}
		// file output from mock client is empty, so no files to test for
	})
}

func Test_writeContainerLogToFile(t *testing.T) {
	conName := "test"
	t.Run("writeContainerLogToFile - empty containerID", func(t *testing.T) {
		expectedErrorString := "Unable to find " + conName + " container - could not get container ID"
		wCLTFerr := writeContainerLogToFile("", conName, getMockDockerClient)
		if wCLTFerr == nil {
			t.Error("Did not receive expected error")
		}
		assert.Equal(t, expectedErrorString, wCLTFerr.Error())
	})
	t.Run("writeContainerLogToFile - success", func(t *testing.T) {
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCLTFerr := writeContainerLogToFile("anything", conName, getMockDockerClient)
		if wCLTFerr != nil {
			t.Error("Error encountered - " + wCLTFerr.Error())
		}
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, conName+".log"))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, "", string(contents))
	})
}

func Test_writeContainerInspectToFile(t *testing.T) {
	conName := "test"
	t.Run("writeContainerInspectToFile - empty containerID", func(t *testing.T) {
		expectedErrorString := "Unable to find " + conName + " container - could not get container ID"
		wCITFerr := writeContainerInspectToFile("", conName, getMockDockerClient)
		if wCITFerr == nil {
			t.Error("Did not receive expected error")
		}
		assert.Equal(t, expectedErrorString, wCITFerr.Error())
	})
	t.Run("writeContainerInspectToFile - success", func(t *testing.T) {
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCITFerr := writeContainerLogToFile("anything", conName, getMockDockerClient)
		if wCITFerr != nil {
			t.Error("Error encountered - " + wCITFerr.Error())
		}
		// file output from mock client is empty, so no files to test for
	})
}

func Test_getContainerID(t *testing.T) {
	goodConName := "/codewind-pfe"
	expectedGoodResult := "pfe"
	t.Run("getContainerID - success", func(t *testing.T) {
		containerID := getContainerID(goodConName, getMockDockerClient)
		assert.Equal(t, expectedGoodResult, containerID)
	})
	//can't test for name not found without another client as mock client always returns a populated list
}

func Test_gatherCodewindVersions(t *testing.T) {
	localConID := "local"
	remoteConID := "remote"
	t.Run("gatherCodewindVersions - local success", func(t *testing.T) {
		diagnosticsLocalDirName = testDir
		gatherCodewindVersions(localConID)
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, "codewind.versions"))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Contains(t, string(contents), "CWCTL VERSION: ")
		assert.Contains(t, string(contents), "PFE VERSION: ")
		assert.Contains(t, string(contents), "PERFORMANCE VERSION: ")
		assert.NotContains(t, string(contents), "GATEKEEPER VERSION: ")
	})
	t.Run("gatherCodewindVersions - remote success", func(t *testing.T) {
		diagnosticsDirName = testDir
		remoteDirName := codewindPrefix + remoteConID
		os.MkdirAll(filepath.Join(testDir, remoteDirName), 0755)
		gatherCodewindVersions(remoteConID)
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, codewindPrefix+remoteConID, "codewind.versions"))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Contains(t, string(contents), "CWCTL VERSION: ")
		assert.Contains(t, string(contents), "PFE VERSION: ")
		assert.Contains(t, string(contents), "PERFORMANCE VERSION: ")
		assert.Contains(t, string(contents), "GATEKEEPER VERSION: ")
	})
}
