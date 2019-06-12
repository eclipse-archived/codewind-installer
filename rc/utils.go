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
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
	"gopkg.in/yaml.v3"
)

var debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

// docker-compose yaml data
var data = `
version: 2
services:
 codewind-pfe:
  image: ${REPOSITORY}codewind-pfe${PLATFORM}:${TAG}
  container_name: codewind-pfe
  user: root
  environment: ["HOST_WORKSPACE_DIRECTORY=${WORKSPACE_DIRECTORY}","CONTAINER_WORKSPACE_DIRECTORY=/microclimate-workspace","HOST_OS=${HOST_OS}","TELEMETRY=${TELEMETRY}","MICROCLIMATE_VERSION=${TAG}","PERFORMANCE_CONTAINER=codewind-performance${PLATFORM}:${TAG}"]
  depends_on: [codewind-performance]
  ports: ["127.0.0.1:9090:9090"]
  volumes: ["/var/run/docker.sock:/var/run/docker.sock","${WORKSPACE_DIRECTORY}:/microclimate-workspace"]
  networks: [network]
 codewind-performance:
  image: codewind-performance${PLATFORM}:${TAG}
  ports: ["127.0.0.1:9095:9095"]
  container_name: codewind-performance
  volumes: ["/var/run/docker.sock:/var/run/docker.sock","${WORKSPACE_DIRECTORY}:/microclimate-workspace"]
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
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fmt.Println("==> created file", tempFilePath)
		return true
	}
	return false
}

// WriteToComposeFile the contents of the docker compose yaml
func WriteToComposeFile(tempFilePath string) bool {

	dataStruct := Compose{}

	unmarshDataErr := yaml.Unmarshal([]byte(data), &dataStruct)
	if unmarshDataErr != nil {
		log.Fatalf("error: %v", unmarshDataErr)
	}

	marshalledData, err := yaml.Marshal(&dataStruct)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if debug == true {
		fmt.Printf("==> "+tempFilePath+" structure is: \n%s\n\n", string(marshalledData))
	} else {
		fmt.Println("==> environment structure written to " + tempFilePath)
	}

	err = ioutil.WriteFile(tempFilePath, marshalledData, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return true
}

// DockerCompose to set up the Codewind environment
func DockerCompose() {

	// Set env variables for the docker compose file
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("failed to get home dir")
	}

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
	os.Setenv("TAG", "latest")
	if GOOS == "windows" {
		os.Setenv("WORKSPACE_DIRECTORY", home+"\\microclimate-workspace")
	} else {
		os.Setenv("WORKSPACE_DIRECTORY", home+"/microclimate-workspace")
	}
	os.Setenv("HOST_OS", GOOS)
	os.Setenv("TELEMETRY", "")
	os.Setenv("COMPOSE_PROJECT_NAME", "microclimate")

	cmd := exec.Command("docker-compose", "-f", "installer-docker-compose.yaml", "up", "-d")
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		fmt.Printf("Failed to start " + err.Error())
		os.Exit(1)
	}
	fmt.Printf("Please wait whilst containers initialize... %s \n", output.String())
	cmd.Wait()
	fmt.Printf(output.String()) // Wait to finish execution, so we can read all output
}

// DeleteTempFile once the the Codewind environment has been created
func DeleteTempFile(tempFilePath string) bool {
	// delete the installer-docker-compose.yaml file
	var err = os.Remove(tempFilePath)
	if err != nil {
		fmt.Println("No more files to delete")
	} else {
		fmt.Println("==> finished deleting file " + tempFilePath)
	}
	return true
}

// PullImage - pull pfe/performance/initialize images from artifactory
func PullImage(image string, auth string) {
	// TODO when images are in dockerhub, handle dockerhub pulls
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	if auth == "" {
		codewindOut, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			log.Fatal(err)
		}
		defer codewindOut.Close()
		io.Copy(os.Stdout, codewindOut)
	} else {
		codewindOut, err := cli.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: auth})
		if err != nil {
			log.Fatal(err)
		}
		defer codewindOut.Close()
		io.Copy(os.Stdout, codewindOut)
	}

}

// TagImage - locally retag the downloaded images
func TagImage(source, tag string) {
	out, err := exec.Command("docker", "tag", source, tag).Output()
	if err != nil {
		fmt.Println("Image Tagging Failed")
		fmt.Printf("%s", err)
	}

	output := string(out[:])
	fmt.Println(output)
}

// PingHealth - pings environment api over a 15 second to check if containers started
func PingHealth(healthEndpoint string) bool {
	var started = false
	for i := 0; i < 15; i++ {
		resp, err := http.Get(healthEndpoint)
		if err != nil {
			fmt.Println("Waiting for Codewind to start...")
		} else {
			if resp.StatusCode == 200 {
				fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				fmt.Println("Codewind successfully started")
				started = true
				break
			}
		}

		time.Sleep(1 * time.Second)
	}

	if started != true {
		fmt.Println("Codewind containers are taking a while to start. Please check the container logs and/or restart Codewind")
	}
	return started
}

// CheckContainerStatus of Codewind running/stopped
func CheckContainerStatus() bool {
	var containerStatus = false

	// Check if the Codewind containers are running
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	containerCount := 0
	for _, container := range containers {
		if strings.Contains(container.Image, "codewind") {
			containerCount++
		}
	}
	if containerCount >= 2 {
		containerStatus = true
		// fmt.Println("Codewind is installed and running")
	} else {
		containerStatus = false
	}
	return containerStatus
}

// CheckImageStatus of Codewind installed/uninstalled
func CheckImageStatus() bool {
	var imageStatus = false
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	// Check if the Codewind images are available
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	imageCount := 0
	for _, image := range images {
		imageRepo := strings.Join(image.RepoDigests, " ")
		if strings.Contains(imageRepo, "codewind") {
			imageCount++
		}
	}
	if imageCount == 3 {
		imageStatus = true
	}

	return imageStatus
}
