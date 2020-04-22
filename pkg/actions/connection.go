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
	"net/http"
	"os"
	"strings"

	"github.com/eclipse/codewind-installer/pkg/connections"
	logr "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ConnectionAddToList : Add new connection to the connections config file and returns the ID of the added entry
func ConnectionAddToList(c *cli.Context) {
	connection, conErr := connections.AddConnectionToList(http.DefaultClient, c)
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}

	type Result struct {
		Status        string `json:"status"`
		StatusMessage string `json:"status_message"`
		ConID         string `json:"id"`
	}

	response, _ := json.Marshal(Result{Status: "OK", StatusMessage: "Connection added", ConID: strings.ToUpper(connection.ID)})
	if printAsJSON {
		fmt.Println(string(response))
	} else {
		logr.Printf("Connection %v added successfully", strings.ToUpper(connection.ID))
	}

	os.Exit(0)
}

// ConnectionUpdate : Update an existing connection
func ConnectionUpdate(c *cli.Context) {
	connection, conErr := connections.UpdateExistingConnection(http.DefaultClient, c)
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}
	type Result struct {
		Status        string `json:"status"`
		StatusMessage string `json:"status_message"`
		ConID         string `json:"id"`
	}

	response, _ := json.Marshal(Result{Status: "OK", StatusMessage: "Connection updated", ConID: strings.ToUpper(connection.ID)})
	if conErr != nil {
		fmt.Println(string(response))
	} else {
		logr.Printf("Connection %v updated successfully", strings.ToUpper(connection.ID))
	}
	os.Exit(0)
}

// ConnectionGetByID : Get connection by its id
func ConnectionGetByID(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	connection, conErr := connections.GetConnectionByID(connectionID)
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}
	response, _ := json.Marshal(connection)
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionRemoveFromList : Removes a connection from the connections config file
func ConnectionRemoveFromList(c *cli.Context) {
	conErr := connections.RemoveConnectionFromList(c)
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}
	response, _ := json.Marshal(connections.Result{Status: "OK", StatusMessage: "Connection removed"})
	if printAsJSON {
		fmt.Println(string(response))
	} else {
		logr.Printf("Connection removed successfully")
	}
	os.Exit(0)
}

// ConnectionListAll : Fetch all connections
func ConnectionListAll(c *cli.Context) {
	allConnections, conErr := connections.GetConnectionsConfig()
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}
	response, _ := json.Marshal(allConnections)
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionResetList : Reset to a single default local connection
func ConnectionResetList(c *cli.Context) {
	conErr := connections.ResetConnectionsFile()
	if conErr != nil {
		HandleConnectionError(conErr)
		os.Exit(1)
	}
	response, _ := json.Marshal(connections.Result{Status: "OK", StatusMessage: "Connection list reset"})
	if printAsJSON {
		fmt.Println(string(response))
	} else {
		logr.Printf("Connection list reset successfully")
	}
	os.Exit(0)
}
