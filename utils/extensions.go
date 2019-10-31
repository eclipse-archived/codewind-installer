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
	"strings"
)

type (
	// Extension represents a project extension defined by codewind.yaml
	Extension struct {
		ProjectType string             `json:"projectType"`
		Detection   string             `json:"detection"`
		Commands    []ExtensionCommand `json:"commands"`
		Config      ExtensionConfig    `json:"config"`
	}

	// ExtensionCommand represents a command defined by a project extension
	ExtensionCommand struct {
		Name    string   `json:"name"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	// ExtensionConfig represents a project extension's config element
	ExtensionConfig struct {
		Style string `json:"style"`
	}
)

// Run a directive on the value
func processDirective(value string, directive string) string {

	// directive for replacing extension
	if strings.HasPrefix(directive, ".") {
		ext := filepath.Ext(value)
		if ext != "" {
			value = strings.TrimSuffix(value, ext)
		}
		value += directive
	}

	return value
}

// Process an argument, substituting values and running directives as required
// syntax: $variable[,directive]
func processArg(arg string, params map[string]string) string {

	if strings.HasPrefix(arg, "$") {

		// attempt to split it into the var part and the directive part
		// find corresponding value in params using var part
		parts := strings.Split(arg, ",")
		value := params[parts[0]]

		if value != "" {
			// process any directives
			if len(parts) >= 2 {
				value = processDirective(value, parts[1])
			}
			return value
		}
	}

	return arg
}

// RunCommand runs a command defined by an extension
func RunCommand(projectPath string, command ExtensionCommand, params map[string]string) error {
	cwd, err := os.Executable()
	if err != nil {
		log.Println("There was a problem with locating the command directory")
		return err
	}
	cwctlPath := filepath.Dir(cwd)
	commandName := filepath.Base(command.Command) // prevent path traversal
	commandBin := filepath.Join(cwctlPath, commandName)

	// check for variable substitution into args
	for i := 0; i < len(command.Args); i++ {
		arg := command.Args[i]
		command.Args[i] = processArg(arg, params)
	}

	cmd := exec.Command(commandBin, command.Args...)
	cmd.Dir = projectPath
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil { // after 'Start' the program is continued and script is executing in background
		log.Println("There was a problem running the command:", commandName)
		return err
	}
	log.Printf("Please wait while the command runs... %s", output.String())
	cmd.Wait()
	log.Println(output.String()) // Wait to finish execution, so we can read all output
	return nil
}
