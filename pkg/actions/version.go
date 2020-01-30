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
	"text/tabwriter"

	"github.com/eclipse/codewind-installer/pkg/appconstants"
	"github.com/eclipse/codewind-installer/pkg/config"
	"github.com/eclipse/codewind-installer/pkg/connections"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/remote"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// GetVersions : Gets versions of Codewind containers
func GetVersions(c *cli.Context) {
	if c.Bool("all") {
		GetAllConnectionVersions()
	} else {
		GetSingleConnectionVersion(c)
	}
}

// GetSingleConnectionVersion : Gets the cwctl and container versions for a single connection
func GetSingleConnectionVersion(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))

	conInfo, conInfoErr := connections.GetConnectionByID(connectionID)
	if conInfoErr != nil {
		HandleConnectionError(conInfoErr)
		os.Exit(1)
	}

	conURL, conErr := config.PFEOriginFromConnection(conInfo)
	if conErr != nil {
		HandleConfigError(conErr)
		os.Exit(1)
	}

	containerVersions, err := apiroutes.GetContainerVersions(conURL, appconstants.VersionNum, conInfo, http.DefaultClient)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if printAsJSON {
		json, _ := json.Marshal(containerVersions)
		fmt.Println(string(json))
	} else {
		var tableContent []string
		tableContent = append(tableContent, "CWCTL VERSION: "+containerVersions.CwctlVersion+"\n")
		tableContent = append(tableContent, "CONNECTION ID \tPFE VERSION\tPERFORMANCE VERSION\tGATEKEEPER VERSION")
		tableContent = append(tableContent, connectionID+"\t"+containerVersions.PFEVersion+"\t"+containerVersions.PerformanceVersion+"\t"+containerVersions.GatekeeperVersion)

		PrintTable(tableContent)
	}
}

// GetAllConnectionVersions : Gets the cwctl and container versions for all connections
func GetAllConnectionVersions() {
	connections, getConnectionsErr := connections.GetAllConnections()
	if getConnectionsErr != nil {
		HandleConnectionError(getConnectionsErr)
		os.Exit(1)
	}

	containerVersionsList, err := apiroutes.GetAllContainerVersions(connections, appconstants.VersionNum, http.DefaultClient)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if printAsJSON {
		json, _ := json.Marshal(containerVersionsList)
		fmt.Println(string(json))
	} else {
		var tableContent []string
		tableContent = append(tableContent, "CWCTL VERSION: "+containerVersionsList.CwctlVersion+"\n")
		tableContent = append(tableContent, "CONNECTION ID \tPFE VERSION\tPERFORMANCE VERSION\tGATEKEEPER VERSION")
		for conID, con := range containerVersionsList.Connections {
			tableContent = append(tableContent, conID+"\t"+con.PFEVersion+"\t"+con.PerformanceVersion+"\t"+con.GatekeeperVersion)
		}

		if len(containerVersionsList.ConnectionErrors) > 0 {
			tableContent = append(tableContent, "\nSOME ERRORS WHILE DETECTING CONNECTION VERSIONS")
			tableContent = append(tableContent, "CONNECTION ID \tERROR")
			for conID, conErr := range containerVersionsList.ConnectionErrors {
				tableContent = append(tableContent, conID+"\t"+conErr.Error())
			}
		}

		PrintTable(tableContent)
	}
}

// RemoteListAll prints information for all remote installations in the given namespace
func RemoteListAll(c *cli.Context) {
	namespace := c.String("namespace")
	remoteInstalls, err := remote.GetExistingDeployments(namespace)
	if err != nil {
		HandleRemInstError(err)
		os.Exit(1)
	}
	if printAsJSON {
		utils.PrettyPrintJSON(remoteInstalls)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintln(w, "Workspace ID \tNamespace \tVersion \tInstall Date \tAuth Realm \tURL")
		for _, install := range remoteInstalls {
			fmt.Fprintln(w, install.WorkspaceID+"\t"+install.Namespace+"\t"+install.Version+"\t"+install.InstallDate+"\t"+install.CodewindAuthRealm+"\t"+install.CodewindURL)
		}
		fmt.Fprintln(w)
		w.Flush()
	}
	os.Exit(0)
}
