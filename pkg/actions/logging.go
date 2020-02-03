/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
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
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// LogLevels : Optionally set then display the log level information
// for a Codewind PFE container
func LogLevels(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	newLogLevel := strings.TrimSpace(strings.ToLower(c.Args().Get(0)))

	conInfo, conInfoErr := connections.GetConnectionByID(connectionID)
	if conInfoErr != nil {
		fmt.Println(conInfoErr.Err)
		os.Exit(1)
	}

	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		fmt.Println(conErr.Err)
		os.Exit(1)
	}

	if newLogLevel != "" {
		err := apiroutes.SetLogLevel(conInfo, conURL, http.DefaultClient, newLogLevel)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	loggingLevels, err := apiroutes.GetLogLevel(conInfo, conURL, http.DefaultClient)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrintJSON(loggingLevels)
}
