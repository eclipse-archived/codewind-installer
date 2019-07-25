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
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/urfave/cli"
)

//CloneTemplate from github
func CloneTemplate(c *cli.Context) {
	//TODO Use go-git to do this in future
	url := c.String("template")
	if url == "" {
		log.Fatal("No repository URL provided")
	}
	cmd := exec.Command("git", "clone", url)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
	fmt.Printf(output.String())
}
