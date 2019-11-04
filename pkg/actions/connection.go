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

	"github.com/eclipse/codewind-installer/pkg/utils/connections"
	"github.com/urfave/cli"
)

// ConnectionAddToList : Add new connection to the connections config file and returns the ID of the added entry
func ConnectionAddToList(c *cli.Context) {
	connection, err := connections.AddConnectionToList(http.DefaultClient, c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	type Result struct {
		Status        string `json:"status"`
		StatusMessage string `json:"status_message"`
		ConID         string `json:"id"`
	}

	response, _ := json.Marshal(Result{Status: "OK", StatusMessage: "Connection added", ConID: strings.ToUpper(connection.ID)})
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionGetByID : Get connection by its id
func ConnectionGetByID(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	connection, err := connections.GetConnectionByID(connectionID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(connection)
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionRemoveFromList : Removes a connection from the connections config file
func ConnectionRemoveFromList(c *cli.Context) {
	err := connections.RemoveConnectionFromList(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(connections.Result{Status: "OK", StatusMessage: "Connection removed"})
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionListAll : Fetch all connections
func ConnectionListAll() {
	allConnections, err := connections.GetConnectionsConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(allConnections)
	fmt.Println(string(response))
	os.Exit(0)
}

// ConnectionResetList : Reset to a single default local connection
func ConnectionResetList() {
	err := connections.ResetConnectionsFile()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	response, _ := json.Marshal(connections.Result{Status: "OK", StatusMessage: "Connection list reset"})
	fmt.Println(string(response))
	os.Exit(0)
}
