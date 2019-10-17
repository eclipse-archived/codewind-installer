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

	"github.com/eclipse/codewind-installer/utils/project"
	"github.com/urfave/cli"
)

func ProjectValidate(c *cli.Context) {
	err := project.ValidateProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}

func ProjectCreate(c *cli.Context) {
	err := project.DownloadTemplate(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}

func ProjectSync(c *cli.Context) {
	err := project.SyncProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}

func ProjectBind(c *cli.Context) {
	err := project.BindProject(c)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//		utils.PrettyPrintJSON(project.Result{Status: "OK"})
	}
	os.Exit(0)
}
