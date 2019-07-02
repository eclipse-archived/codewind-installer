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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/eclipse/codewind-installer/errors"
	"github.com/moby/moby/client"
	"gopkg.in/yaml.v3"
)

// docker-compose yaml data
var data = `
version: 2
services:
 codewind-pfe:
  image: ${REPOSITORY}codewind-pfe${PLATFORM}:${TAG}
  container_name: codewind-pfe
  user: root
  environment: ["HOST_WORKSPACE_DIRECTORY=${WORKSPACE_DIRECTORY}","CONTAINER_WORKSPACE_DIRECTORY=/codewind-workspace","HOST_OS=${HOST_OS}","CODEWIND_VERSION=${TAG}","PERFORMANCE_CONTAINER=codewind-performance${PLATFORM}:${TAG}"]
  depends_on: [codewind-performance]
  ports: ["127.0.0.1:9090:9090"]
  volumes: ["/var/run/docker.sock:/var/run/docker.sock","${WORKSPACE_DIRECTORY}:/codewind-workspace"]
  networks: [network]
 codewind-performance:
  image: codewind-performance${PLATFORM}:${TAG}
  ports: ["127.0.0.1:9095:9095"]
  container_name: codewind-performance
  volumes: ["/var/run/docker.sock:/var/run/docker.sock","${WORKSPACE_DIRECTORY}:/codewind-workspace"]
  networks: [network]
networks:
  network:
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
	NETWORK struct {
		Network map[string]string `yaml:"network"`
	} `yaml:"networks"`
}

// CreateTempFile in the same directory as the binary for docker compose
func CreateTempFile(tempFilePath string) bool {

	var _, err = os.Stat(tempFilePath)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(tempFilePath)
		errors.CheckErr(err, 201, "")
		defer file.Close()

		fmt.Println("==> created file", tempFilePath)
		return true
	}
	return false
}

// WriteToComposeFile the contents of the docker compose yaml
func WriteToComposeFile(tempFilePath string, debug bool) bool {
	if tempFilePath == "" {
		return false
	}

	dataStruct := Compose{}

	unmarshDataErr := yaml.Unmarshal([]byte(data), &dataStruct)
	errors.CheckErr(unmarshDataErr, 202, "")

	marshalledData, err := yaml.Marshal(&dataStruct)
	errors.CheckErr(err, 203, "")

	if debug == true {
		fmt.Printf("==> "+tempFilePath+" structure is: \n%s\n\n", string(marshalledData))
	} else {
		fmt.Println("==> environment structure written to " + tempFilePath)
	}

	err = ioutil.WriteFile(tempFilePath, marshalledData, 0644)
	errors.CheckErr(err, 204, "")
	return true
}

// DockerCompose to set up the Codewind environment
func DockerCompose(tag string) {

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
		os.Setenv("WORKSPACE_DIRECTORY", "C:\\codewind-workspace")
	} else {
		os.Setenv("WORKSPACE_DIRECTORY", home+"/codewind-workspace")
	}
	os.Setenv("HOST_OS", GOOS)
	os.Setenv("COMPOSE_PROJECT_NAME", "codewind")

	cmd := exec.Command("docker-compose", "-f", "installer-docker-compose.yaml", "up", "-d")
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		DeleteTempFile("installer-docker-compose.yaml")
		errors.CheckErr(err, 101, "Is docker-compose installed?")
	}
	fmt.Printf("Please wait whilst containers initialize... %s \n", output.String())
	cmd.Wait()
	fmt.Printf(output.String()) // Wait to finish execution, so we can read all output

	if strings.Contains(output.String(), "ERROR") || strings.Contains(output.String(), "error") {
		DeleteTempFile("installer-docker-compose.yaml")
		os.Exit(1)
	}

	if strings.Contains(output.String(), "The image for the service you're trying to recreate has been removed") {
		DeleteTempFile("installer-docker-compose.yaml")
		os.Exit(1)
	}
}

// DeleteTempFile once the the Codewind environment has been created
func DeleteTempFile(tempFilePath string) (boolean bool, err error) {

	var _, file = os.Stat(tempFilePath)

	if os.IsNotExist(file) {
		errors.CheckErr(file, 206, "No files to delete")
		return false, file
	}

	os.Remove(tempFilePath)
	fmt.Println("==> finished deleting file " + tempFilePath)
	return true, nil
}

// PullImage - pull pfe/performance/initialize images from artifactory
func PullImage(image string, auth string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	var codewindOut io.ReadCloser
	if auth == "" {
		codewindOut, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})
	} else {
		codewindOut, err = cli.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: auth})
	}

	errors.CheckErr(err, 100, "")

	defer codewindOut.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(codewindOut, os.Stderr, termFd, isTerm, nil)
}

// TagImage - locally retag the downloaded images
func TagImage(source, tag string) {
	out, err := exec.Command("docker", "tag", source, tag).Output()
	errors.CheckErr(err, 102, "Image Tagging Failed")

	output := string(out[:])
	fmt.Println(output)
}

// PingHealth - pings environment api over a 15 second to check if containers started
func PingHealth(healthEndpoint string) bool {
	var started = false
	fmt.Println("Waiting for Codewind to start")
	for i := 0; i < 120; i++ {
		resp, err := http.Get(healthEndpoint)
		if err != nil {
			fmt.Printf(".")
		} else {
			if resp.StatusCode == 200 {
				fmt.Println("\nHTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				fmt.Println("Codewind successfully started")
				started = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	if started != true {
		log.Fatal("Codewind containers are taking a while to start. Please check the container logs and/or restart Codewind")
	}
	return started
}

// CheckContainerStatus of Codewind running/stopped
func CheckContainerStatus() bool {
	var containerStatus = false
	containerArr := [2]string{}
	containerArr[0] = "codewind-pfe"
	containerArr[1] = "codewind-performance"

	containers := GetContainerList()

	containerCount := 0
	for _, container := range containers {
		for _, key := range containerArr {
			if strings.HasPrefix(container.Image, key) {
				containerCount++
			}
		}
	}
	if containerCount >= 2 {
		containerStatus = true
	} else {
		containerStatus = false
	}
	return containerStatus
}

// CheckImageStatus of Codewind installed/uninstalled
func CheckImageStatus() bool {
	var imageStatus = false
	imageArr := [6]string{}
	imageArr[0] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe"
	imageArr[1] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-performance"
	imageArr[2] = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-initialize"
	imageArr[3] = "ibmcom/codewind-pfe"
	imageArr[4] = "ibmcom/codewind-performance"
	imageArr[5] = "ibmcom/codewind-initialize"

	images := GetImageList()

	imageCount := 0
	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		for _, key := range imageArr {
			if strings.HasPrefix(imageRepo, key) {
				imageCount++
			}
		}
	}
	if imageCount >= 3 {
		imageStatus = true
	}

	return imageStatus
}

// RemoveImage of Codewind and project
func RemoveImage(imageID string) {
	cmd := exec.Command("docker", "rmi", imageID, "-f")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	errors.CheckErr(err, 105, "Failed to remove image - Please make sure all containers are stopped")
}

// GetContainerList from docker
func GetContainerList() []types.Container {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	errors.CheckErr(err, 107, "")

	return containers
}

// GetImageList from docker
func GetImageList() []types.ImageSummary {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	errors.CheckErr(err, 109, "")

	return images
}

// GetNetworkList from docker
func GetNetworkList() []types.NetworkResource {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	errors.CheckErr(err, 110, "")

	return networks
}

// StopContainer will stop only codewind containers
func StopContainer(container types.Container) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	// Stop the running container
	if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
		errors.CheckErr(err, 108, "")
	}

	// Remove the container so it isnt lingering in the background
	if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
		errors.CheckErr(err, 108, "")
	}
}

// RemoveNetwork will remove docker network
func RemoveNetwork(network types.NetworkResource) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	errors.CheckErr(err, 200, "")

	if err := cli.NetworkRemove(ctx, network.ID); err != nil {
		errors.CheckErr(err, 111, "Cannot remove "+network.Name+". Use 'stop-all' flag to ensure all containers have been terminated")
	}
}
