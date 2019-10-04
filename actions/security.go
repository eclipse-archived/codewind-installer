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
	"encoding/json"
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
	registeredClient, err := security.SecClientGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if registeredClient != nil {
		utils.PrettyPrintJSON(registeredClient)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(0)
}

// SecurityClientGetSecret : Retrieve a client secret from Keycloak
func SecurityClientGetSecret(c *cli.Context) {
	registeredClientSecret, err := security.SecClientGetSecret(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if registeredClientSecret != nil {
		utils.PrettyPrintJSON(registeredClientSecret)
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
	registeredUser, err := security.SecUserGet(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if registeredUser != nil {
		utils.PrettyPrintJSON(registeredUser)
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "Not found"})
	os.Exit(0)
}

// SecurityUserSetPassword : Set a users password in Keycloak
func SecurityUserSetPassword(c *cli.Context) {
	err := security.SecUserSetPW(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	utils.PrettyPrintJSON(security.Result{Status: "OK"})
	os.Exit(0)
}

// SecurityKeyUpdate : Creates or updates a key in the platforms keyring
func SecurityKeyUpdate(c *cli.Context) {
	err := security.SecKeyUpdate(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(security.Result{Status: "OK"})
	fmt.Println(string(response))
	os.Exit(0)
}

// SecurityKeyValidate : Checks the key is available in the platform keyring
func SecurityKeyValidate(c *cli.Context) {
	_, err := security.SecKeyGetSecret(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(security.Result{Status: "OK"})
	fmt.Println(string(response))
	os.Exit(0)
}
