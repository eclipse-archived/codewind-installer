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
	goErr "errors"
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
var diagnosticsLocalDirName = filepath.Join(diagnosticsDirName, "local")

const codewindPrefix = "codewind-"
const codewindProjectPrefix = "cw-"
const dgProjectDirName = "projects"

var collectingAll = false

func logDG(input string) {
	if !printAsJSON {
		fmt.Print(input)
	}
}

type dgWarning struct {
	WarningType string `json:"warning"`
	WarningDesc string `json:"warning_description"`
}

type dgResultStruct struct {
	DgSuccess             bool        `json:"success"`
	DgOutputDir           string      `json:"outputdir"`
	DgWarningsEncountered []dgWarning `json:"warnings_encountered"`
}

var dgWarningArray = []dgWarning{}

func warnDG(warning, description string) {
	if printAsJSON {
		dgWarningArray = append(dgWarningArray, dgWarning{WarningType: warning, WarningDesc: description})
	} else {
		logDG(warning + ": " + description + "\n")
	}
}

//DiagnosticsCollect to gather logs and project files to aid diagnosis of Codewind errors
func DiagnosticsCollect(c *cli.Context) {
	collectingAll = c.Bool("all")
	connectionID := c.String("conid")
	dirErr := os.MkdirAll(diagnosticsDirName, 0755)
	if dirErr != nil {
		errors.CheckErr(dirErr, 205, "")
	}
	logDG("Diagnostics files will be written to " + diagnosticsDirName + "\n")
	if collectingAll {
		connectionList, conErr := connections.GetAllConnections()
		if conErr != nil {
			warnDG("connections_error", "Unable to get Connections "+conErr.Error())
		} else {
			for _, connection := range connectionList {
				if connection.ClientID != "" {
					dgRemoteCommand(connection.ID, c.Bool("projects"))
				}
			}
		}
		dgLocalCommand(c)
	} else if connectionID != "local" {
		dgRemoteCommand(connectionID, c.Bool("projects"))
	} else {
		dgLocalCommand(c)
	}
	// Attempt to gather Eclipse logs
	gatherCodewindEclipseLogs(c.String("eclipseWorkspaceDir"))

	// Attempt to gather VSCode logs
	gatherCodewindVSCodeLogs()

	// Attempt to gather IntelliJ logs
	gatherCodewindIntellijLogs(c.String("intellijLogsDir"))
	if !c.Bool("nozip") {
		createZipAndRemoveCollectedFiles()
	}
	// check to see if we got any data back
	entries, _ := ioutil.ReadDir(diagnosticsDirName)
	if len(entries) == 0 {
		// clean up and output failure message
		err := os.RemoveAll(diagnosticsDirName)
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
		if printAsJSON {
			result := dgResultStruct{DgSuccess: false, DgOutputDir: "has been deleted", DgWarningsEncountered: dgWarningArray}
			json, _ := json.Marshal(result)
			fmt.Println(string(json))
		} else {
			logDG("No diagnostics data was able to be collected - empty directory " + diagnosticsDirName + " has been deleted.")
		}
		os.Exit(1)
	}
	if printAsJSON {
		result := dgResultStruct{DgSuccess: true, DgOutputDir: diagnosticsDirName, DgWarningsEncountered: dgWarningArray}
		json, _ := json.Marshal(result)
		fmt.Println(string(json))
	}
}

//DiagnosticsRemove to remove the diagnostics directory and all its contents
func DiagnosticsRemove(c *cli.Context) {
	logDG("Deleting all collected diagnostics files ... ")
	err := os.RemoveAll(diagnosticsMasterDirName)
	if err != nil {
		errors.CheckErr(err, 206, "")
	}
	logDG("done\n")
}

func dgRemoteCommand(conid string, collectProjects bool) {
	connectionID, workspaceID := confirmConnectionIDAndWorkspaceID(conid)
	if connectionID == "" {
		return
	}
	existingDeployments, edErr := remote.GetExistingDeployments("")
	if edErr != nil {
		warnDG("existing_deployment_error", edErr.Error())
		return
	}
	config, err := remote.GetKubeConfig()
	if err != nil {
		warnDG("kube_config_error", err.Error())
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		warnDG("kube_client_error", err.Error())
		return
	}
	found := false
	kubeNameSpace := ""
	for _, existingDeployment := range existingDeployments {
		if strings.ToUpper(existingDeployment.WorkspaceID) == strings.ToUpper(workspaceID) {
			kubeNameSpace = existingDeployment.Namespace
			found = true
			break
		}
	}
	if !found {
		warnDG("existing_deployment_error", "Unable to locate existing deployment with Workspace ID "+workspaceID)
		return
	}
	diagnosticsRemoteDirName := filepath.Join(diagnosticsDirName, connectionID)
	cwBasePods, cwBPErr := clientset.CoreV1().Pods(kubeNameSpace).List(metav1.ListOptions{LabelSelector: "codewindWorkspace=" + workspaceID})
	if cwBPErr != nil {
		warnDG("kube_podlist_error", "Unable to retrieve Kubernetes Pods: "+cwBPErr.Error())
	} else {
		connDirErr := os.MkdirAll(diagnosticsRemoteDirName, 0755)
		if connDirErr != nil {
			errors.CheckErr(connDirErr, 205, "")
		}
		collectPodInfo(clientset, cwBasePods.Items, connectionID)
	}
	if collectProjects {
		cwProjPods, cwPPErr := clientset.CoreV1().Pods(kubeNameSpace).List(metav1.ListOptions{FieldSelector: "spec.serviceAccountName=" + codewindPrefix + workspaceID, LabelSelector: "codewindWorkspace!=" + workspaceID})
		if cwPPErr != nil {
			warnDG("kube_podlist_error", "Unable to retrieve Kubernetes Pods: "+cwPPErr.Error())
		} else {
			connDirErr := os.MkdirAll(filepath.Join(diagnosticsRemoteDirName, dgProjectDirName), 0755)
			if connDirErr != nil {
				errors.CheckErr(connDirErr, 205, "")
			}
			collectPodInfo(clientset, cwProjPods.Items, filepath.Join(connectionID, dgProjectDirName))
		}
	}
	gatherCodewindVersions(connectionID)
}

func confirmConnectionIDAndWorkspaceID(conid string) (string, string) {
	connectionList, conErr := connections.GetAllConnections()
	if conErr != nil {
		warnDG("connections_error", "Unable to get Connections "+conErr.Error())
		return "", ""
	}
	connectionID := strings.TrimSpace(strings.ToLower(conid))
	for _, connection := range connectionList {
		// could have been passed a remote connection label instead of an ID
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.Label) {
			connectionID = connection.ID
		}
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.ID) {
			return connection.ID, strings.Replace(connection.ClientID, codewindPrefix, "", 1)
		}
	}
	// if we reach here it means we couldn't find a connection
	warnDG("connection_not_found", "Unable to associate "+connectionID+" with existing connection")
	return "", ""
}

func collectPodInfo(clientset *kubernetes.Clientset, podArray []corev1.Pod, workspaceDirName string) {
	for _, pod := range podArray {
		podName := pod.ObjectMeta.Name
		logDG("Collecting information from pod " + podName + " ... ")
		// Pod struct contains all details to be found in kubectl describe
		writeJSONStructToFile(pod, filepath.Join(workspaceDirName, podName+".describe"))
		writePodLogToFile(clientset, pod, filepath.Join(workspaceDirName, podName))
		logDG("done\n")
	}
}

func writePodLogToFile(clientset *kubernetes.Clientset, pod corev1.Pod, podName string) error {
	podLogOpts := corev1.PodLogOptions{}
	// get pod logs
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		warnDG("Unable to obtain logs for pod "+podName, err.Error())
	}
	defer podLogs.Close()
	return writeStreamToFile(podLogs, podName+".log")
}

func dgLocalCommand(c *cli.Context) {
	localDirErr := os.MkdirAll(filepath.Join(diagnosticsLocalDirName, dgProjectDirName), 0755)
	if localDirErr != nil {
		errors.CheckErr(localDirErr, 205, "")
	}
	collectCodewindContainers()

	// Collect Codewind PFE workspace
	logDG("Collecting local Codewind workspace ... ")

	pfeContainerID := getContainerID(docker.PfeContainerName)
	copyCodewindWorkspace(pfeContainerID)
	logDG("done\n")

	if c.Bool("projects") {
		collectCodewindProjectContainers()
	}

	// Collect codewind versions
	gatherCodewindVersions("local")

	// Collect docker-compose file
	logDG("Collecting local docker-compose.yaml ... ")
	utils.CopyFile(filepath.Join(codewindHome, "docker-compose.yaml"), filepath.Join(diagnosticsLocalDirName, "docker-compose.yaml"))
	logDG("done\n")
}

// Collect Codewind container inspection & logs
func collectCodewindContainers() {
	for _, cwContainerName := range docker.LocalCWContainerNames {
		logDG("Collecting information from container " + cwContainerName + " ... ")
		containerID := getContainerID(cwContainerName)
		writeContainerInspectToFile(containerID, filepath.Join("local", cwContainerName))
		writeContainerLogToFile(containerID, filepath.Join("local", cwContainerName))
		logDG("done\n")
	}
}

func collectCodewindProjectContainers() {
	// Collect project container inspection & logs
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Unable to get Docker client", dockerErr.Error())
		return
	}
	// using getContainerListWithOptions to pick up all containers, including stopped ones
	allContainers, cListErr := docker.GetContainerListWithOptions(dockerClient, types.ContainerListOptions{All: true})
	if cListErr != nil {
		warnDG("Unable to get Docker container list", cListErr.Error())
		return
	}
	for _, cwContainer := range docker.GetCodewindProjectContainers(allContainers) {
		logDG("Collecting information from container " + cwContainer.Names[0] + " ... ")
		relativeFilePath := filepath.Join("local", dgProjectDirName, cwContainer.Names[0])
		writeContainerInspectToFile(cwContainer.ID, relativeFilePath)
		writeContainerLogToFile(cwContainer.ID, relativeFilePath)
		logDG("done\n")
	}
}

func gatherCodewindEclipseLogs(codewindEclipseWSDir string) {
	// Attempt to gather Eclipse logs
	if codewindEclipseWSDir != "" {
		codewindEclipseWSLogDir := filepath.Join(codewindEclipseWSDir, ".metadata")
		if _, err := os.Stat(codewindEclipseWSLogDir); !os.IsNotExist(err) {
			files, dirErr := ioutil.ReadDir(codewindEclipseWSLogDir)
			if dirErr != nil {
				warnDG("Unable to collect Eclipse logs - directory read error ", dirErr.Error())
			}
			logDG("Collecting Eclipse Logs ... ")
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
			logDG("done\n")
		} else {
			warnDG("Unable to collect Eclipse logs", "workspace metadata directory not found")
		}
	} else {
		warnDG("Unable to collect Eclipse logs", "workspace not specified")
	}
}

func gatherCodewindVSCodeLogs() {
	logDG("Collecting VSCode logs ... ")
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
				warnDG("walk error ", err.Error())
			}
			logDG("done\n")
		} else {
			warnDG("Unable to collect VSCode logs", "cannot find logs directory")
		}
	} else {
		warnDG("Unable to collect VSCode logs", "cannot find logs directory")
	}
}

func findIntellijDirectory(inDir string) string {
	foundFile := ""
	dgDir, err := os.Open(inDir)
	if err != nil {
		return foundFile
	}
	defer dgDir.Close()
	filenames, err := dgDir.Readdirnames(-1)
	if err != nil {
		return foundFile
	}
	for _, filename := range filenames {
		if strings.Contains(filename, "IntelliJ") {
			foundFile = filename
			break
		}
	}
	return foundFile
}

func gatherCodewindIntellijLogs(codewindIntellijLogDir string) {
	logDG("Collecting Intellij logs ... ")
	intellijLogsDir := codewindIntellijLogDir
	if intellijLogsDir == "" {
		// attempt to use default path
		switch runtime.GOOS {
		case "darwin":
			libraryLogsDir := filepath.Join(homeDir, "Library", "Logs", "JetBrains")
			intellijDirName := findIntellijDirectory(libraryLogsDir)
			if intellijDirName != "" {
				intellijLogsDir = filepath.Join(libraryLogsDir, intellijDirName)
			}
		case "linux":
			jetBrainsDir := filepath.Join(homeDir, ".cache", "JetBrains")
			intellijDirName := findIntellijDirectory(jetBrainsDir)
			if intellijDirName != "" {
				intellijLogsDir = filepath.Join(jetBrainsDir, intellijDirName, "log")
			}
		case "windows":
			jetBrainsDir := filepath.Join(homeDir, "AppData", "Local", "JetBrains")
			intellijDirName := findIntellijDirectory(jetBrainsDir)
			if intellijDirName != "" {
				intellijLogsDir = filepath.Join(jetBrainsDir, intellijDirName, "log")
			}
		}
	}
	if len(intellijLogsDir) > 0 {
		diagnosticsIntellijLogPath := filepath.Join(diagnosticsDirName, "intellijLogs")
		if _, err := os.Stat(intellijLogsDir); !os.IsNotExist(err) {
			err := filepath.Walk(intellijLogsDir, func(path string, info os.FileInfo, err error) error {
				localPath := filepath.Join(diagnosticsIntellijLogPath, strings.Replace(path, intellijLogsDir, "", 1))
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
				warnDG("walk error ", err.Error())
			}
			logDG("done\n")
		} else {
			warnDG("Unable to collect Intellij logs", "cannot find logs directory")
		}
	} else {
		warnDG("Unable to collect Intellij logs", "cannot find logs directory")
	}
}

func createZipAndRemoveCollectedFiles() {
	// zip
	diagnosticsZipFileName := "diagnostics." + nowTime + ".zip"
	logDG("Creating " + diagnosticsZipFileName + " ... ")
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
	logDG("done\n")
}

func gatherCodewindVersions(connectionID string) {
	logDG("Collecting version information ... ")
	containerVersions, cvErr := GetContainerVersions(connectionID)
	errorString := ""
	if cvErr != nil {
		if strings.Contains(cvErr.Error(), "certificate signed by unknown authority") {
			warnDG("Problems getting Codewind container versions - please run command again specifying global option '--insecure'", cvErr.Error())
		} else {
			//just log and continue; version file will have "Unknown" values
			warnDG("Problems getting Codewind container versions", cvErr.Error())
		}
		errorString = cvErr.Error()
	}
	versionsByteArray := []byte(
		"CWCTL VERSION: " + containerVersions.CwctlVersion + errorString + "\n" +
			"PFE VERSION: " + containerVersions.PFEVersion + errorString + "\n" +
			"PERFORMANCE VERSION: " + containerVersions.PerformanceVersion + errorString + "\n")
	if connectionID == "local" {
		dockerClientVersion, dockerServerVersion := getDockerVersions()
		versionsByteArray = append(versionsByteArray, []byte(
			"DOCKER CLIENT VERSION: "+dockerClientVersion+"\n"+
				"DOCKER SERVER VERSION: "+dockerServerVersion+"\n")...,
		)
	} else {
		versionsByteArray = append(versionsByteArray, []byte(
			"GATEKEEPER VERSION: "+containerVersions.GatekeeperVersion+errorString+"\n")...,
		)
	}
	versionsErr := ioutil.WriteFile(filepath.Join(diagnosticsDirName, connectionID, "codewind.versions"), versionsByteArray, 0644)
	if versionsErr != nil {
		errors.CheckErr(versionsErr, 201, "")
	}
	logDG("done\n")
}

func getDockerVersions() (clientVersion, serverVersion string) {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Problems getting docker client", dockerErr.Error())
		return "Unable to determine version - " + dockerErr.Error(), "Unable to determine version - " + dockerErr.Error()
	}
	dockerClientVersion := docker.GetClientVersion(dockerClient)
	dockerServerVersion, gsvErr := docker.GetServerVersion(dockerClient)
	if gsvErr != nil {
		warnDG("Problems getting docker server version", gsvErr.Error())
		return dockerClientVersion, "Unable to determine version - " + gsvErr.Error()
	}
	return dockerClientVersion, dockerServerVersion.Version
}

//getContainerID - returns the ID of the container filtered by name
func getContainerID(containerName string) string {
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Unable to get Docker client", dockerErr.Error())
		return ""
	}
	nameFilter := filters.NewArgs(filters.Arg("name", containerName))
	container, getErr := docker.GetContainerListWithOptions(dockerClient, types.ContainerListOptions{All: true, Filters: nameFilter})
	if getErr != nil {
		warnDG("Unable to get Docker container list", getErr.Error())
		return ""
	}
	if len(container) > 0 {
		return container[0].ID
	}
	return ""
}

//writeContainerInspectToFile - writes the results of `docker inspect containerId` to a file
func writeContainerInspectToFile(containerID, containerName string) error {
	if containerID == "" {
		warnDG("Unable to find "+containerName+" container", "could not get container ID")
		return goErr.New("Unable to find " + containerName + " container - could not get container ID")
	}
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Unable to get Docker client", dockerErr.Error())
		return dockerErr
	}
	inspectedContents, inspectErr := docker.InspectContainer(dockerClient, containerID)
	if inspectErr != nil {
		warnDG("Unable to inspect Container ID "+containerID, inspectErr.Error())
		return inspectErr
	}
	return writeJSONStructToFile(inspectedContents, containerName+".inspect")
}

//writeContainerLogToFile - writes the results of `docker logs containerId` to a file
func writeContainerLogToFile(containerID, containerName string) error {
	if containerID == "" {
		warnDG("Unable to find "+containerName+" container", "could not get container ID")
		return goErr.New("Unable to find " + containerName + " container - could not get container ID")
	}
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Unable to get Docker client", dockerErr.Error())
		return dockerErr
	}
	logStream, logErr := docker.GetContainerLogs(dockerClient, containerID)
	if logErr != nil {
		warnDG("Unable to get container logs for container "+containerID, logErr.Error())
		return logErr
	}
	return writeStreamToFile(logStream, containerName+".log")
}

//copyCodewindWorkspace - copies the Codewind PFE container's workspace to diagnostics
func copyCodewindWorkspace(containerID string) error {
	if containerID == "" {
		warnDG("Unable to find Codewind PFE container", "could not get container ID")
		return goErr.New("Unable to find Codewind PFE container - could not get container ID")
	}
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		warnDG("Unable to get Docker client", dockerErr.Error())
		return dockerErr
	}
	codewindWorkspace := "codewind-workspace"
	for _, path := range []string{".appsody", ".config", ".extensions", ".logs", ".projects"} {
		tarFileStream, fileErr := docker.GetFilesFromContainer(dockerClient, containerID, "/"+codewindWorkspace+"/"+path)
		if fileErr != nil {
			warnDG("Unable to get files from container ID "+containerID, fileErr.Error())
			return fileErr
		}
		defer tarFileStream.Close()
		// Extracting tarred files
		tarBallReader := tar.NewReader(tarFileStream)

		extractErr := utils.ExtractTarToFileSystem(tarBallReader, filepath.Join(diagnosticsLocalDirName, codewindWorkspace))
		if extractErr != nil {
			return extractErr
		}
	}
	return nil
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
