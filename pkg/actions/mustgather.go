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
	"archive/tar"
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
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

var codewindHome = filepath.Join(homeDir, ".codewind")
var nowTime = time.Now().Format("20060102150405")
var mustGatherMasterDirName = filepath.Join(codewindHome, "mustgather")
var mustGatherDirName = filepath.Join(mustGatherMasterDirName, nowTime)

var isLoud = true

func logMG(input string) {
	if isLoud {
		fmt.Println(input)
	}
}

//MustGatherCommand to gather logs and project files to aid diagnosis of Codewind errors
func MustGatherCommand(c *cli.Context) {
	if c.Bool("clean") {
		logMG("Deleting all collected mustgather files")
		err := os.RemoveAll(mustGatherMasterDirName)
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
	} else {
		mgCommand(c)
	}
}

func mgCommand(c *cli.Context) {
	if c.Bool("quiet") {
		isLoud = false
	}
	dirErr := os.MkdirAll(filepath.Join(mustGatherDirName, "projects"), 0755)
	if dirErr != nil {
		errors.CheckErr(dirErr, 205, "")
	}
	logMG("Mustgather files will be collected at " + mustGatherDirName)
	collectCodewindContainers()

	// Collect Codewind PFE workspace
	logMG("Collecting Codewind workspace")
	pfeContainerID := getContainerID(docker.PfeContainerName)
	copyCodewindWorkspace(pfeContainerID)

	if c.Bool("projects") {
		collectCodewindProjectContainers()
	}

	// Collect docker-compose file
	logMG("Collecting docker-compose.yaml")
	utils.CopyFile(filepath.Join(codewindHome, "docker-compose.yaml"), filepath.Join(mustGatherDirName, "docker-compose.yaml"))

	// Collect codewind versions
	gatherCodewindVersions()

	// Attempt to gather Eclipse logs
	gatherCodewindEclipseLogs(c.String("eclipseWorkspaceDir"))

	// Attempt to gather VSCode logs
	gatherCodewindVSCodeLogs()

	if !c.Bool("nozip") {
		createZipAndRemoveCollectedFiles()
	}
}

// Collect Codewind container inspection & logs
func collectCodewindContainers() {
	for _, cwContainerName := range docker.ContainerNames {
		logMG("Collecting information from container " + cwContainerName)
		containerID := getContainerID(cwContainerName)
		writeContainerInspectToFile(containerID, cwContainerName)
		writeContainerLogToFile(containerID, cwContainerName)
	}
}

func collectCodewindProjectContainers() {
	// Collect project container inspection & logs
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	// using getContainerListWithOptions to pick up all containers, including stopped ones
	allContainers, cListErr := docker.GetContainerListWithOptions(dockerClient, types.ContainerListOptions{All: true})
	if cListErr != nil {
		HandleDockerError(cListErr)
		os.Exit(1)
	}
	for _, cwContainer := range docker.GetCodewindProjectContainers(allContainers) {
		logMG("Collecting information from container " + cwContainer.Names[0])
		writeContainerInspectToFile(cwContainer.ID, filepath.Join("projects", cwContainer.Names[0]))
		writeContainerLogToFile(cwContainer.ID, filepath.Join("projects", cwContainer.Names[0]))
	}
}

func gatherCodewindEclipseLogs(codewindEclipseWSDir string) {
	// Attempt to gather Eclipse logs
	if codewindEclipseWSDir != "" {
		codewindEclipseWSLogDir := filepath.Join(codewindEclipseWSDir, ".metadata")
		if _, err := os.Stat(codewindEclipseWSLogDir); !os.IsNotExist(err) {
			files, dirErr := ioutil.ReadDir(codewindEclipseWSLogDir)
			if dirErr != nil {
				logMG("Unable to collect Eclipse logs - directory read error " + dirErr.Error())
			}
			logMG("Collecting Eclipse Logs")
			eclipseLogDir := "eclipseLogs"
			mustGatherEclipseLogPath := filepath.Join(mustGatherDirName, eclipseLogDir)
			logDirErr := os.MkdirAll(mustGatherEclipseLogPath, 0755)
			if logDirErr != nil {
				errors.CheckErr(logDirErr, 205, "")
			}
			for _, f := range files {
				fileName := f.Name()
				if f.Mode().IsRegular() && strings.HasSuffix(fileName, ".log") {
					utils.CopyFile(filepath.Join(codewindEclipseWSLogDir, fileName), filepath.Join(mustGatherDirName, eclipseLogDir, fileName))
				}
			}
		} else {
			logMG("Unable to collect Eclipse logs - workspace metadata directory not found")
		}
	} else {
		logMG("Unable to collect Eclipse logs - workspace not specified")
	}
}

func gatherCodewindVSCodeLogs() {
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
		if _, err := os.Stat(vsCodeLogsDir); !os.IsNotExist(err) {
			err := filepath.Walk(vsCodeLogsDir, func(path string, info os.FileInfo, err error) error {
				localPath := filepath.Join(mustGatherVsCodeLogPath, strings.Replace(path, vsCodeDir, "", 1))
				if info.IsDir() {
					logDirErr := os.MkdirAll(localPath, 0755)
					if logDirErr != nil {
						errors.CheckErr(logDirErr, 205, "")
					}
				}
				if info.Mode().IsRegular() {
					// strip out mustGatherDirName from target, as copyFileHere adds it back
					utils.CopyFile(path, localPath)
				}
				return nil
			})
			if err != nil {
				logMG("walk error " + err.Error())
			}
		} else {
			logMG("Unable to collect VSCode logs - cannot find logs directory")
		}
	} else {
		logMG("Unable to collect VSCode logs - cannot find logs directory")
	}
}

func createZipAndRemoveCollectedFiles() {
	// zip
	logMG("Creating mustgather.zip")
	mustGatherZipFileName := "mustgather." + nowTime + ".zip"
	zipErr := utils.Zip(mustGatherZipFileName, mustGatherDirName)
	if zipErr != nil {
		errors.CheckErr(zipErr, 401, "")
	}
	// remove other files & directories from mustgather directory
	mgDir, err := os.Open(mustGatherDirName)
	if err != nil {
		errors.CheckErr(err, 205, "")
	}
	defer mgDir.Close()
	filenames, err := mgDir.Readdirnames(-1)
	if err != nil {
		errors.CheckErr(err, 205, "")
	}
	for _, filename := range filenames {
		if filename == mustGatherZipFileName {
			continue
		}
		err = os.RemoveAll(filepath.Join(mustGatherDirName, filename))
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
	}
}

func gatherCodewindVersions() {
	logMG("Collecting version information")
	//dockerClient, dockerErr := docker.NewDockerClient()
	//if dockerErr != nil {
	//	HandleDockerError(dockerErr)
	//	os.Exit(1)
	//}
	//dockerClientVersion := docker.GetClientVersion(dockerClient)
	//dockerServerVersion, gsvErr := docker.GetServerVersion(dockerClient)
	containerVersions := GetContainerVersions("local")
	versionsByteArray := []byte(
		"CWCTL VERSION: " + containerVersions.CwctlVersion + "\n" +
			"PFE VERSION: " + containerVersions.PFEVersion + "\n" +
			"PERFORMANCE VERSION: " + containerVersions.PerformanceVersion)
	versionsErr := ioutil.WriteFile(filepath.Join(mustGatherDirName, "codewind.versions"), versionsByteArray, 0644)
	if versionsErr != nil {
		errors.CheckErr(versionsErr, 201, "")
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
	if len(container) > 0 {
		return container[0].ID
	}
	return ""
}

//writeContainerInspectToFile - writes the results of `docker inspect containerId` to a file
func writeContainerInspectToFile(containerID, containerName string) error {
	if containerID == "" {
		logMG("Unable to find " + containerName + " container")
		return nil
	}
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
	if containerID == "" {
		logMG("Unable to find " + containerName + " container")
		return nil
	}
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

//copyCodewindWorkspace - copies the Codewind PFE container's workspace to mustgather
func copyCodewindWorkspace(containerID string) error {
	if containerID == "" {
		logMG("Unable to find Codewind PFE container")
		return nil
	}
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	tarFileStream, fileErr := docker.GetFilesFromContainer(dockerClient, containerID, "/codewind-workspace")
	if fileErr != nil {
		HandleDockerError(fileErr)
		os.Exit(1)
	}
	defer tarFileStream.Close()
	// Extracting tarred files
	tarBallReader := tar.NewReader(tarFileStream)

	return utils.ExtractTarToFileSystem(tarBallReader, mustGatherDirName)
}
