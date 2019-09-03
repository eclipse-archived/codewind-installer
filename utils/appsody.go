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
	"os/exec"
	"log"
)
 
 // SuccessfullyCallAppsodyInit calls Appsody Init to initialise Appsody projects and returns a boolean to indicate success
 func SuccessfullyCallAppsodyInit(projectPath string) (bool, error) {
	cmd := exec.Command("appsody", "init")
	cmd.Dir = projectPath
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		log.Println("There was a problem initializing the Appsody project: ", err, ". Project was not initialized.")
		return false, err
	}
	log.Printf("Please wait while the Appsody project is initialized... %s \n", output.String())
	cmd.Wait()
	log.Println(output.String()) // Wait to finish execution, so we can read all output
	return true, nil
 }
 