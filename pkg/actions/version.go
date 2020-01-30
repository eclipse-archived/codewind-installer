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
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/eclipse/codewind-installer/pkg/apiroutes"
	"github.com/eclipse/codewind-installer/pkg/remote"
	"github.com/eclipse/codewind-installer/pkg/utils"
	"github.com/urfave/cli"
)

// GetVersions : Gets versions of Codewind containers
func GetVersions(c *cli.Context) {
	connectionID := strings.TrimSpace(strings.ToLower(c.String("conid")))
	containerVersions, err := apiroutes.GetContainerVersions(connectionID, http.DefaultClient)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrintJSON(containerVersions)
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
