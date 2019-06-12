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
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

func TestToggleDebug(t *testing.T) {
	os.Setenv("DEBUG", "true")
	var debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	assert.Equal(t, debug, true, "should return true: debug flag should be true")
}

func TestRemoveImage(t *testing.T) {
	performanceImage := "docker.io/ibmcom/codewind-performance-amd64"
	PullImage(performanceImage, "")
	RemoveImage(performanceImage)
}
func TestCheckImageStatusFalse(t *testing.T) {
	// Test checks that image list can be searched
	// False return as no images have been installed for this test
	result := CheckImageStatus()
	assert.Equal(t, result, false, "should return false: no images are installed")
}

func TestCheckContainerStatusFalse(t *testing.T) {
	// Test checks that container list can be searched
	// False return as no containers have been started for this test
	result := CheckContainerStatus()
	assert.Equal(t, result, false, "should return false: no containers are started")
}

func TestPullDockerImage(t *testing.T) {
	performanceImage := "docker.io/ibmcom/codewind-performance-amd64"
	performanceImageTarget := "codewind-performance-amd64:latest"
	PullImage(performanceImage, "")
	TagImage(performanceImage, performanceImageTarget)

	ctx := context.Background()
	cli, _ := client.NewEnvClient()
	images, _ := cli.ImageList(ctx, types.ImageListOptions{})
	imageStatus := false
	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		if strings.Contains(imageRepo, "codewind") {
			imageStatus = true
			assert.Equal(t, imageStatus, true, "should return true: imageStatus should be true")
		}
	}
	cmd := exec.Command("docker", "image", "rm", "ibmcom/codewind-performance-amd64", performanceImageTarget, "-f")
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
	got := WriteToComposeFile("TestFile.yaml")
	assert.Equal(t, got, true, "should return true: should write data to a temp file")
	os.Remove("TestFile.yaml")
}

func TestWriteToComposeFileFail(t *testing.T) {
	writeToFile := WriteToComposeFile("")
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
