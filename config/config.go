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

	"github.com/eclipse/codewind-installer/utils"
)

// PFEHost is the host at which PFE is running, e.g. "127.0.0.1:9090"
func PFEHost() string {
	hostname, port, err := utils.GetPFEHostAndPort()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	return hostname + ":" + port
}

// PFEOrigin is the origin from which PFE is running, e.g. "http://127.0.0.1:9090"
func PFEOrigin() string {
	return "http://" + PFEHost()
}

// PFEApiRoute is the API route at which the PFE REST API can be accessed, e.g. "http://127.0.0.1:9090/api/v1/"
func PFEApiRoute() string {
	return PFEOrigin() + "/api/v1/"
}
