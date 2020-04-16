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

package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/eclipse/codewind-installer/pkg/errors"
	"gopkg.in/yaml.v2"
)

// WriteToComposeFile the contents of the docker compose yaml
func WriteToComposeFile(dockerComposeFile string, debug bool) bool {
	if dockerComposeFile == "" {
		return false
	}

	secretFileName, secretErr := writeDockerConfigSecretFile(path.Dir(dockerComposeFile))

	fmt.Println("BEFORE")
	errors.CheckErr(secretErr, 204, "")
	dataStruct := Compose{}
	fmt.Print("HERE")
	data := fmt.Sprintf(composeTemplate, secretFileName)

	unmarshDataErr := yaml.Unmarshal([]byte(data), &dataStruct)
	errors.CheckErr(unmarshDataErr, 202, "")

	if debug == true && len(dataStruct.SERVICES.PFE.Ports) > 0 {
		debugPort := DetermineDebugPortForPFE()
		// Add the debug port to the docker compose data
		dataStruct.SERVICES.PFE.Ports = append(dataStruct.SERVICES.PFE.Ports, "127.0.0.1:"+debugPort+":9777")
	}

	marshalledData, err := yaml.Marshal(&dataStruct)
	errors.CheckErr(err, 203, "")

	if debug == true {
		fmt.Printf("==> %s structure is: \n%s\n\n", dockerComposeFile, string(marshalledData))
	} else {
		fmt.Println("==> environment structure written to " + filepath.ToSlash(dockerComposeFile))
	}

	err = ioutil.WriteFile(dockerComposeFile, marshalledData, 0644)
	errors.CheckErr(err, 204, "")
	return true
}

func writeDockerConfigSecretFile(parentPath string) (string, error) {
	// TODO handle error coming from here
	dockerConfig, err := getDockerCredentials("local")
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	dockerConfigBytes, jsonErr := json.MarshalIndent(dockerConfig, "", "  ")
	if jsonErr != nil {
		return "", jsonErr
	}
	encoded := base64.StdEncoding.EncodeToString(dockerConfigBytes)
	secretFile := path.Join(parentPath, dockerConfigSecretFile)
	err = ioutil.WriteFile(secretFile, []byte(encoded), 0600)
	return secretFile, err
}

// ClearDockerConfigSecret We erase the contents rather than deleting
// the file as the docker-compose file expects the secret to be present.
func ClearDockerConfigSecret(parentPath string) error {
	// Most callers won't handle this error as this shouldn't block shutdown.
	secretFile := path.Join(parentPath, dockerConfigSecretFile)
	return ioutil.WriteFile(secretFile, []byte{}, 0600)
}

// PingHealth - pings environment api every 15 seconds to check if containers started
func PingHealth(healthEndpoint string) (bool, *DockerError) {
	var started = false
	fmt.Println("Waiting for Codewind to start")

	dockerClient, err := NewDockerClient()
	if err != nil {
		return false, err
	}

	hostname, port, err := GetPFEHostAndPort(dockerClient)
	if err != nil {
		return false, err
	}
	for i := 0; i < 120; i++ {
		resp, err := http.Get("http://" + hostname + ":" + port + healthEndpoint)
		if err != nil {
			fmt.Printf(".")
		} else {
			if resp.StatusCode == 200 {
				fmt.Println("\nHTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				fmt.Println("Codewind successfully started on http://" + hostname + ":" + port)
				started = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	if started != true {
		log.Fatal("Codewind containers are taking a while to start. Please check the container logs and/or restart Codewind")
	}
	return started, nil
}
