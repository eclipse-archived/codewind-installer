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

package desktoputils

import (
	"os"
	"runtime"
)

// *** When moving to go 12 we should use func UserHomeDir() instead of this function ***

// GetHomeDir of current system
func GetHomeDir() string {
	homeDir := ""
	const GOOS string = runtime.GOOS
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
