// /*******************************************************************************
//  * Copyright (c) 2019 IBM Corporation and others.
//  * All rights reserved. This program and the accompanying materials
//  * are made available under the terms of the Eclipse Public License v2.0
//  * which accompanies this distribution, and is available at
//  * http://www.eclipse.org/legal/epl-v20.html
//  *
//  * Contributors:
//  *     IBM Corporation - initial API and implementation
//  *******************************************************************************/

package utils

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestToggleDebug(t *testing.T) {
	os.Setenv("DEBUG", "true")
	var debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	assert.Equal(t, debug, true, "should return true: debug flag should be true")
}

func TestRemoveImage(t *testing.T) {
	performanceImage := "docker.io/eclipse/codewind-performance-amd64"
	client, err := NewDockerClient()
	if err != nil {
		t.Fail()
	}
	PullImage(client, performanceImage, false)
	RemoveImage(performanceImage)
}
func TestCheckImageStatusFalse(t *testing.T) {
	// Test checks that image list can be searched
	// False return as no images have been installed for this test
	client, err := NewDockerClient()
	if err != nil {
		t.Fail()
	}
	result, err := CheckImageStatus(client)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, result, false, "should return false: no images are installed")
}

func TestCheckContainerStatusFalse(t *testing.T) {
	// Test checks that container list can be searched
	// False return as no containers have been started for this test
	client, err := NewDockerClient()
	if err != nil {
		t.Fail()
	}
	result, err := CheckContainerStatus(client)

	if err != nil {
		t.Fail()
	}
	assert.Equal(t, result, false, "should return false: no containers are started")
}

func TestPullDockerImage(t *testing.T) {
	performanceImage := "docker.io/eclipse/codewind-performance-amd64"
	performanceImageTarget := "codewind-performance-amd64:latest"
	client, dockerErr := NewDockerClient()
	if dockerErr != nil {
		t.Fail()
	}
	PullImage(client, performanceImage, false)

	ctx := context.Background()
	images, _ := client.ImageList(ctx, types.ImageListOptions{})
	imageStatus := false
	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		if strings.Contains(imageRepo, "codewind") {
			imageStatus = true
			assert.Equal(t, imageStatus, true, "should return true: imageStatus should be true")
		}
	}
	cmd := exec.Command("docker", "image", "rm", "eclipse/codewind-performance-amd64", performanceImageTarget, "-f")
	cmd.Stdin = strings.NewReader("Deleting pulled image")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to delete test images")
	}
}

func TestCreateTempFile(t *testing.T) {
	file := CreateTempFile("TestFile.yaml")
	assert.Equal(t, file, true, "should return true: should create a temp file")
	os.Remove("./TestFile.yaml")
}

func TestWriteToComposeFile(t *testing.T) {
	os.Create("TestFile.yaml")
	got := WriteToComposeFile("TestFile.yaml", false)
	assert.Equal(t, got, true, "should return true: should write data to a temp file")
	os.Remove("TestFile.yaml")
}

func TestWriteToComposeFileFail(t *testing.T) {
	writeToFile := WriteToComposeFile("", false)
	assert.Equal(t, writeToFile, false, "should return false: should fail to write data")
}

func TestDeleteTempFile(t *testing.T) {
	os.Create("TestFile.yaml")
	removeFile, _ := DeleteTempFile("TestFile.yaml")
	assert.Equal(t, removeFile, true, "should return true: should delete the temp file")
}

func TestDeleteTempFileFail(t *testing.T) {
	errString := "stat TestFile.yaml: no such file or directory"
	_, err := DeleteTempFile("TestFile.yaml")
	assert.EqualError(t, err, errString)
}

func TestRemoveDuplicateEntries(t *testing.T) {
	dupArr := []string{"test", "test", "test"}
	result := RemoveDuplicateEntries(dupArr)

	if len(result) != 1 {
		log.Fatal("Test 1: Failed to delete duplicate array index")
	}

	dupArr = []string{"", "test", "test"}
	result = RemoveDuplicateEntries(dupArr)
	if len(result) != 1 {
		log.Fatal("Test 2: Failed to delete duplicate array index")
	}

	dupArr = []string{"", "", ""}
	result = RemoveDuplicateEntries(dupArr)
	if len(result) != 0 {
		log.Fatal("Test 3: Failed to identify empty array values")
	}
}

func Test_GetContainersToRemove(t *testing.T) {
	tests := map[string]struct {
		containerList      []types.Container
		expectedContainers []string
	}{
		"Returns project containers (cw-)": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-nodejsexpress"},
				},
				types.Container{
					Names: []string{"/cw-springboot"},
				},
			},
			expectedContainers: []string{
				"/cw-nodejsexpress",
				"/cw-springboot",
			},
		},
		"Ignores a non-codewind container": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-valid-container"},
				},
				types.Container{
					Names: []string{"invalid-container"},
				},
			},
			expectedContainers: []string{
				"/cw-valid-container",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			containersToRemove := GetContainersToRemove(test.containerList)
			assert.Equal(t, len(test.expectedContainers), len(containersToRemove))
			for _, container := range containersToRemove {
				assert.Contains(t, test.expectedContainers, container.Names[0])
			}
		})
	}
}
