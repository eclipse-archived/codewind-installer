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
	"encoding/json"
	goErr "errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/eclipse/codewind-installer/pkg/errors"
	logr "github.com/sirupsen/logrus"
)

// codewind-docker-compose.yaml data
var data = `
version: 2
services:
 codewind-pfe:
  image: ${REPOSITORY}codewind-pfe${PLATFORM}:${TAG}
  container_name: codewind-pfe
  user: root
  environment: ["HOST_WORKSPACE_DIRECTORY=${WORKSPACE_DIRECTORY}","CONTAINER_WORKSPACE_DIRECTORY=/codewind-workspace","HOST_OS=${HOST_OS}","CODEWIND_VERSION=${TAG}","PERFORMANCE_CONTAINER=codewind-performance${PLATFORM}:${TAG}","HOST_HOME=${HOST_HOME}","HOST_MAVEN_OPTS=${HOST_MAVEN_OPTS}"]
  depends_on: [codewind-performance]
  ports: ["127.0.0.1:${PFE_EXTERNAL_PORT}:9090"]
  volumes: ["/var/run/docker.sock:/var/run/docker.sock","cw-workspace:/codewind-workspace","${WORKSPACE_DIRECTORY}:/mounted-workspace"]
  networks: [network]
 codewind-performance:
  image: codewind-performance${PLATFORM}:${TAG}
  ports: ["127.0.0.1:9095:9095"]
  container_name: codewind-performance
  networks: [network]
networks:
  network:
   driver_opts:
    com.docker.network.bridge.host_binding_ipv4: "127.0.0.1"
volumes:
  cw-workspace:
`

// Compose struct for the docker compose yaml file
type Compose struct {
	Version  string `yaml:"version"`
	SERVICES struct {
		PFE struct {
			Image         string   `yaml:"image"`
			ContainerName string   `yaml:"container_name"`
			User          string   `yaml:"user"`
			Environment   []string `yaml:"environment"`
			DependsOn     []string `yaml:"depends_on"`
			Ports         []string `yaml:"ports"`
			Volumes       []string `yaml:"volumes"`
			Networks      []string `yaml:"networks"`
		} `yaml:"codewind-pfe"`
		PERFORMANCE struct {
			Image         string   `yaml:"image"`
			Ports         []string `yaml:"ports"`
			ContainerName string   `yaml:"container_name"`
			Volumes       []string `yaml:"volumes"`
			Networks      []string `yaml:"networks"`
		} `yaml:"codewind-performance"`
	} `yaml:"services"`
	VOLUME struct {
		CodewindWorkspace map[string]string `yaml:"cw-workspace"`
	} `yaml:"volumes"`
	NETWORKS struct {
		NETWORK struct {
			DRIVEROPTS struct {
				HostIP string `yaml:"com.docker.network.bridge.host_binding_ipv4"`
			} `yaml:"driver_opts"`
		} `yaml:"network"`
	} `yaml:"networks"`
}

// constant to identify the internal port of PFE in its container
const internalPFEPort = 9090

// constants to identify the range of external ports on which to expose PFE
const (
	minTCPPort   = 10000
	maxTCPPort   = 11000
	minDebugPort = 34000
	maxDebugPort = 35000
)

// DockerCompose to set up the Codewind environment
func DockerCompose(tempFilePath string, tag string) *DockerError {

	// Set env variables for the docker compose file
	home := os.Getenv("HOME")

	const GOARCH string = runtime.GOARCH
	const GOOS string = runtime.GOOS
	fmt.Println("System architecture is: ", GOARCH)
	fmt.Println("Host operating system is: ", GOOS)

	if GOARCH == "x86_64" || GOARCH == "amd64" {
		os.Setenv("PLATFORM", "-amd64")
	} else {
		os.Setenv("PLATFORM", "-"+GOARCH)
	}

	os.Setenv("REPOSITORY", "")
	os.Setenv("TAG", tag)
	if GOOS == "windows" {
		os.Setenv("WORKSPACE_DIRECTORY", "C:\\codewind-data")
		// In Windows, calling the env variable "HOME" does not return
		// the user directory correctly
		os.Setenv("HOST_HOME", os.Getenv("USERPROFILE"))

	} else {
		os.Setenv("WORKSPACE_DIRECTORY", home+"/codewind-data")
		os.Setenv("HOST_HOME", home)
	}
	os.Setenv("HOST_OS", GOOS)
	os.Setenv("COMPOSE_PROJECT_NAME", "codewind")
	os.Setenv("HOST_MAVEN_OPTS", os.Getenv("MAVEN_OPTS"))
	fmt.Printf("Attempting to find available port\n")
	portAvailable, port := IsTCPPortAvailable(minTCPPort, maxTCPPort)
	if !portAvailable {
		fmt.Printf("No available external ports in range, will default to Docker-assigned port")
	}
	os.Setenv("PFE_EXTERNAL_PORT", port)

	cmd := exec.Command("docker-compose", "-f", tempFilePath, "up", "-d", "--force-recreate")
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		DeleteTempFile(tempFilePath)
		return &DockerError{errOpDockerCompose, err, err.Error()}
	}
	fmt.Printf("Please wait whilst containers initialize... %s \n", output.String())
	cmd.Wait()
	fmt.Printf(output.String()) // Wait to finish execution, so we can read all output

	if strings.Contains(output.String(), "ERROR") || strings.Contains(output.String(), "error") {
		DeleteTempFile(tempFilePath)
		os.Exit(1)
	}

	if strings.Contains(output.String(), "The image for the service you're trying to recreate has been removed") {
		DeleteTempFile(tempFilePath)
		os.Exit(1)
	}
	return nil
}

// PullImage - pull pfe/performance images from dockerhub
func PullImage(image string, jsonOutput bool) *DockerError {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return &DockerError{errOpClientCreate, err, err.Error()}
	}

	var codewindOut io.ReadCloser

	codewindOut, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})

	if err != nil {
		return &DockerError{errOpImagePull, err, err.Error()}
	}

	if jsonOutput == true {
		defer codewindOut.Close()
		io.Copy(os.Stdout, codewindOut)
	} else {
		defer codewindOut.Close()
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		jsonmessage.DisplayJSONMessagesStream(codewindOut, os.Stderr, termFd, isTerm, nil)
	}
	return nil
}

// ValidateImageDigest - will ensure the image digest matches that of the one in dockerhub
// returns imageID, docker error
func ValidateImageDigest(image string) (string, *DockerError) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	errors.CheckErr(err, 200, "")
	// call docker api for image digest
	queryDigest, err := cli.DistributionInspect(ctx, image, "")
	if err != nil {
		logr.Error(err)
	}

	// turn digest -> []byte -> string
	digest, _ := json.Marshal(queryDigest.Descriptor.Digest)
	logr.Traceln("Query image digest is.. ", queryDigest.Descriptor.Digest)

	// get local image digest
	imageList, err := GetImageList()
	if err != nil {
		logr.Error(err)
	}
	imageName := strings.TrimPrefix(image, "docker.io/")
	imageArr := []string{
		imageName,
	}

	for _, image := range imageList {
		imageRepo := strings.Join(image.RepoDigests, " ")
		imageTags := strings.Join(image.RepoTags, " ")
		for _, index := range imageArr {
			if strings.Contains(imageTags, index) {
				if strings.Contains(imageRepo, strings.Replace(string(digest), "\"", "", -1)) {
					length := len(strings.Replace(string(digest), "\"", "", -1))
					last10 := strings.Replace(string(digest), "\"", "", -1)[length-10 : length]
					logr.Tracef("Validation for image digest ..%v succeeded\n", last10)
				} else {
					logr.Traceln("Local image digest did not match queried image digest from dockerhub - This could be a result of a bad download")
					valError := goErr.New(textBadDigest)
					return image.ID, &DockerError{errOpValidate, valError, valError.Error()}
				}
			}
		}
	}
	return "", nil
}

// TagImage - locally retag the downloaded images
func TagImage(source, tag string) *DockerError {
	out, err := exec.Command("docker", "tag", source, tag).Output()
	if err != nil {
		return &DockerError{errOpImageTag, err, err.Error()}
	}

	output := string(out[:])
	fmt.Println(output)
	return nil
}

// CheckContainerStatus of Codewind running/stopped
func CheckContainerStatus() (bool, *DockerError) {
	containerArr := [2]string{}
	containerArr[0] = "codewind-pfe"
	containerArr[1] = "codewind-performance"

	containers, err := GetContainerList()
	if err != nil {
		return false, err
	}

	containerCount := 0
	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				containerCount++
			}
		}
	}
	if containerCount >= 2 {
		return true, nil
	}
	return false, nil
}

// CheckImageStatus of Codewind installed/uninstalled
func CheckImageStatus() (bool, *DockerError) {
	imageArr := [2]string{}
	imageArr[0] = "eclipse/codewind-pfe"
	imageArr[1] = "eclipse/codewind-performance"

	images, err := GetImageList()
	if err != nil {
		return false, err
	}

	imageCount := 0
	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		for _, key := range imageArr {
			if strings.HasPrefix(imageRepo, key) {
				imageCount++
			}
		}
	}
	if imageCount < 2 {
		return false, nil
	}
	return true, nil
}

// CheckImageTag returns false if codewind images with given tag don't exist
func CheckImageTag(tag string) (bool, *DockerError) {
	tags, err := GetImageTags()
	if err != nil {
		return false, err
	}
	if !StringInSlice(tag, tags) {
		return false, nil
	}
	return true, nil
}

// RemoveImage of Codewind and project
func RemoveImage(imageID string) *DockerError {
	cmd := exec.Command("docker", "rmi", imageID, "-f")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return &DockerError{errOpImageRemove, err, err.Error()}
	}
	return nil
}

// GetContainerList from docker
func GetContainerList() ([]types.Container, *DockerError) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return nil, &DockerError{errOpClientCreate, err, err.Error()}
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, &DockerError{errOpContainerList, err, err.Error()}
	}

	return containers, nil
}

// GetImageList from docker
func GetImageList() ([]types.ImageSummary, *DockerError) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return nil, &DockerError{errOpClientCreate, err, err.Error()}
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, &DockerError{errOpImageList, err, err.Error()}
	}

	return images, nil
}

// GetNetworkList from docker
func GetNetworkList() ([]types.NetworkResource, *DockerError) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return nil, &DockerError{errOpClientCreate, err, err.Error()}
	}

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, &DockerError{errOpNetworkList, err, err.Error()}
	}

	return networks, nil
}

// StopContainer will stop only codewind containers
func StopContainer(container types.Container) *DockerError {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return &DockerError{errOpClientCreate, err, err.Error()}
	}

	// Check if the container will remove after it is stopped
	isAutoRemoved, isAutoRemovedErr := getContainerAutoRemovePolicy(container.ID)
	if isAutoRemovedErr != nil {
		errors.CheckErr(err, 108, "")
	}

	// Stop the running container
	if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
		return &DockerError{errOpContainerError, err, err.Error()}
	}

	if !isAutoRemoved {
		// Remove the container so it isnt lingering in the background
		if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			return &DockerError{errOpContainerError, err, err.Error()}
		}
	}
	return nil
}

// getContainerAutoRemovePolicy will get the auto remove policy of a given container
func getContainerAutoRemovePolicy(containerID string) (bool, *DockerError) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return false, &DockerError{errOpClientCreate, err, err.Error()}
	}

	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, &DockerError{errOpContainerInspect, err, err.Error()}
	}

	return containerInfo.HostConfig.AutoRemove, nil
}

// RemoveNetwork will remove docker network
func RemoveNetwork(network types.NetworkResource) *DockerError {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return &DockerError{errOpClientCreate, err, err.Error()}
	}

	if err := cli.NetworkRemove(ctx, network.ID); err != nil {
		return &DockerError{errOpContainerError, err, err.Error()}
	}
	return nil
}

// GetPFEHostAndPort will return the current hostname and port that PFE is running on
func GetPFEHostAndPort() (string, string, error) {
	// on Che, can assume PFE is always on localhost:9090
	if os.Getenv("CHE_API_EXTERNAL") != "" {
		return "localhost", "9090", nil
	}

	containerStatus, err := CheckContainerStatus()

	if err != nil {
		return "", "", err
	}

	if !containerStatus {
		return "", "", nil
	}

	containerList, err := GetContainerList()
	for _, container := range containerList {
		if strings.HasPrefix(container.Image, "codewind-pfe") {
			for _, port := range container.Ports {
				if port.PrivatePort == internalPFEPort {
					return port.IP, strconv.Itoa(int(port.PublicPort)), nil
				}
			}
		}
	}
	return "", "", nil
}

// GetImageTags of Codewind images
func GetImageTags() ([]string, *DockerError) {
	imageArr := [2]string{}
	imageArr[0] = "eclipse/codewind-pfe"
	imageArr[1] = "eclipse/codewind-performance"
	tagArr := []string{}
	images, err := GetImageList()
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		imageTags := strings.Join(image.RepoTags, " ")
		for _, key := range imageArr {
			if strings.HasPrefix(imageRepo, key) || strings.HasPrefix(imageTags, key) {
				if len(image.RepoTags) > 0 {
					tag := image.RepoTags[0]
					tag = strings.Split(tag, ":")[1]
					tagArr = append(tagArr, tag)
				} else {
					log.Println("No tag available. Defaulting to ''")
					tagArr = append(tagArr, "")
				}
			}
		}
	}

	tagArr = RemoveDuplicateEntries(tagArr)
	return tagArr, nil
}

// IsTCPPortAvailable checks to find the next available port and returns it
func IsTCPPortAvailable(minTCPPort int, maxTCPPort int) (bool, string) {
	var status string
	for port := minTCPPort; port < maxTCPPort; port++ {
		conn, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			log.Println("Unable to connect to port", port, ":", err)
		} else {
			status = "Port " + strconv.Itoa(port) + " Available"
			fmt.Println(status)
			conn.Close()
			return true, strconv.Itoa(port)
		}
	}
	return false, ""
}

// DetermineDebugPortForPFE determines a debug port to use for PFE based on the external PFE port
func DetermineDebugPortForPFE() (pfeDebugPort string) {
	_, debugPort := IsTCPPortAvailable(minDebugPort, maxDebugPort)
	return debugPort
}

// GetContainerTags of the Codewind version(s) currently running
func GetContainerTags() ([]string, *DockerError) {
	containerArr := [2]string{}
	containerArr[0] = "codewind-pfe"
	containerArr[1] = "codewind-performance"
	tagArr := []string{}

	containers, err := GetContainerList()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				tag := strings.Split(container.Image, ":")[1]
				tagArr = append(tagArr, tag)
			}
		}
	}

	tagArr = RemoveDuplicateEntries(tagArr)
	return tagArr, nil
}
