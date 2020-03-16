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

package actions

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/project"
	"github.com/eclipse/codewind-installer/pkg/remote"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//InstallCommand to pull images from dockerhub
func InstallCommand(c *cli.Context) {
	tag := c.String("tag")

	imageArr := [2]string{
		"docker.io/eclipse/codewind-pfe-amd64:" + tag,
		"docker.io/eclipse/codewind-performance-amd64:" + tag,
	}

	// creates a new docker client, which is passed into the functions that interact with the docker API
	dockerClient, dockerErr := docker.NewDockerClient()
	if dockerErr != nil {
		HandleDockerError(dockerErr)
		os.Exit(1)
	}

	for i := 0; i < len(imageArr); i++ {
		docker.PullImage(dockerClient, imageArr[i], printAsJSON)
		imageID, dockerError := docker.ValidateImageDigest(dockerClient, imageArr[i])

		if dockerError != nil {
			logr.Tracef("%v checksum validation failed. Trying to pull image again", imageArr[i])
			// remove bad image
			docker.RemoveImage(imageID)

			// pull image again
			docker.PullImage(dockerClient, imageArr[i], printAsJSON)

			// validate the new image
			_, dockerError = docker.ValidateImageDigest(dockerClient, imageArr[i])

			if dockerError != nil {
				if printAsJSON {
					fmt.Println(dockerError)
				} else {
					logr.Errorf("Validation of image '%v' checksum failed - Removing image", imageArr[i])
				}
				// Clean up the second bad image
				docker.RemoveImage(imageID)
				os.Exit(1)
			}
		}
	}

	fmt.Println("Image Install Successful")
}

// DoRemoteInstall : Deploy a remote PFE and support containers
func DoRemoteInstall(c *cli.Context) {

	// Since remote will always use Self Signed Certificates initially, turn on insecure flag
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	printAsJSON := c.GlobalBool("json")

	session := c.String("session")
	if session == "" {
		session = strings.ToUpper(strconv.FormatInt(utils.CreateTimestamp(), 36))
	}

	if c.Int("pvcsize") < 0 || c.Int("pvcsize") > 999 {
		logr.Error("Codewind PVC size should be between 1 and 999 GB")
		os.Exit(1)
	}

	codewindPVCSize := c.Int("pvcsize")
	if codewindPVCSize < 1 {
		codewindPVCSize = 1
	}

	keycloakHost := c.String("kurl")
	if keycloakHost != "" {
		u, err := url.Parse(keycloakHost)
		if err != nil {
			logr.Error("Supplied Keycloak URL is invalid")
			os.Exit(1)
		}
		keycloakHost = u.Hostname()
	}

	deployOptions := remote.DeployOptions{
		Namespace:             c.String("namespace"),
		IngressDomain:         c.String("ingress"),
		KeycloakUser:          c.String("kadminuser"),
		KeycloakPassword:      c.String("kadminpass"),
		KeycloakDevUser:       c.String("kdevuser"),
		KeycloakDevPassword:   c.String("kdevpass"),
		KeycloakRealm:         c.String("krealm"),
		KeycloakClient:        c.String("kclient"),
		KeycloakURL:           c.String("kurl"),
		KeycloakOnly:          c.Bool("konly"),
		KeycloakHost:          keycloakHost,
		GateKeeperTLSSecure:   true,
		KeycloakTLSSecure:     true,
		CodewindSessionSecret: session,
		CodewindPVCSize:       strconv.Itoa(codewindPVCSize) + "Gi",
		LogLevel:              c.GlobalString("loglevel"),
	}

	deploymentResult, remInstError := remote.DeployRemote(&deployOptions)
	if remInstError != nil {
		if printAsJSON {
			fmt.Println(remInstError.Error())
		} else {
			logr.Errorf("Error: %v - %v\n", remInstError.Op, remInstError.Desc)
		}
		os.Exit(1)
	}

	// If performing a Keycloak only install,  display just the keycloak URL
	if deployOptions.KeycloakOnly {
		keycloakURL := deploymentResult.KeycloakURL
		if deployOptions.KeycloakTLSSecure {
			keycloakURL = "https://" + keycloakURL
		} else {
			keycloakURL = "http://" + keycloakURL
		}
		if printAsJSON {
			result := project.Result{Status: "OK", StatusMessage: "Keycloak Install Successful: " + keycloakURL}
			response, _ := json.Marshal(result)
			fmt.Println(string(response))
		} else {
			logr.Infoln("Keycloak is available at: " + keycloakURL)
		}
		os.Exit(0)
	}

	// We're doing a full install. Wait Gatekeeper to startup and for PFE to respond

	gatekeeperURL := deploymentResult.GatekeeperURL

	logr.Infoln("Waiting for Codewind Gatekeeper to start on " + gatekeeperURL)
	utils.WaitForService(gatekeeperURL+"/health", 200, 500)

	logr.Infoln("Waiting for Codewind PFE to start")
	utils.WaitForService(gatekeeperURL+"/api/pfe/ready", 200, 500)

	result := project.Result{Status: "OK", StatusMessage: "Install Successful: " + gatekeeperURL}
	if printAsJSON {
		response, _ := json.Marshal(result)
		fmt.Println(string(response))
	} else {
		logr.Infoln("Codewind is available at: " + gatekeeperURL)
	}
	os.Exit(0)
}
