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
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/eclipse/codewind-installer/utils"
	"github.com/urfave/cli"
)

// CloneTemplate from github
func CloneTemplate(c *cli.Context) {
	var tempPath = ""

	const GOOS string = runtime.GOOS
	if GOOS == "windows" {
		tempPath = os.Getenv("TEMP") + "\\"
	} else {
		tempPath = "/tmp/"
	}

	destination := c.String("destination")
	branch := c.String("branch")
	zipURL := utils.GetZipURL(c)
	time := time.Now().Format(time.RFC3339)
	time = strings.Replace(time, ":", "-", -1) // ":" is illegal char in windows
	tempName := tempPath + branch + "_" + time
	zipFileName := tempName + ".zip"

	// download files in zip format
	if err := utils.DownloadFile(zipFileName, zipURL); err != nil {
		log.Fatal(err)
	}

	// unzip into /tmp dir
	utils.UnZip(zipFileName, destination)

	//delete zip file
	utils.DeleteTempFile(zipFileName)
}
