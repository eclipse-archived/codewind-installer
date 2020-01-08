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

	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
)

// HandleDockerError prints a Docker error, in JSON format if the global flag is set, and as a string if not
func HandleDockerError(err *utils.DockerError) {
	// printAsJSON is a global variable, set in commands.go
	if printAsJSON {
		fmt.Println(err.Error())
	} else {
		logr.Error(err.Desc)
	}
}
