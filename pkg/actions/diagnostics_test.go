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
	"archive/zip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var testDir = filepath.Join(".", "testDir", "diagnostics")

type testStruct struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// mock docker clients

func getMockDockerClient() (docker.DockerClient, *docker.DockerError) {
	return &docker.MockDockerClientWithCw{}, nil
}

func getMockDockerErrorClient() (docker.DockerClient, *docker.DockerError) {
	return &docker.MockDockerErrorClient{}, nil
}

var dockerClientError = docker.DockerError{Op: "Error", Err: errors.New("error"), Desc: "test error"}

func getDockerClientError() (docker.DockerClient, *docker.DockerError) {
	return nil, &dockerClientError
}

// mock kubernetes pods & client

var mockPFEPod = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "codewind-pfe-somename",
		Namespace:   "default",
		Annotations: map[string]string{},
		Labels: map[string]string{
			"app":               "codewind-pfe",
			"codewindWorkspace": "",
		},
	},
}

var mockProjectPod = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "cw-myproj-something",
		Namespace:   "default",
		Annotations: map[string]string{},
	},
	Spec: v1.PodSpec{
		ServiceAccountName: "codewind-",
	},
}

var initialReplicas = int32(1)

var mockDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-deployment",
		Labels: map[string]string{
			"app":               "codewind-pfe",
			"codewindWorkspace": "",
		},
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &initialReplicas,
	},
}

var mockClientset = fake.NewSimpleClientset(mockPFEPod, mockProjectPod, mockDeployment)

func returnMockConnections() ([]connections.Connection, *connections.ConError) {
	return []connections.Connection{
		{
			ID:       "local",
			Label:    "Codewind local connection",
			URL:      "",
			AuthURL:  "",
			Realm:    "",
			ClientID: "",
			Username: "",
		},
		{
			ID:       "KA5GC1JF",
			Label:    "remote",
			URL:      "https://codewind-gatekeeper-something.apps.somewhere.com",
			AuthURL:  "https://codewind-keycloak-something.apps.somewhere.com",
			Realm:    "codewind",
			ClientID: "codewind-remote",
			Username: "developer",
		},
	}, nil
}

//unzip file needed as utils.UnZip is not a straight unzipper of zips
func unzipFile(t *testing.T, filePath, destination string) error {
	t.Log("Unzipping "+filePath+" to "+destination)
	zipReader, _ := zip.OpenReader(filePath)
	if zipReader == nil {
		t.Log("Error - zipreader is empty")
		return fmt.Errorf("file '%s' is empty", filePath)
	}

	os.MkdirAll(destination, 0755)
	for _, file := range zipReader.File {

		zippedFile, err := file.Open()
		if err != nil {
			t.Log("Unable to open zipped file " + file.Name))
			return errors.New("Unable to open zipped file")
		}

		extractedFilePath := filepath.Join(destination, file.Name)
		t.Log("extractedFilePath = "+extractedFilePath)

		if file.FileInfo().IsDir() {
			// For debug:
			// fmt.Println("Directory Created:", extractedFilePath)
			t.Log("file is dir")
			os.MkdirAll(extractedFilePath, file.Mode())
			zippedFile.Close()
		} else {
			// For debug:
			// fmt.Println("File extracted:", file.Name)
			t.Log("file is file")
			os.MkdirAll(filepath.Dir(extractedFilePath), file.Mode())
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
			    zippedFile.Close()
				return errors.New("unable to open file " + file.Name)
			}

			io.Copy(outputFile, zippedFile)
			outputFile.Close()
			zippedFile.Close()
		}
	}
	zipReader.Close()
	return nil
}

// function to tell if a directory is empty
func isEmptyDir(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	info, staterr := f.Stat()
	if staterr != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}

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
		assert.Equal(t, testStructure, readStructure)
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
	t.Run("copyCodewindWorkspace - docker client error", func(t *testing.T) {
		diagnosticsLocalDirName = testDir
		cCWerr := copyCodewindWorkspace("test", getDockerClientError)
		if cCWerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &dockerClientError, cCWerr)
	})
	t.Run("copyCodewindWorkspace - error from docker client", func(t *testing.T) {
		diagnosticsLocalDirName = testDir
		expectedError := docker.DockerError{Op: docker.ErrOpContainerError, Err: docker.ErrCopyFromContainer, Desc: docker.ErrCopyFromContainer.Error()}
		cCWerr := copyCodewindWorkspace("test", getMockDockerErrorClient)
		if cCWerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &expectedError, cCWerr)
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
	t.Run("writeContainerLogToFile - docker client error", func(t *testing.T) {
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCLTFerr := writeContainerLogToFile("anything", conName, getDockerClientError)
		if wCLTFerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &dockerClientError, wCLTFerr)
	})
	t.Run("writeContainerLogToFile - error from docker client", func(t *testing.T) {
		expectedError := docker.DockerError{Op: docker.ErrOpContainerLogs, Err: docker.ErrContainerLogs, Desc: docker.ErrContainerLogs.Error()}
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCLTFerr := writeContainerLogToFile("anything", conName, getMockDockerErrorClient)
		if wCLTFerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &expectedError, wCLTFerr)
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
	t.Run("writeContainerLogToFile - docker client error", func(t *testing.T) {
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCITFerr := writeContainerInspectToFile("anything", conName, getDockerClientError)
		if wCITFerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &dockerClientError, wCITFerr)
	})
	t.Run("writeContainerInspectToFile - error from docker client", func(t *testing.T) {
		expectedError := docker.DockerError{Op: docker.ErrOpContainerInspect, Err: docker.ErrContainerInspect, Desc: docker.ErrContainerInspect.Error()}
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCITFerr := writeContainerInspectToFile("anything", conName, getMockDockerErrorClient)
		if wCITFerr == nil {
			t.Error("Expected error not encountered")
		}
		assert.Equal(t, &expectedError, wCITFerr)
	})
	t.Run("writeContainerInspectToFile - success", func(t *testing.T) {
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		wCITFerr := writeContainerInspectToFile("anything", conName, getMockDockerClient)
		if wCITFerr != nil {
			t.Error("Error encountered - " + wCITFerr.Error())
		}
		file, fileErr := ioutil.ReadFile(filepath.Join(testDir, conName+".inspect"))
		if fileErr != nil {
			t.Error("Error encountered - " + fileErr.Error())
		}
		readStructure := types.ContainerJSON{}
		_ = json.Unmarshal([]byte(file), &readStructure)
		// fakeclient returns a JSON object where the only thing it sets is AutoRemove
		assert.Equal(t, true, readStructure.ContainerJSONBase.HostConfig.AutoRemove)
	})
}

func Test_getContainerID(t *testing.T) {
	goodConName := "/codewind-pfe"
	expectedGoodResult := "pfe"
	expectedBadResult := ""
	t.Run("writeContainerLogToFile - docker client error", func(t *testing.T) {
		containerID := getContainerID(goodConName, getDockerClientError)
		assert.Equal(t, expectedBadResult, containerID)
	})
	t.Run("writeContainerInspectToFile - error from docker client", func(t *testing.T) {
		containerID := getContainerID(goodConName, getMockDockerErrorClient)
		assert.Equal(t, expectedBadResult, containerID)
	})
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
		diagnosticsDirName = testDir
		os.MkdirAll(filepath.Join(testDir, localConID), 0755)
		gatherCodewindVersions(localConID)
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, localConID, "codewind.versions"))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Contains(t, string(contents), "CWCTL VERSION: ")
		assert.Contains(t, string(contents), "PFE VERSION: ")
		assert.Contains(t, string(contents), "PERFORMANCE VERSION: ")
		assert.Contains(t, string(contents), "DOCKER CLIENT VERSION: ")
		assert.Contains(t, string(contents), "DOCKER SERVER VERSION: ")
		assert.NotContains(t, string(contents), "GATEKEEPER VERSION: ")
		os.RemoveAll(filepath.Join(testDir, localConID))
	})
	t.Run("gatherCodewindVersions - remote success", func(t *testing.T) {
		diagnosticsDirName = testDir
		os.MkdirAll(filepath.Join(testDir, remoteConID), 0755)
		gatherCodewindVersions(remoteConID)
		contents, rfErr := ioutil.ReadFile(filepath.Join(testDir, remoteConID, "codewind.versions"))
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Contains(t, string(contents), "CWCTL VERSION: ")
		assert.Contains(t, string(contents), "PFE VERSION: ")
		assert.Contains(t, string(contents), "PERFORMANCE VERSION: ")
		assert.Contains(t, string(contents), "GATEKEEPER VERSION: ")
		assert.NotContains(t, string(contents), "DOCKER CLIENT VERSION: ")
		assert.NotContains(t, string(contents), "DOCKER SERVER VERSION: ")
	})
}

func Test_createZipAndRemoveCollectedFiles(t *testing.T) {
	t.Run("createZipAndRemoveCollectedFiles - success", func(t *testing.T) {
		diagnosticsDirName = testDir
		testDgDir, _ := os.Open(testDir)
		testfilenames, _ := testDgDir.Readdirnames(-1)
		// end zip won't contain empty dirs, so remove those from expected output
		expectedFilenames := []string{}
		for _, fileName := range testfilenames {
			if !isEmptyDir(filepath.Join(testDir, fileName)) {
				expectedFilenames = append(expectedFilenames, fileName)
			}
		}
		testDgDir.Close()
		fmt.Println(testfilenames)
		nowTime = time.Now().Format("20060102150405")
		expectedZipFileName := "diagnostics." + nowTime + ".zip"
		expectedZipFilePath := filepath.Join(diagnosticsDirName, expectedZipFileName)
		createZipAndRemoveCollectedFiles()
		assert.FileExists(t, expectedZipFilePath, "Unable to find "+expectedZipFileName)
		// Jenkins test - what's in the zip file
		t.Log("Contents of " + expectedZipFilePath)
		read, _ := zip.OpenReader(expectedZipFilePath)
		for _, file := range read.File {
			t.Log(file.Name)
		}
		read.Close()
		unzipErr := unzipFile(t, expectedZipFilePath, testDir)
		if unzipErr != nil (
			t.Error("Problems encountered unzipping "+expectedZipFilePath+": "+unzipErr.Error())
		)
		testDgAfterDir, _ := os.Open(testDir)
		testfilenamesAfter, _ := testDgAfterDir.Readdirnames(-1)
		testDgAfterDir.Close()
		fmt.Println(testfilenamesAfter)
		for _, fileName := range expectedFilenames {
			assert.Contains(t, testfilenamesAfter, fileName)
		}
		os.Remove(expectedZipFilePath)
	})
}

func Test_findIntellijDirectory(t *testing.T) {
	t.Run("findIntellijDirectory - success", func(t *testing.T) {
		expectedResult := "IntelliJTest"
		testDirectoryPath := filepath.Join(testDir, expectedResult)
		os.MkdirAll(testDirectoryPath, 0755)
		result := findIntellijDirectory(testDir)
		assert.Equal(t, expectedResult, result)
		os.Remove(testDirectoryPath)
	})
}

func Test_gatherCodewindIntellijLogs(t *testing.T) {
	t.Run("gatherCodewindIntellijLogs - success with provided path", func(t *testing.T) {
		testLogsPath := filepath.Join(testDir, "IntelliJTest")
		testLogsDir := "logDir"
		testLogsFile := "testFile.log"
		expectedLogFileContents := "Log: Test"
		expectedLogDirectory := "intellijLogs"
		expectedLogOutputPath := filepath.Join(testDir, expectedLogDirectory, testLogsDir, testLogsFile)
		testLogsFilePath := filepath.Join(testLogsPath, testLogsDir, testLogsFile)
		os.MkdirAll(filepath.Dir(testLogsFilePath), 0755)
		ioutil.WriteFile(testLogsFilePath, []byte(expectedLogFileContents), 0666)
		diagnosticsDirName = testDir
		gatherCodewindIntellijLogs(testLogsPath)
		assert.FileExists(t, expectedLogOutputPath, "Unable to find expected log file "+expectedLogOutputPath)
		contents, rfErr := ioutil.ReadFile(expectedLogOutputPath)
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, expectedLogFileContents, string(contents))
	})
	t.Run("gatherCodewindIntellijLogs - success with default", func(t *testing.T) {
		homeDir = testDir
		testLogsIntelliDir := "IntelliJTest"
		var testLogsPath string
		switch runtime.GOOS {
		case "darwin":
			testLogsPath = filepath.Join(testDir, "Library", "Logs", "JetBrains", testLogsIntelliDir)
		case "linux":
			testLogsPath = filepath.Join(testDir, ".cache", "JetBrains", testLogsIntelliDir, "log")
		case "windows":
			testLogsPath = filepath.Join(testDir, "AppData", "Local", "JetBrains", testLogsIntelliDir, "log")
		}
		testLogsDir := "logDir"
		testLogsFile := "testFile.log"
		expectedLogFileContents := "Log: Default Test"
		expectedLogDirectory := "intellijLogs"
		expectedLogOutputPath := filepath.Join(testDir, expectedLogDirectory, testLogsDir, testLogsFile)
		testLogsFilePath := filepath.Join(testLogsPath, testLogsDir, testLogsFile)
		os.MkdirAll(filepath.Dir(testLogsFilePath), 0755)
		ioutil.WriteFile(testLogsFilePath, []byte(expectedLogFileContents), 0666)
		diagnosticsDirName = testDir
		gatherCodewindIntellijLogs("")
		assert.FileExists(t, expectedLogOutputPath, "Unable to find expected log file "+expectedLogOutputPath)
		contents, rfErr := ioutil.ReadFile(expectedLogOutputPath)
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, expectedLogFileContents, string(contents))
	})
}

func Test_gatherCodewindVSCodeLogs(t *testing.T) {
	t.Run("gatherCodewindVSCodeLogs - success with default", func(t *testing.T) {
		homeDir = testDir
		var testLogsPath string
		switch runtime.GOOS {
		case "darwin":
			testLogsPath = filepath.Join(testDir, "Library", "Application Support", "Code", "logs")
		case "linux":
			testLogsPath = filepath.Join(testDir, ".config", "Code", "logs")
		case "windows":
			testLogsPath = filepath.Join(testDir, "AppData", "Roaming", "Code", "logs")
		}
		testLogsDir := "logDir"
		testLogsFile := "testFile.log"
		expectedLogFileContents := "Log: Default Test"
		expectedLogDirectory := "vsCodeLogs"
		expectedLogOutputPath := filepath.Join(testDir, expectedLogDirectory, "logs", testLogsDir, testLogsFile)
		testLogsFilePath := filepath.Join(testLogsPath, testLogsDir, testLogsFile)
		os.MkdirAll(filepath.Dir(testLogsFilePath), 0755)
		ioutil.WriteFile(testLogsFilePath, []byte(expectedLogFileContents), 0666)
		diagnosticsDirName = testDir
		gatherCodewindVSCodeLogs()
		assert.FileExists(t, expectedLogOutputPath, "Unable to find expected log file "+expectedLogOutputPath)
		contents, rfErr := ioutil.ReadFile(expectedLogOutputPath)
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, expectedLogFileContents, string(contents))
	})
}

func Test_gatherCodewindEclipseLogs(t *testing.T) {
	t.Run("gatherCodewindEclipseLogs - success with provided path", func(t *testing.T) {
		testEclipsePath := filepath.Join(testDir, "EclipseTest")
		testLogsPath := filepath.Join(testEclipsePath, ".metadata")
		testLogsFile := "testFile.log"
		expectedLogFileContents := "Log: Test"
		expectedLogDirectory := "eclipseLogs"
		expectedLogOutputPath := filepath.Join(testDir, expectedLogDirectory, testLogsFile)
		testLogsFilePath := filepath.Join(testLogsPath, testLogsFile)
		os.MkdirAll(filepath.Dir(testLogsFilePath), 0755)
		ioutil.WriteFile(testLogsFilePath, []byte(expectedLogFileContents), 0666)
		diagnosticsDirName = testDir
		gatherCodewindEclipseLogs(testEclipsePath)
		assert.FileExists(t, expectedLogOutputPath, "Unable to find expected log file "+expectedLogOutputPath)
		contents, rfErr := ioutil.ReadFile(expectedLogOutputPath)
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, expectedLogFileContents, string(contents))
	})
}

func Test_collectCodewindProjectContainers(t *testing.T) {
	printAsJSON = false
	t.Run("collectCodewindProjectContainers - docker client error", func(t *testing.T) {
		warning := "Unable to get Docker client"
		description := dockerClientError.Error()
		expectedConsoleOutput := warning + ": " + description + "\n"
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		collectCodewindProjectContainers(getDockerClientError)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))
	})
	t.Run("writeContainerLogToFile - error from docker client", func(t *testing.T) {
		expectedError := docker.DockerError{Op: docker.ErrOpContainerList, Err: docker.ErrContainerList, Desc: docker.ErrContainerList.Error()}
		expectedConsoleOutput := "Unable to get Docker container list: " + expectedError.Error() + "\n"
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		collectCodewindProjectContainers(getMockDockerErrorClient)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))

	})
	t.Run("collectCodewindProjectContainers - success but can't find containers", func(t *testing.T) {
		expectedConsoleOutput := "Collecting information from container /cw-testProject ... Unable to find"
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		collectCodewindProjectContainers(getMockDockerClient)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), expectedConsoleOutput)
	})
}

func Test_collectCodewindContainers(t *testing.T) {
	printAsJSON = false
	t.Run("collectCodewindContainers - success but can't find containers", func(t *testing.T) {
		expectedConsoleOutput := "Collecting information from container codewind-pfe ... Unable to inspect Container ID"
		os.MkdirAll(testDir, 0755)
		diagnosticsDirName = testDir
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		collectCodewindContainers(getMockDockerClient)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), expectedConsoleOutput)
	})
}

func Test_dgLocalCommand(t *testing.T) {
	printAsJSON = false
	t.Run("dgLocalCommand - success collecting projects", func(t *testing.T) {
		expectedConsoleOutput := "Collecting local Codewind workspace ... "
		expectedYamlFileContent := "Expected yaml file content"
		yamlFileName := "docker-compose.yaml"
		diagnosticsLocalDirName = filepath.Join(testDir, "local")
		expectedYamlOutputPath := filepath.Join(diagnosticsLocalDirName, yamlFileName)
		os.MkdirAll(diagnosticsLocalDirName, 0755)
		codewindHome = testDir
		ioutil.WriteFile(filepath.Join(codewindHome, yamlFileName), []byte(expectedYamlFileContent), 0666)
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		dgLocalCommand(true)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), expectedConsoleOutput)
		assert.FileExists(t, expectedYamlOutputPath, "Unable to find expected yaml file "+expectedYamlOutputPath)
		contents, rfErr := ioutil.ReadFile(expectedYamlOutputPath)
		if rfErr != nil {
			t.Error("Error encountered - " + rfErr.Error())
		}
		assert.Equal(t, expectedYamlFileContent, string(contents))
	})
}

func Test_writePodLogToFile(t *testing.T) {
	printAsJSON = false
	t.Run("writePodLogToFile", func(t *testing.T) {
		//kubernetes.fake panics if you try to use streams
		defer func() {
			if err := recover(); err != nil {
				t.Log("Got expected panic")
			} else {
				t.Error("Did not panic as expected")
			}
		}()
		writePodLogToFile(mockClientset, *mockPFEPod, mockPFEPod.GetName())
	})
}

func Test_collectPodInfo(t *testing.T) {
	printAsJSON = false
	diagnosticsDirName = testDir
	t.Run("collectPodInfo", func(t *testing.T) {
		//kubernetes.fake panics if you try to use streams
		defer func() {
			if err := recover(); err != nil {
				t.Log("Got expected panic")
				assert.FileExists(t, filepath.Join(testDir, "local", mockPFEPod.GetName()+".describe"), "Unable to find expected file "+filepath.Join(testDir, "local", mockPFEPod.GetName()+".describe"))
			} else {
				t.Error("Did not panic as expected")
			}
		}()
		collectPodInfo(mockClientset, []v1.Pod{*mockPFEPod}, "local")
	})
}

func Test_confirmConnectionIDAndWorkspaceID(t *testing.T) {
	oldGetAllConnections := getAllConnections
	getAllConnections = returnMockConnections
	t.Run("confirmConnectionIDAndWorkspaceID - connection not found", func(t *testing.T) {
		expectedConsoleOutput := "connection_not_found: Unable to associate  with existing connection\n"
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		connectionID, workspaceID := confirmConnectionIDAndWorkspaceID("")
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), expectedConsoleOutput)
		assert.Equal(t, "", connectionID)
		assert.Equal(t, "", workspaceID)
	})
	t.Run("confirmConnectionIDAndWorkspaceID - correct ID", func(t *testing.T) {
		connectionID, workspaceID := confirmConnectionIDAndWorkspaceID("local")
		assert.Equal(t, "local", connectionID)
		assert.Equal(t, "", workspaceID)
	})
	t.Run("confirmConnectionIDAndWorkspaceID - correct Label", func(t *testing.T) {
		connectionID, workspaceID := confirmConnectionIDAndWorkspaceID("Codewind local connection")
		assert.Equal(t, "local", connectionID)
		assert.Equal(t, "", workspaceID)
	})
	getAllConnections = oldGetAllConnections
}

func Test_getDockerVersions(t *testing.T) {
	printAsJSON = false
	t.Run("getDockerVersions - docker client error", func(t *testing.T) {
		warning := "Problems getting docker client"
		description := dockerClientError.Error()
		versionText := "Unable to determine version - "
		expectedConsoleOutput := warning + ": " + description + "\n"
		expectedVersionOutput := versionText + description
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		clientVersion, serverVersion := getDockerVersions(getDockerClientError)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))
		assert.Equal(t, expectedVersionOutput, clientVersion)
		assert.Equal(t, expectedVersionOutput, serverVersion)
	})
	t.Run("getDockerVersions - error from docker client", func(t *testing.T) {
		expectedError := docker.DockerError{Op: docker.ErrDockerVersion, Err: docker.ErrServerVersion, Desc: docker.ErrServerVersion.Error()}
		expectedConsoleOutput := "Problems getting docker server version: " + expectedError.Error() + "\n"
		versionText := "Unable to determine version - "
		expectedVersionOutput := versionText + expectedError.Error()
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		clientVersion, serverVersion := getDockerVersions(getMockDockerErrorClient)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))
		assert.Equal(t, "", clientVersion)
		assert.Equal(t, expectedVersionOutput, serverVersion)

	})
	t.Run("getDockerVersions - success but can't find containers", func(t *testing.T) {
		clientVersion, serverVersion := getDockerVersions(getMockDockerClient)
		assert.Equal(t, "", clientVersion)
		assert.Equal(t, "", serverVersion)

	})
}

func Test_dgRemoteCommand(t *testing.T) {
	oldGetAllConnections := getAllConnections
	getAllConnections = returnMockConnections
	// this runs collectPodInfo, which panics - so success is collecting the panic
	printAsJSON = false
	t.Run("dgRemoteCommand - codewind pod panic", func(t *testing.T) {
		mockClientset = fake.NewSimpleClientset(mockPFEPod, mockDeployment)
		diagnosticsDirName = testDir
		defer func() {
			if err := recover(); err != nil {
				t.Log("Got expected panic")
				assert.FileExists(t, filepath.Join(testDir, "local", mockPFEPod.GetName()+".describe"), "Unable to find expected file "+filepath.Join(testDir, "local", mockPFEPod.GetName()+".describe"))
			} else {
				t.Error("Did not panic as expected")
			}
		}()
		dgRemoteCommand("local", true, mockClientset)
	})
	t.Run("dgRemoteCommand - project pod panic", func(t *testing.T) {
		mockClientset = fake.NewSimpleClientset(mockProjectPod, mockDeployment)
		diagnosticsDirName = testDir
		defer func() {
			if err := recover(); err != nil {
				t.Log("Got expected panic")
				assert.FileExists(t, filepath.Join(testDir, "local", "projects", mockProjectPod.GetName()+".describe"), "Unable to find expected file "+filepath.Join(testDir, "local", "projects", mockPFEPod.GetName()+".describe"))
			} else {
				t.Error("Did not panic as expected")
			}
		}()
		dgRemoteCommand("local", true, mockClientset)
	})
	getAllConnections = oldGetAllConnections
}

func Test_DiagnosticsCollect(t *testing.T) {
	printAsJSON = false
	localCommandCalled := "Local Commmand function was called\n"
	remoteCommandCalled := "Remote Command function was called\n"
	oldDgFunctions := dgFunctions
	dgFunctions.CollectLocal = func(flag bool) { logDG(localCommandCalled) }
	dgFunctions.CollectRemote = func(str string, flag bool, clientset kubernetes.Interface) { logDG(remoteCommandCalled) }
	oldGetAllConnections := getAllConnections
	getAllConnections = returnMockConnections
	t.Run("DiagnosticsCollect - collect all ", func(t *testing.T) {
		diagnosticsDirName = testDir
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		flagSet.Bool("all", true, "")
		context := cli.NewContext(app, flagSet, nil)
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		DiagnosticsCollect(context)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), localCommandCalled)
		assert.Contains(t, string(out), remoteCommandCalled)
	})
	t.Run("DiagnosticsCollect - collect remote ", func(t *testing.T) {
		diagnosticsDirName = testDir
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		flagSet.String("conid", "remote", "")
		context := cli.NewContext(app, flagSet, nil)
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		DiagnosticsCollect(context)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.NotContains(t, string(out), localCommandCalled)
		assert.Contains(t, string(out), remoteCommandCalled)
	})
	t.Run("DiagnosticsCollect - collect local ", func(t *testing.T) {
		diagnosticsDirName = testDir
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		context := cli.NewContext(app, flagSet, nil)
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		DiagnosticsCollect(context)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Contains(t, string(out), localCommandCalled)
		assert.NotContains(t, string(out), remoteCommandCalled)
	})
	t.Run("DiagnosticsCollect - json output ", func(t *testing.T) {
		diagnosticsDirName = testDir
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		context := cli.NewContext(app, flagSet, nil)
		printAsJSON = true
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		DiagnosticsCollect(context)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		readStructure := dgResultStruct{}
		_ = json.Unmarshal(out, &readStructure)
		assert.Equal(t, true, readStructure.DgSuccess)
		assert.Equal(t, testDir, readStructure.DgOutputDir)
		printAsJSON = false
	})
	t.Run("DiagnosticsCollect - no collection with JSON ", func(t *testing.T) {
		dirToDelete := filepath.Join(testDir, "dgCollectDir")
		expectedOutput := "No diagnostics data was able to be collected - empty directory " + dirToDelete + " has been deleted."
		expectedExitValue := 1
		observedExitValue := 0
		os.MkdirAll(dirToDelete, 0755)
		oldExitFunc := dgFunctions.Exit
		dgFunctions.Exit = func(code int) { observedExitValue = code }
		oldDGDir := diagnosticsDirName
		diagnosticsDirName = dirToDelete
		oldHomeDir := homeDir
		homeDir = dirToDelete
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		flagSet.Bool("nozip", true, "")
		context := cli.NewContext(app, flagSet, nil)
		printAsJSON = false
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		DiagnosticsCollect(context)
		w.Close()
		os.Stdout = originalStdout
		out, _ := ioutil.ReadAll(r)
		homeDir = oldHomeDir
		diagnosticsDirName = oldDGDir
		dgFunctions.Exit = oldExitFunc
		_, err := os.Stat(dirToDelete)
		assert.Contains(t, string(out), expectedOutput)
		assert.Equal(t, observedExitValue, expectedExitValue)
		assert.True(t, os.IsNotExist(err), dirToDelete+" still exists")
	})
	dgFunctions = oldDgFunctions
	getAllConnections = oldGetAllConnections
}

func Test_DiagnosticsRemove(t *testing.T) {
	t.Run("DiagnosticsRemove - success", func(t *testing.T) {
		dirToDelete := filepath.Join(testDir, "dgRemoveDir")
		os.MkdirAll(dirToDelete, 0755)
		oldDGMDir := diagnosticsMasterDirName
		diagnosticsMasterDirName = dirToDelete
		expectedConsoleOutput := "Deleting all collected diagnostics files ... done\n"
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		app := cli.NewApp()
		flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
		context := cli.NewContext(app, flagSet, nil)
		DiagnosticsRemove(context)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		_, err := os.Stat(dirToDelete)
		assert.True(t, os.IsNotExist(err), dirToDelete+" still exists")
		assert.Contains(t, string(out), expectedConsoleOutput)
		diagnosticsMasterDirName = oldDGMDir
	})
}
