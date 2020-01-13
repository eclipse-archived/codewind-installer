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

package remote

import (
	"errors"
	"flag"
	"net/http"

	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/utils"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// SetupKeycloak : sets up keycloak with a realm, client and user
func SetupKeycloak(codewindInstance Codewind, deployOptions *DeployOptions) error {

	useExistingKeycloak := false
	if deployOptions.KeycloakURL != "" {
		useExistingKeycloak = true
	}

	// Access role to be created and added to user account
	accessRoleName := "codewind-" + codewindInstance.WorkspaceID

	// Construct keycloak authentication URL or use the supplied flag
	authURL := KeycloakPrefix + codewindInstance.Ingress
	if deployOptions.KeycloakTLSSecure {
		authURL = "https://" + authURL
	} else {
		authURL = "http://" + authURL
	}
	if useExistingKeycloak {
		authURL = deployOptions.KeycloakURL
	}

	// construct the Gatekeeper URL
	gateKeeperProtocol := "http://"
	if deployOptions.GateKeeperTLSSecure {
		gateKeeperProtocol = "https://"
	}
	gatekeeperPublicURL := gateKeeperProtocol + GatekeeperPrefix + codewindInstance.Ingress

	// Wait for the Keycloak service to respond
	logr.Infoln("Waiting for Keycloak to start")
	startErr := utils.WaitForService(authURL, 200, 500)
	if startErr != nil {
		return errors.New("Keycloak did not start in a reasonable about of time")
	}

	logr.Infoln("Configuring Keycloak...")
	logr.Infoln(authURL)

	// Authenticate to get an admin access token
	flagSet := flag.NewFlagSet("authentication", 0)
	flagSet.String("host", authURL, "doc")
	flagSet.String("realm", "master", "doc")
	flagSet.String("username", deployOptions.KeycloakUser, "doc")
	flagSet.String("password", deployOptions.KeycloakPassword, "doc")
	flagSet.String("client", "admin-cli", "doc")
	c := cli.NewContext(nil, flagSet, nil)

	tokens, secErr := security.SecAuthenticate(http.DefaultClient, c, "", "")
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	// No additional Keycloak config required for a Keycloak only deployment
	if deployOptions.KeycloakOnly {
		return nil
	}

	secErr = configureKeycloakRealm(deployOptions, authURL, tokens)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	secErr = configureKeycloakClient(deployOptions, authURL, tokens, gatekeeperPublicURL)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}
	secErr = configureKeycloakAccessRole(deployOptions, authURL, tokens, accessRoleName)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	secErr = configureKeycloakUser(deployOptions, authURL, tokens)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	secErr = grantUserAccessToDeployment(deployOptions, authURL, tokens, accessRoleName)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}

	registeredSecret, secErr := fetchClientSecret(deployOptions, authURL, tokens)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return secErr.Err
	}
	deployOptions.ClientSecret = registeredSecret.Secret
	return nil
}

func configureKeycloakRealm(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken) *security.SecError {
	// Check if realm is already registered
	realm, _ := security.SecRealmGet(authURL, tokens.AccessToken, deployOptions.KeycloakRealm)
	if realm != nil && realm.ID != "" {
		logr.Infof("Updating existing Keycloak realm '%v'", realm.DisplayName)
	} else {
		// Create a new realm
		logr.Infoln("Creating Keycloak realm")
		realmFlagset := flag.NewFlagSet("setupRealm", 0)
		realmFlagset.String("host", authURL, "doc")
		realmFlagset.String("newrealm", deployOptions.KeycloakRealm, "doc")
		realmFlagset.String("accesstoken", tokens.AccessToken, "doc")
		c := cli.NewContext(nil, realmFlagset, nil)
		secErr := security.SecRealmCreate(c)
		if secErr != nil {
			return secErr
		}
	}
	return nil
}

func configureKeycloakClient(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken, gatekeeperPublicURL string) *security.SecError {

	// Check if the client is already registered
	logr.Infof("Checking for Keycloak client '%v'", deployOptions.KeycloakClient)
	clientFlagset := flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("clientid", deployOptions.KeycloakClient, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c := cli.NewContext(nil, clientFlagset, nil)
	registeredClient, _ := security.SecClientGet(c)
	if registeredClient != nil && registeredClient.ID != "" {
		logr.Infof("Updating existing Keycloak client '%v'", registeredClient.Name)
		secErr := security.SecClientAppendURL(c, gatekeeperPublicURL)
		if secErr != nil {
			return secErr
		}
	} else {
		// Create a new client
		logr.Infoln("Creating Keycloak client")
		clientFlagset := flag.NewFlagSet("setupClient", 0)
		clientFlagset.String("host", authURL, "doc")
		clientFlagset.String("redirect", gatekeeperPublicURL+"/*", "doc")
		clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
		clientFlagset.String("newclient", deployOptions.KeycloakClient, "doc")
		clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
		c := cli.NewContext(nil, clientFlagset, nil)
		secErr := security.SecClientCreate(c)
		if secErr != nil {
			return secErr
		}
	}
	return nil
}

func configureKeycloakAccessRole(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken, accessRoleName string) *security.SecError {
	// Create a new access role for this deployment
	logr.Infof("Creating access role '%v' in realm '%v'", accessRoleName, deployOptions.KeycloakRealm)
	clientFlagset := flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("role", accessRoleName, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c := cli.NewContext(nil, clientFlagset, nil)
	secErr := security.SecRoleCreate(c)
	if secErr != nil {
		return secErr
	}
	return nil
}

func configureKeycloakUser(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken) *security.SecError {

	// Check if user is already registered
	clientFlagset := flag.NewFlagSet("setupUser", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c := cli.NewContext(nil, clientFlagset, nil)
	registeredUser, secErr := security.SecUserGet(c)
	if secErr == nil && registeredUser != nil {
		logr.Infof("Existing Keycloak user '%v' found, skipping user-create & skipping password reset", registeredUser.Username)
	} else {
		// Create an initial user
		logr.Infoln("Creating Keycloak initial user")
		userCreateFlagset := flag.NewFlagSet("createUser", 0)
		userCreateFlagset.String("host", authURL, "doc")
		userCreateFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
		userCreateFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
		userCreateFlagset.String("accesstoken", tokens.AccessToken, "doc")
		c := cli.NewContext(nil, userCreateFlagset, nil)
		secErr := security.SecUserCreate(c)
		if secErr != nil {
			return secErr
		}

		// Create an initial user password
		logr.Infoln("Updating Keycloak user password")
		userPassFlagset := flag.NewFlagSet("updateUser", 0)
		userPassFlagset.String("host", authURL, "doc")
		userPassFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
		userPassFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
		userPassFlagset.String("newpw", deployOptions.KeycloakDevPassword, "doc")
		userPassFlagset.String("accesstoken", tokens.AccessToken, "doc")
		c = cli.NewContext(nil, userPassFlagset, nil)
		secErr = security.SecUserSetPW(c)
		if secErr != nil {
			return secErr
		}
	}
	return nil
}

// Grant the user access to this Deployment
func grantUserAccessToDeployment(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken, accessRoleName string) *security.SecError {
	logr.Printf("Grant '%v' access to this deployment ", deployOptions.KeycloakDevUser)
	clientFlagset := flag.NewFlagSet("setupClient", 0)
	clientFlagset.String("host", authURL, "doc")
	clientFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientFlagset.String("role", accessRoleName, "doc")
	clientFlagset.String("accesstoken", tokens.AccessToken, "doc")
	clientFlagset.String("name", deployOptions.KeycloakDevUser, "doc")
	c := cli.NewContext(nil, clientFlagset, nil)
	secErr := security.SecUserAddRole(c)
	if secErr != nil {
		return secErr
	}
	return nil
}

// fetchClientSecret : Load client secret
func fetchClientSecret(deployOptions *DeployOptions, authURL string, tokens *security.AuthToken) (*security.RegisteredClientSecret, *security.SecError) {
	logr.Infoln("Fetching client secret")
	clientSecFlagset := flag.NewFlagSet("getClientSecret", 0)
	clientSecFlagset.String("host", authURL, "doc")
	clientSecFlagset.String("realm", deployOptions.KeycloakRealm, "doc")
	clientSecFlagset.String("clientid", deployOptions.KeycloakClient, "doc")
	clientSecFlagset.String("accesstoken", tokens.AccessToken, "doc")
	c := cli.NewContext(nil, clientSecFlagset, nil)
	registeredSecret, secErr := security.SecClientGetSecret(c)
	if secErr != nil {
		utils.PrettyPrintJSON(secErr)
		return nil, secErr
	}

	return registeredSecret, nil
}
