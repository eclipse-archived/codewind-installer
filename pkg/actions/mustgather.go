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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/eclipse/codewind-installer/pkg/appconstants"
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/urfave/cli"
)

var codewindHome = filepath.Join(homeDir, ".codewind")
var mustGatherDirName = filepath.Join(codewindHome, "mustgather", time.Now().Format("20060102150405"))

func logMG(input ...string) {
	fmt.Println(input[0])
}

//MustGatherCommand to gather logs and project files to aid diagnosis of Codewind errors
func MustGatherCommand(c *cli.Context) {
	dirErr := os.MkdirAll(filepath.Join(mustGatherDirName, "projects"), 0755)
	if dirErr != nil {
		errors.CheckErr(dirErr, 205, "")
	}
	logMG("Mustgather files will be collected at " + mustGatherDirName)

	// Collect Codewind container inspection & logs
	for _, cwContainerName := range docker.ContainerNames {
		logMG("Collecting information from container " + cwContainerName)
		containerID := getContainerID(cwContainerName)
		writeContainerInspectToFile(containerID, cwContainerName)
		writeContainerLogToFile(containerID, cwContainerName)
	}

	// Collect project container inspection & logs
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	allContainers, cListErr := docker.GetContainerListWithOptions(dockerClient, types.ContainerListOptions{All: true})
	if cListErr != nil {
		HandleDockerError(cListErr)
		os.Exit(1)
	}
	for _, cwContainer := range docker.GetContainersToRemove(allContainers) {
		logMG("Collecting information from container " + cwContainer.Names[0])
		writeContainerInspectToFile(cwContainer.ID, filepath.Join("projects", cwContainer.Names[0]))
		writeContainerLogToFile(cwContainer.ID, filepath.Join("projects", cwContainer.Names[0]))
	}

	// Collect docker-compose file
	copyFileHere(filepath.Join(codewindHome, "docker-compose.yaml"), "docker-compose.yaml")

	// Collect codewind version
	logMG("Collecting CWCTL version")
	d1 := []byte(appconstants.VersionNum)
	versionErr := ioutil.WriteFile(filepath.Join(mustGatherDirName, "cwctl.version"), d1, 0644)
	if versionErr != nil {
		errors.CheckErr(versionErr, 201, "")
	}

	// Attempt to gather VSCode logs
	logMG("Collecting VSCode logs")
	vsCodeDir := ""
	switch runtime.GOOS {
	case "darwin":
		vsCodeDir = filepath.Join(homeDir, "Library", "Application Support", "Code")
	case "linux":
		vsCodeDir = filepath.Join(homeDir, ".config", "Code")
	case "windows":
		vsCodeDir = filepath.Join(homeDir, "AppData", "Roaming", "Code")
	}
	if len(vsCodeDir) > 0 {
		vsCodeLogsDir := filepath.Join(vsCodeDir, "logs")
		
		mustGatherVsCodeLogPath := filepath.Join(mustGatherDirName, "vsCodeLogs")
		dirErr := os.MkdirAll(mustGatherVsCodeLogPath, 0755)
		if dirErr != nil {
			errors.CheckErr(dirErr, 205, "")
		}
		err := filepath.Walk(vsCodeLogsDir, func(path string, info os.FileInfo, err error) error {
			localPath := filepath.Join(mustGatherVsCodeLogPath, strings.Replace(path, vsCodeDir, "", 1))
			logMG("path = " + path)
			logMG("localPath = " + localPath)
			if info.IsDir() {
				logMG(path + " is a Directory")
				// dirErr := os.MkdirAll(, 0755)
				// if dirErr != nil {
				//   errors.CheckErr(dirErr, 205, "")
				// }
			}
			if info.Mode().IsRegular() {
				logMG(path + " is a Regular file")
			}
			return nil
		})
		if err != nil {
			logMG("walk error " + err.Error())
		}

	} else {
		logMG("Unable to collect VSCode logs - cannot find logs directory")
	}

}

//getContainerID - returns the ID of the container filtered by name
func getContainerID(containerName string) string {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	nameFilter := filters.NewArgs(filters.Arg("name", containerName))
	container, getErr := docker.GetContainerListWithOptions(dockerClient, types.ContainerListOptions{All: true, Filters: nameFilter})
	if getErr != nil {
		HandleDockerError(getErr)
		os.Exit(1)
	}
	return container[0].ID
}

//writeContainerInspectToFile - writes the results of `docker inspect containerId` to a file
func writeContainerInspectToFile(containerID, containerName string) error {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	inspectedContents, inspectErr := docker.InspectContainer(dockerClient, containerID)
	if inspectErr != nil {
		HandleDockerError(inspectErr)
		os.Exit(1)
	}
	fileContents, _ := json.MarshalIndent(inspectedContents, "", " ")
	err := ioutil.WriteFile(filepath.Join(mustGatherDirName, containerName+".inspect"), fileContents, 0644)
	return err
}

//writeContainerLogToFile - writes the results of `docker logs containerId` to a file
func writeContainerLogToFile(containerID, containerName string) error {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	logStream, logErr := docker.GetContainerLogs(dockerClient, containerID)
	if logErr != nil {
		HandleDockerError(logErr)
		os.Exit(1)
	}
	outFile, createErr := os.Create(filepath.Join(mustGatherDirName, containerName+".log"))
	if createErr != nil {
		errors.CheckErr(createErr, 201, "")
	}
	defer outFile.Close()
	_, err := io.Copy(outFile, logStream)
	return err
}

//copyFileHere - copies the contents of the source file to a target file in the mustgather directory
func copyFileHere(sourceFilePath, targetFile string) error {
	logMG("Collecting " + targetFile)
	sourceFileStat, err := os.Stat(sourceFilePath)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", sourceFilePath)
	}

	source, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}
	defer source.Close()
	dst := filepath.Join(mustGatherDirName, targetFile)
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
