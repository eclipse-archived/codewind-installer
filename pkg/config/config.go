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
	"errors"
	"os"

	"github.com/eclipse/codewind-installer/pkg/connections"
	"github.com/eclipse/codewind-installer/pkg/utils"
)

// ConfigError : config package errors
type ConfigError struct {
	Op   string
	Err  error
	Desc string
}

const errOpConfConNotFound = "config_connection_notfound"
const errOpConfPFEHostnamePortNotFound = "config_pfe_hostname_port_notfound"
const textHostnameOrPortNotFound = "Hostname or port for PFE not found"

// PFEOriginFromConnection is used when GetConnectionByID(conID) has already been called to stop it being run twice in one function
func PFEOriginFromConnection(connection *connections.Connection) (string, *ConfigError) {
	if connection.ID != "local" {
		return connection.URL, nil
	}
	localURL, localErr := getLocalHostnameAndPort()
	if localErr != nil {
		return "", &ConfigError{errOpConfConNotFound, localErr.Err, localErr.Desc}
	}
	return localURL, nil
}

func getLocalHostnameAndPort() (string, *ConfigError) {
	val, ok := os.LookupEnv("CHE_API_EXTERNAL")
	if ok && (val != "") {
		return "https://localhost:9090", nil
	}
	hostname, port, err := utils.GetPFEHostAndPort()
	if err != nil {
		return "", &ConfigError{errOpConfPFEHostnamePortNotFound, err, err.Desc}
	} else if hostname == "" || port == "" {
		pfeHostPortErr := errors.New(textHostnameOrPortNotFound)
		return "", &ConfigError{errOpConfPFEHostnamePortNotFound, pfeHostPortErr, textHostnameOrPortNotFound}
	}
	return "http://" + hostname + ":" + port, nil
}
