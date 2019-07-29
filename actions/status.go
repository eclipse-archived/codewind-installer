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

	"github.com/eclipse/codewind-installer/utils"
)

//StatusCommand to show the status
func StatusCommand() {
	if utils.CheckContainerStatus() {
		fmt.Println("Codewind is installed and running")
		os.Exit(202)
	}

	if utils.CheckImageStatus() {
		fmt.Println("Codewind is installed but not running")
		os.Exit(201)
	} else {
		fmt.Println("Codewind is not installed")
		os.Exit(200)
	}
	return
}
