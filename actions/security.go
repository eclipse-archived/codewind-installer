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
	"github.com/eclipse/codewind-installer/utils/security"
	"github.com/urfave/cli"
)

// SecurityTokenGet : Authenticate and retrieve an access_token
func SecurityTokenGet(c *cli.Context) {
	auth, err := security.SecAuthenticate(c, "", "")
	if err == nil && auth != nil {
		utils.PrettyPrintJSON(auth)
	} else {
		fmt.Println(err.Error())
	}
	os.Exit(0)
}

// SecurityCreateRealm : Create a realm in Keycloak
func SecurityCreateRealm(c *cli.Context) {
	err := security.SecRealmCreate(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityClientCreate : Create a new client in Keycloak
func SecurityClientCreate(c *cli.Context) {
	err := security.SecClientCreate(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityClientGet : Retrieve a client configuration from Keycloak
func SecurityClientGet(c *cli.Context) {
	registredClient, err := security.SecClientGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if registredClient != nil {
		utils.PrettyPrintJSON(registredClient)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(0)
}

// SecurityUserCreate : Create a user in a Keycloak realm
func SecurityUserCreate(c *cli.Context) {
	err := security.SecUserCreate(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		utils.PrettyPrintJSON(security.Result{Status: "OK"})
	}
	os.Exit(0)
}

// SecurityUserGet : Retrieve the user detail from Keycloak
func SecurityUserGet(c *cli.Context) {
	registredUser, err := security.SecUserGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if registredUser != nil {
		utils.PrettyPrintJSON(registredUser)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(0)
}

// SecurityUserSetPW : Set a users password in Keycloak
func SecurityUserSetPW(c *cli.Context) {
	err := security.SecUserSetPW(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "OK"})
	os.Exit(0)
}
