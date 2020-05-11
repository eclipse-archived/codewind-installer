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
		fmt.Print(input)
	}
}

type dgWarning struct {
	WarningType string `json:"warning"`
	WarningDesc string `json:"warning_description"`
}

var dgWarningArray = []dgWarning{}

func warnDG(warning, description string) {
	if printAsJSON {
		dgWarningArray = append(dgWarningArray, dgWarning{WarningType: warning, WarningDesc: description})
	} else {
		logDG(warning + ": " + description + "\n")
	}
}

func errDG(err, description string) {
	if printAsJSON {
		outputStruct := struct {
			ErrorType string `json:"error"`
			ErrorDesc string `json:"error_description"`
		}{ErrorType: err, ErrorDesc: description}
		json, _ := json.Marshal(outputStruct)
		fmt.Println(string(json))
	} else {
		logDG(err + ": " + description + "\n")
	}
}

//DiagnosticsCommand to gather logs and project files to aid diagnosis of Codewind errors
func DiagnosticsCommand(c *cli.Context) {
	if c.Bool("quiet") || printAsJSON {
		isLoud = false
	}
	if c.Bool("clean") {
		logDG("Deleting all collected diagnostics files ... ")
		err := os.RemoveAll(diagnosticsMasterDirName)
		if err != nil {
			errors.CheckErr(err, 206, "")
		}
		logDG("done\n")
	} else {
		dirErr := os.MkdirAll(filepath.Join(diagnosticsDirName, dgProjectDirName), 0755)
		if dirErr != nil {
			errors.CheckErr(dirErr, 205, "")
		}
		logDG("Diagnostics files will be written to " + diagnosticsDirName + "\n")
		if c.String("conid") != "local" {
			dgRemoteCommand(c)
		} else {
			dgLocalCommand(c)
		}
		dgSharedCommand(c)
		if printAsJSON {
			outputStruct := struct {
				DgOutputDir           string      `json:"outputdir"`
				DgWarningsEncountered []dgWarning `json:"warnings_encountered"`
			}{DgOutputDir: diagnosticsDirName, DgWarningsEncountered: dgWarningArray}
			json, _ := json.Marshal(outputStruct)
			fmt.Println(string(json))
		}
	}
}

func dgRemoteCommand(c *cli.Context) {
	// find the connectionID specified by conid - could be ID or Label
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	kubeNameSpace := ""
	workspaceID := ""
	connectionList, conErr := connections.GetAllConnections()
	if conErr != nil {
		errDG("connections_error", conErr.Error())
		os.Exit(1)
	}
	found := false
	for _, connection := range connectionList {
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.Label) {
			connectionID = connection.ID
		}
		if strings.ToUpper(connectionID) == strings.ToUpper(connection.ID) {
			found = true
			workspaceID = strings.Replace(connection.ClientID, codewindPodPrefix, "", 1)
			break
		}
	}
	if !found {
		errDG("connection_not_found", "Unable to associate "+connectionID+" with existing connection")
		os.Exit(1)
	}
	existingDeployments, edErr := remote.GetExistingDeployments("")
	if edErr != nil {
		errDG("existing_deployment_error", edErr.Error())
		os.Exit(1)
	}
	found = false
	for _, existingDeployment := range existingDeployments {
		if strings.ToUpper(existingDeployment.WorkspaceID) == strings.ToUpper(workspaceID) {
			kubeNameSpace = existingDeployment.Namespace
			found = true
			break
		}
	}
	if !found {
		errDG("existing_deployment_error", "Unable to locate existing deployment with Workspace ID "+workspaceID)
		os.Exit(1)
	}
	config, err := remote.GetKubeConfig()
	if err != nil {
		errDG("kube_config_error", err.Error())
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		errDG("kube_client_error", err.Error())
		os.Exit(1)
	}
	cwBasePods, nspErr := clientset.CoreV1().Pods(kubeNameSpace).List(metav1.ListOptions{LabelSelector: "codewindWorkspace=" + workspaceID})
	if nspErr != nil {
		errDG("kube_podlist_error", nspErr.Error())
		os.Exit(1)
	}
	collectPodInfo(clientset, cwBasePods.Items)
	if c.Bool("projects") {
		logDG("Collecting project containers")
		cwProjPods, cwPPErr := clientset.CoreV1().Pods(kubeNameSpace).List(metav1.ListOptions{FieldSelector: "spec.serviceAccountName=" + codewindPodPrefix + workspaceID, LabelSelector: "codewindWorkspace!=" + workspaceID})
		if cwPPErr != nil {
			errDG("kube_podlist_error", "Unable to retrieve Kubernetes Pods: "+cwPPErr.Error())
			os.Exit(1)
		}
		collectPodInfo(clientset, cwProjPods.Items)
	}

	// Collect codewind versions
	gatherCodewindVersions(connectionID)
}

func collectPodInfo(clientset *kubernetes.Clientset, podArray []corev1.Pod) {
	for _, pod := range podArray {
		podName := pod.ObjectMeta.Name
		logDG("Collecting information from pod " + podName + " ... ")
		// Pod struct contains all details to be found in kubectl describe
		writeJSONStructToFile(pod, podName+".describe")
		writePodLogToFile(clientset, pod, podName)
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
	collectCodewindContainers()

	// Collect Codewind PFE workspace
	logDG("Collecting Codewind workspace ... ")
	pfeContainerID := getContainerID(docker.PfeContainerName)
	copyCodewindWorkspace(pfeContainerID)
	logDG("done\n")

	if c.Bool("projects") {
		collectCodewindProjectContainers()
	}

	// Collect codewind versions
	gatherCodewindVersions("local")
}

func dgSharedCommand(c *cli.Context) {

	// Collect docker-compose file
	logDG("Collecting docker-compose.yaml ... ")
	utils.CopyFile(filepath.Join(codewindHome, "docker-compose.yaml"), filepath.Join(diagnosticsDirName, "docker-compose.yaml"))
	logDG("done\n")

	// Attempt to gather Eclipse logs
	gatherCodewindEclipseLogs(c.String("eclipseWorkspaceDir"))

	// Attempt to gather VSCode logs
	gatherCodewindVSCodeLogs()

	// Attempt to gather IntelliJ logs
	gatherCodewindIntellijLogs(c.String("intellijLogsDir"))

	if !c.Bool("nozip") {
		createZipAndRemoveCollectedFiles()
	}
}

// Collect Codewind container inspection & logs
func collectCodewindContainers() {
	for _, cwContainerName := range docker.LocalCWContainerNames {
		logDG("Collecting information from container " + cwContainerName + " ... ")
		containerID := getContainerID(cwContainerName)
		writeContainerInspectToFile(containerID, cwContainerName)
		writeContainerLogToFile(containerID, cwContainerName)
		logDG("done\n")
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
		logDG("Collecting information from container " + cwContainer.Names[0] + " ... ")
		relativeFilePath := filepath.Join(dgProjectDirName, cwContainer.Names[0])
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
	if cvErr != nil {
		if strings.Contains(cvErr.Error(), "certificate signed by unknown authority") {
			warnDG("Problems getting Codewind container versions - please run command again specifying global option '--insecure'", cvErr.Error())
		} else {
			//just log and continue; version file will have "Unknown" values
			warnDG("Problems getting Codewind container versions", cvErr.Error())
		}
	}
	versionsByteArray := []byte(
		"CWCTL VERSION: " + containerVersions.CwctlVersion + "\n" +
			"PFE VERSION: " + containerVersions.PFEVersion + "\n" +
			"PERFORMANCE VERSION: " + containerVersions.PerformanceVersion)
	if connectionID != "local" {
		versionsByteArray = []byte(
			"CWCTL VERSION: " + containerVersions.CwctlVersion + "\n" +
				"PFE VERSION: " + containerVersions.PFEVersion + "\n" +
				"PERFORMANCE VERSION: " + containerVersions.PerformanceVersion + "\n" +
				"GATEKEEPER VERSION: " + containerVersions.GatekeeperVersion)
	}
	versionsErr := ioutil.WriteFile(filepath.Join(diagnosticsDirName, "codewind.versions"), versionsByteArray, 0644)
	if versionsErr != nil {
		errors.CheckErr(versionsErr, 201, "")
	}
	logDG("done\n")
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
		warnDG("Unable to find "+containerName+" container", "could not get container ID")
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
		warnDG("Unable to find "+containerName+" container", "could not get container ID")
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
		warnDG("Unable to find Codewind PFE container", "could not get container ID")
		return nil
	}
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}
	codewindWorkspace := "codewind-workspace"
	for _, path := range []string{".appsody", ".config", ".extensions", ".logs", ".projects"} {
		tarFileStream, fileErr := docker.GetFilesFromContainer(dockerClient, containerID, "/"+codewindWorkspace+"/"+path)
		if fileErr != nil {
			HandleDockerError(fileErr)
			os.Exit(1)
		}
		defer tarFileStream.Close()
		// Extracting tarred files
		tarBallReader := tar.NewReader(tarFileStream)

		extractErr := utils.ExtractTarToFileSystem(tarBallReader, filepath.Join(diagnosticsDirName, codewindWorkspace))
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
