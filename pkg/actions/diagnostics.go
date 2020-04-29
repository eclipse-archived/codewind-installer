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
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/errors"
	"github.com/eclipse/codewind-installer/pkg/remote"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var codewindHome = filepath.Join(homeDir, ".codewind")
var nowTime = time.Now().Format("20060102150405")
var diagnosticsMasterDirName = filepath.Join(codewindHome, "diagnostics")
var diagnosticsDirName = filepath.Join(diagnosticsMasterDirName, nowTime)

const codewindPodPrefix = "codewind-"
const codewindProjectPrefix = "cw-"
const dgProjectDirName = "projects"

var isLoud = true

func logDG(input string) {
	if isLoud {
		fmt.Println(input)
	}
}

//DiagnosticsCommand to gather logs and project files to aid diagnosis of Codewind errors
func DiagnosticsCommand(c *cli.Context) {
	if c.Bool("clean") {
		logDG("Deleting all collected diagnostics files")
		err := os.RemoveAll(diagnosticsMasterDirName)
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
	} else {
		if c.Bool("quiet") || c.GlobalBool("json") {
			isLoud = false
		}
		dirErr := os.MkdirAll(filepath.Join(diagnosticsDirName, dgProjectDirName), 0755)
		if dirErr != nil {
			errors.CheckErr(dirErr, 205, "")
		}
		logDG("Diagnostics files will be written to " + diagnosticsDirName)
		if c.String("conid") != "local" {
			dgRemoteCommand(c)
		} else {
			dgLocalCommand(c)
		}
		dgSharedCommand(c)
		if c.GlobalBool("json") {
			outputStruct := struct {
				DgOutputDir string `json:"outputdir"`
			}{DgOutputDir: diagnosticsDirName}
			json, _ := json.Marshal(outputStruct)
			fmt.Println(string(json))
		}
	}
}

func dgRemoteCommand(c *cli.Context) {
	// find the connectionID specified by conid - could be ID or Label
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	kubeNameSpace := ""
	clientID := ""
	connectionList, conErr := connections.GetAllConnections()
	if conErr != nil {
		fmt.Println("Unable to get Connections " + conErr.Error())
		os.Exit(1)
	}
	found := false
	for _, connection := range connectionList {
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.Label) {
			connectionID = connection.ID
		}
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.ID) {
			found = true
			clientID = strings.Replace(connection.ClientID, codewindPodPrefix, "", 1)
			break
		}
	}
	if !found {
		fmt.Println("Unable to associate " + connectionID + " with existing connection")
		os.Exit(1)
	}
	existingDeployments, edErr := remote.GetExistingDeployments("")
	if edErr != nil {
		fmt.Println("Unable to get existing deployments " + edErr.Error())
		os.Exit(1)
	}
	found = false
	for _, existingDeployment := range existingDeployments {
		if strings.ToUpper(existingDeployment.WorkspaceID) == strings.ToUpper(clientID) {
			kubeNameSpace = existingDeployment.Namespace
			found = true
			break
		}
	}
	if !found {
		fmt.Println("Unable to locate existing deployment with Workspace ID " + clientID)
		os.Exit(1)
	}
	config, err := remote.GetKubeConfig()
	if err != nil {
		fmt.Printf("Unable to retrieve Kubernetes Config %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Unable to retrieve Kubernetes clientset %v\n", err)
		os.Exit(1)
	}
	nameSpacePods, nspErr := clientset.CoreV1().Pods(kubeNameSpace).List(metav1.ListOptions{})
	if nspErr != nil {
		fmt.Printf("Unable to retrieve Kubernetes Pods %v\n", nspErr)
		os.Exit(1)
	}
	for _, pod := range nameSpacePods.Items {
		podName := pod.ObjectMeta.Name
		if strings.HasPrefix(podName, codewindPodPrefix) {
			logDG("Collecting information from pod " + podName)
			// Pod struct contains all details to be found in kubectl describe
			writeJSONStructToFile(pod, podName+".describe")
			writePodLogToFile(clientset, pod, podName)
		}
		if c.Bool("projects") && strings.HasPrefix(podName, codewindProjectPrefix) {
			logDG("Collecting information from pod " + podName)
			// Pod struct contains all details to be found in kubectl describe
			writeJSONStructToFile(pod, filepath.Join(dgProjectDirName, podName+".describe"))
			writePodLogToFile(clientset, pod, filepath.Join(dgProjectDirName, podName))
		}
		if strings.HasPrefix(podName, docker.PfeContainerName) {
			logDG("Collecting Codewind workspace")
		}
	}
}

func writePodLogToFile(clientset *kubernetes.Clientset, pod corev1.Pod, podName string) error {
	podLogOpts := corev1.PodLogOptions{}
	// get pod logs
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		logDG("Unable to obtain logs for pod " + podName)
	}
	defer podLogs.Close()
	return writeStreamToFile(podLogs, podName+".log")
}

func dgLocalCommand(c *cli.Context) {
	collectCodewindContainers()

	// Collect Codewind PFE workspace
	logDG("Collecting Codewind workspace")
	pfeContainerID := getContainerID(docker.PfeContainerName)
	copyCodewindWorkspace(pfeContainerID)

	if c.Bool("projects") {
		collectCodewindProjectContainers()
	}
}

func dgSharedCommand(c *cli.Context) {

	// Collect docker-compose file
	logDG("Collecting docker-compose.yaml")
	utils.CopyFile(filepath.Join(codewindHome, "docker-compose.yaml"), filepath.Join(diagnosticsDirName, "docker-compose.yaml"))

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
	for _, cwContainerName := range docker.LocalCWContainerNames {
		logDG("Collecting information from container " + cwContainerName)
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
		logDG("Collecting information from container " + cwContainer.Names[0])
		relativeFilePath := filepath.Join(dgProjectDirName, cwContainer.Names[0])
		writeContainerInspectToFile(cwContainer.ID, relativeFilePath)
		writeContainerLogToFile(cwContainer.ID, relativeFilePath)
	}
}

func gatherCodewindEclipseLogs(codewindEclipseWSDir string) {
	// Attempt to gather Eclipse logs
	if codewindEclipseWSDir != "" {
		codewindEclipseWSLogDir := filepath.Join(codewindEclipseWSDir, ".metadata")
		if _, err := os.Stat(codewindEclipseWSLogDir); !os.IsNotExist(err) {
			files, dirErr := ioutil.ReadDir(codewindEclipseWSLogDir)
			if dirErr != nil {
				logDG("Unable to collect Eclipse logs - directory read error " + dirErr.Error())
			}
			logDG("Collecting Eclipse Logs")
			eclipseLogDir := "eclipseLogs"
			diagnosticsEclipseLogPath := filepath.Join(diagnosticsDirName, eclipseLogDir)
			logDirErr := os.MkdirAll(diagnosticsEclipseLogPath, 0755)
			if logDirErr != nil {
				errors.CheckErr(logDirErr, 205, "")
			}
			for _, f := range files {
				fileName := f.Name()
				if f.Mode().IsRegular() && strings.HasSuffix(fileName, ".log") {
					utils.CopyFile(filepath.Join(codewindEclipseWSLogDir, fileName), filepath.Join(diagnosticsEclipseLogPath, fileName))
				}
			}
		} else {
			logDG("Unable to collect Eclipse logs - workspace metadata directory not found")
		}
	} else {
		logDG("Unable to collect Eclipse logs - workspace not specified")
	}
}

func gatherCodewindVSCodeLogs() {
	logDG("Collecting VSCode logs")
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
		diagnosticsVsCodeLogPath := filepath.Join(diagnosticsDirName, "vsCodeLogs")
		dirErr := os.MkdirAll(diagnosticsVsCodeLogPath, 0755)
		if dirErr != nil {
			errors.CheckErr(dirErr, 205, "")
		}
		if _, err := os.Stat(vsCodeLogsDir); !os.IsNotExist(err) {
			err := filepath.Walk(vsCodeLogsDir, func(path string, info os.FileInfo, err error) error {
				localPath := filepath.Join(diagnosticsVsCodeLogPath, strings.Replace(path, vsCodeDir, "", 1))
				if info.IsDir() {
					logDirErr := os.MkdirAll(localPath, 0755)
					if logDirErr != nil {
						errors.CheckErr(logDirErr, 205, "")
					}
				}
				if info.Mode().IsRegular() {
					utils.CopyFile(path, localPath)
				}
				return nil
			})
			if err != nil {
				logDG("walk error " + err.Error())
			}
		} else {
			logDG("Unable to collect VSCode logs - cannot find logs directory")
		}
	} else {
		logDG("Unable to collect VSCode logs - cannot find logs directory")
	}
}

func createZipAndRemoveCollectedFiles() {
	// zip
	diagnosticsZipFileName := "diagnostics." + nowTime + ".zip"
	logDG("Creating " + diagnosticsZipFileName)
	zipErr := utils.Zip(diagnosticsZipFileName, diagnosticsDirName)
	if zipErr != nil {
		errors.CheckErr(zipErr, 401, "")
	}
	// remove other files & directories from diagnostics directory
	dgDir, err := os.Open(diagnosticsDirName)
	if err != nil {
		errors.CheckErr(err, 205, "")
	}
	defer dgDir.Close()
	filenames, err := dgDir.Readdirnames(-1)
	if err != nil {
		errors.CheckErr(err, 205, "")
	}
	for _, filename := range filenames {
		if filename == diagnosticsZipFileName {
			continue
		}
		err = os.RemoveAll(filepath.Join(diagnosticsDirName, filename))
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
	}
}

func gatherCodewindVersions() {
	logDG("Collecting version information")
	//dockerClient, dockerErr := docker.NewDockerClient()
	//if dockerErr != nil {
	//	HandleDockerError(dockerErr)
	//	os.Exit(1)
	//}
	//dockerClientVersion := docker.GetClientVersion(dockerClient)
	//dockerServerVersion, gsvErr := docker.GetServerVersion(dockerClient)
	containerVersions, cvErr := GetContainerVersions("local")
	if cvErr != nil {
		//just log and continue; version file will have "Unknown" values
		logDG("Problems getting Codewind container versions")
	}
	versionsByteArray := []byte(
		"CWCTL VERSION: " + containerVersions.CwctlVersion + "\n" +
			"PFE VERSION: " + containerVersions.PFEVersion + "\n" +
			"PERFORMANCE VERSION: " + containerVersions.PerformanceVersion)
	versionsErr := ioutil.WriteFile(filepath.Join(diagnosticsDirName, "codewind.versions"), versionsByteArray, 0644)
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
		logDG("Unable to find " + containerName + " container")
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
	return writeJSONStructToFile(inspectedContents, containerName+".inspect")
}

//writeContainerLogToFile - writes the results of `docker logs containerId` to a file
func writeContainerLogToFile(containerID, containerName string) error {
	if containerID == "" {
		logDG("Unable to find " + containerName + " container")
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
	return writeStreamToFile(logStream, containerName+".log")
}

//copyCodewindWorkspace - copies the Codewind PFE container's workspace to diagnostics
func copyCodewindWorkspace(containerID string) error {
	if containerID == "" {
		logDG("Unable to find Codewind PFE container")
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

	return utils.ExtractTarToFileSystem(tarBallReader, diagnosticsDirName)
}

//writeJSONStructToFile - writes the given struct to file as JSON
func writeJSONStructToFile(structure interface{}, targetFilePath string) error {
	fileContents, _ := json.MarshalIndent(structure, "", " ")
	err := ioutil.WriteFile(filepath.Join(diagnosticsDirName, targetFilePath), fileContents, 0644)
	return err
}

//writeStreamToFile - writes the given stream to a file
func writeStreamToFile(stream io.ReadCloser, targetFilePath string) error {
	outFile, createErr := os.Create(filepath.Join(diagnosticsDirName, targetFilePath))
	if createErr != nil {
		errors.CheckErr(createErr, 201, "")
	}
	defer outFile.Close()
	_, err := io.Copy(outFile, stream)
	return err
}
