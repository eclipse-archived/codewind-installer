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

package main

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/eclipse/codewind-installer/pkg/actions"
	"github.com/eclipse/codewind-installer/pkg/connections"
)

func main() {
	connections.InitConfigFileIfRequired()
	cheInit()
	actions.Commands()
}

func cheInit() {
	val, ok := os.LookupEnv("CHE_API_EXTERNAL")

	if ok && (val != "") {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}
