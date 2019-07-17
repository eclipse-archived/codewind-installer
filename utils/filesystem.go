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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eclipse/codewind-installer/errors"
	"gopkg.in/yaml.v3"
)

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
