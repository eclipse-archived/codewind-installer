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
	"encoding/json"
	"fmt"
	"os"

	"github.com/eclipse/codewind-installer/utils/deployments"
	"github.com/urfave/cli"
)

// DeploymentAddToList : Add new deployment to the deployments config file
func DeploymentAddToList(c *cli.Context) {
	err := deployments.AddDeploymentToList(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(deployments.Result{Status: "OK", StatusMessage: "Deployment added"})
	fmt.Println(string(response))
	os.Exit(0)
}

// DeploymentRemoveFromList : Removes a deployment from the deployments config file
func DeploymentRemoveFromList(c *cli.Context) {
	err := deployments.RemoveDeploymentFromList(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(deployments.Result{Status: "OK", StatusMessage: "Deployment removed"})
	fmt.Println(string(response))
	os.Exit(0)
}

// DeploymentGetTarget : Fetch the target deployment
func DeploymentGetTarget() {
	targetDeployment, err := deployments.GetTargetDeployment()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(targetDeployment)
	fmt.Println(string(response))
	os.Exit(0)
}

// DeploymentSetTarget : Set a new deployment by ID
func DeploymentSetTarget(c *cli.Context) {
	err := deployments.SetTargetDeployment(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(deployments.Result{Status: "OK", StatusMessage: "New target set"})
	fmt.Println(string(response))
	os.Exit(0)
}

// DeploymentListAll : Fetch all deployments
func DeploymentListAll() {
	allDeployments, err := deployments.GetAllDeployments()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(allDeployments)
	fmt.Println(string(response))
	os.Exit(0)
}

// DeploymentResetList : Reset to a single default local deployment
func DeploymentResetList() {
	err := deployments.ResetDeploymentsFile()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(deployments.Result{Status: "OK", StatusMessage: "Deployment list reset"})
	fmt.Println(string(response))
	os.Exit(0)
}
