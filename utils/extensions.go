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
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type (
	// ExtensionCommand represents a command defined by a project extension
	ExtensionCommand struct {
		Name    string   `json:"name"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
)

// RunCommand runs a command defined by an extension
func RunCommand(projectPath string, command ExtensionCommand) error {
	cwd, err := os.Executable()
	if err != nil {
		log.Println("There was a problem with locating the command directory")
		return err
	}
	installerPath := filepath.Dir(cwd)
	commandName := filepath.Base(command.Command) // prevent path traversal
	commandBin := filepath.Join(installerPath, commandName)
	cmd := exec.Command(commandBin, command.Args...)
	cmd.Dir = projectPath
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		log.Println("There was a problem running the command:", commandName)
		return err
	}
	log.Printf("Please wait while the project is initialized... %s", output.String())
	cmd.Wait()
	log.Println(output.String()) // Wait to finish execution, so we can read all output
	return nil
}
