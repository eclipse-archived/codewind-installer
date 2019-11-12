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
	"os"
	"strconv"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/eclipse/codewind-installer/pkg/utils/project"
	"github.com/eclipse/codewind-installer/pkg/utils/remote"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//InstallCommand to pull images from dockerhub
func InstallCommand(c *cli.Context) {
	tag := c.String("tag")
	jsonOutput := c.Bool("json") || c.GlobalBool("json")

	imageArr := [2]string{"docker.io/eclipse/codewind-pfe-amd64:",
		"docker.io/eclipse/codewind-performance-amd64:"}

	targetArr := [2]string{"codewind-pfe-amd64:",
		"codewind-performance-amd64:"}

	for i := 0; i < len(imageArr); i++ {
		utils.PullImage(imageArr[i]+tag, jsonOutput)
		utils.TagImage(imageArr[i]+tag, targetArr[i]+tag)
	}

	fmt.Println("Image Tagging Successful")
}

// DoRemoteInstall : Deploy a remote PFE and support containers
func DoRemoteInstall(c *cli.Context) {

	// Since remote will always use Self Signed Certificates initally, turn on insecure flag
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	printAsJSON := c.GlobalBool("json")

	session := c.String("session")
	if session == "" {
		session = strings.ToUpper(strconv.FormatInt(utils.CreateTimestamp(), 36))
	}

	deployOptions := remote.DeployOptions{
		Namespace:             c.String("namespace"),
		IngressDomain:         c.String("ingress"),
		InstallKeycloak:       c.Bool("addkeycloak"),
		KeycloakUser:          c.String("kadminuser"),
		KeycloakPassword:      c.String("kadminpass"),
		KeycloakDevUser:       c.String("kdevuser"),
		KeycloakDevPassword:   c.String("kdevpass"),
		KeycloakRealm:         c.String("krealm"),
		KeycloakClient:        c.String("kclient"),
		GateKeeperTLSSecure:   true,
		KeycloakTLSSecure:     true,
		CodewindSessionSecret: session,
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
