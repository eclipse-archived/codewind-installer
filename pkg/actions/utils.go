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
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/docker"
	"github.com/eclipse/codewind-installer/pkg/project"
	"github.com/eclipse/codewind-installer/pkg/remote"
	logr "github.com/sirupsen/logrus"
)

// HandleDockerError prints a Docker error, in JSON format if the global flag is set and as a string if not
func HandleDockerError(err *docker.DockerError) {
	// printAsJSON is a global variable, set in commands.go
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleTemplateError prints a Template error, in JSON format if the global flag is set, and as a string if not
func HandleTemplateError(err *TemplateError) {
	// printAsJSON is a global variable, set in commands.go
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleConnectionError prints a Connection error, in JSON format if the global flag is set and as a string if not
func HandleConnectionError(err *connections.ConError) {
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleProjectError prints a Project error, in JSON format if the global flag is set and as a string if not
func HandleProjectError(err *project.ProjectError) {
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleConfigError prints a Config error, in JSON format if the global flag is set and as a string if not
func HandleConfigError(err *config.ConfigError) {
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleRemInstError prints a RemInst error, in JSON format if the global flag is set and as a string if not
func HandleRemInstError(err *remote.RemInstError) {
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// HandleRegistryError prints a Registry error, in JSON format if the global flag is set, and as a string if not
func HandleRegistryError(err *RegistryError) {
	// printAsJSON is a global variable, set in commands.go
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}

// PrintTable prints a formatted table into the terminal
func PrintTable(content []string) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	for _, line := range content {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w)
	w.Flush()
}
