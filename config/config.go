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
	"github.com/eclipse/codewind-installer/utils"
)

var hostname, port = utils.GetPFEHostAndPort()

// PFEHostname is the hostname at which PFE is running, e.g. "127.0.0.1"
var PFEHostname = hostname

// PFEPort is the port on which PFE is running, e.g. "9090"
var PFEPort = port

// PFEHost is the host at which PFE is running, e.g. "127.0.0.1:9090"
var PFEHost = hostname + ":" + port

// PFEOrigin is the origin from which PFE is running, e.g. "http://127.0.0.1:9090"
var PFEOrigin = "http://" + PFEHost

// PFEApiRoute is the API route at which the PFE REST API can be accessed, e.g. "http://127.0.0.1:9090/api/v1/"
var PFEApiRoute = PFEOrigin + "/api/v1/"
