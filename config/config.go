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
	"os"

	"github.com/eclipse/codewind-installer/pkg/utils"
)

// PFEOrigin is the origin from which PFE is running, e.g. "http://127.0.0.1:9090"
func PFEOrigin() string {
	hostname, port := utils.GetPFEHostAndPort()

	val, ok := os.LookupEnv("CHE_API_EXTERNAL")
	if ok && (val != "") {
		return "https://" + hostname + ":" + port
	}

	return "http://" + hostname + ":" + port
}
