/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
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
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func clearAllDiagnostics() {
	app := cli.NewApp()
	flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
	flagSet.Bool("clean", true, "")
	context := cli.NewContext(app, flagSet, nil)
	DiagnosticsCommand(context)
}

func TestDiagnosticsCommand_NoArgs(t *testing.T) {
	clearAllDiagnostics()
	app := cli.NewApp()
	flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
	flagSet.String("conid", "local", "")
	context := cli.NewContext(app, flagSet, nil)
	t.Run("success case - no arguments specified", func(t *testing.T) {
		DiagnosticsCommand(context)
		assert.DirExists(t, filepath.Join(homeDir, ".codewind", "diagnostics"))
	})
}
