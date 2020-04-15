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
	"archive/zip"
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
var nowTime = time.Now().Format("20060102150405")
var mustGatherDirName = filepath.Join(codewindHome, "mustgather", nowTime)

var isLoud = true

func logMG(input ...string) {
	if isLoud {
		fmt.Println(input[0])
	}
}

//MustGatherCommand to gather logs and project files to aid diagnosis of Codewind errors
func MustGatherCommand(c *cli.Context) {
	if c.Bool("quiet") {
		isLoud = false
	}
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

	// Collect Codewind PFE workspace
	logMG("Collecting Codewind workspace")
	pfeContainerID := getContainerID(docker.PfeContainerName)
	copyCodewindWorkspace(pfeContainerID)

	if c.Bool("projects") {
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
	}

	// Collect docker-compose file
	logMG("Collecting docker-compose.yaml")
	copyFileHere(filepath.Join(codewindHome, "docker-compose.yaml"), "docker-compose.yaml")

	// Collect codewind version
	logMG("Collecting CWCTL version")
	d1 := []byte(appconstants.VersionNum)
	versionErr := ioutil.WriteFile(filepath.Join(mustGatherDirName, "cwctl.version"), d1, 0644)
	if versionErr != nil {
		errors.CheckErr(versionErr, 201, "")
	}

	// Attempt to gather Eclipse logs
	codewindEclipseWSDir := c.String("eclipseWorkspaceDir")
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
					copyFileHere(filepath.Join(codewindEclipseWSLogDir, fileName), filepath.Join(eclipseLogDir, fileName))
				}
			}
		} else {
			logMG("Unable to collect Eclipse logs - workspace metadata directory not found")
		}
	} else {
		logMG("Unable to collect Eclipse logs - workspace not specified")
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
					copyFileHere(path, strings.Replace(localPath, mustGatherDirName, "", 1))
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

	// zip
	logMG("Creating mustgather.zip")
	newZipFileName := "mustgather." + nowTime + ".zip"
	newZipFile, zipCreateErr := os.Create(filepath.Join(mustGatherDirName, newZipFileName))
	if zipCreateErr != nil {
		logMG("Unable to create zip file - " + zipCreateErr.Error())
	} else {
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		// Add files to zip
		err := filepath.Walk(mustGatherDirName, func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() && (info.Name() != newZipFileName) {
				fileToZip, err := os.Open(path)
				if err != nil {
					return err
				}
				defer fileToZip.Close()

				// Get the file information
				info, err := fileToZip.Stat()
				if err != nil {
					return err
				}

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}

				// Using FileInfoHeader() above only uses the basename of the file. If we want
				// to preserve the folder structure we can overwrite this with the full path.
				header.Name = strings.Replace(path, mustGatherDirName+string(os.PathSeparator), "", 1)

				// Change to deflate to gain better compression
				// see http://golang.org/pkg/archive/zip/#pkg-constants
				header.Method = zip.Deflate

				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				_, err = io.Copy(writer, fileToZip)
				return err
			}
			return nil
		})
		if err != nil {
			logMG("walk error " + err.Error())
		}
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

//copyCodewindWorkspace - copies the Codewind PFE container's workspace to mustgather
func copyCodewindWorkspace(containerID string) error {
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

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return (err)
		}

		// get the individual filename and extract to the current directory
		filename := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			err = os.MkdirAll(filepath.Join(mustGatherDirName, filename), os.FileMode(header.Mode)) // or use 0755 if you prefer
			if err != nil {
				return err
			}

		case tar.TypeReg:
			// handle normal file
			writer, err := os.Create(filepath.Join(mustGatherDirName, filename))
			if err != nil {
				return err
			}

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filepath.Join(mustGatherDirName, filename), os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			writer.Close()
		default:
			logMG("Unable to untar type : " + string(header.Typeflag) + " in file " + filename)
		}
	}
	return nil
}

//copyFileHere - copies the contents of the source file to a target file in the mustgather directory
func copyFileHere(sourceFilePath, targetFile string) error {
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
