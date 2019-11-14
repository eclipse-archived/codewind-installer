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

package config

import (
	"fmt"
	"os"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

type ConfigError struct {
	Op   string
	Err  error
	Desc string
}

const errOpConfConNotFound = "config_connection_notfound"

// PFEOrigin is the origin from which PFE is running, e.g. "http://127.0.0.1:9090"
func PFEOrigin(conID string) (string, *ConfigError) {
	conInfo, conErr := connections.GetConnectionByID(conID)
	var PFEURL string
	if conErr != nil {
		return PFEURL, &ConfigError{errOpConfConNotFound, conErr.Err, conErr.Error()}
	}
	if conInfo.ID != "local" {
		PFEURL = conInfo.URL
	} else {
		hostname, port := utils.GetPFEHostAndPort()
		val, ok := os.LookupEnv("CHE_API_EXTERNAL")
		if ok && (val != "") {
			PFEURL = "https://" + hostname + ":" + port
		}
		PFEURL = "http://" + hostname + ":" + port
	}
	fmt.Print(PFEURL)
	return PFEURL, nil
}
