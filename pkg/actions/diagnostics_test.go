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
	"io/ioutil"
	"os"
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

// clearAllDiagnostics()
// app := cli.NewApp()
// flagSet := flag.NewFlagSet("userFlags", flag.ContinueOnError)
// flagSet.String("conid", "local", "")
// context := cli.NewContext(app, flagSet, nil)
// t.Run("local success case - no arguments specified", func(t *testing.T) {
// 	originalStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w
// 	DiagnosticsCommand(context)
// 	w.Close()
// 	out, _ := ioutil.ReadAll(r)
// 	os.Stdout = originalStdout
// 	fmt.Println("Spitting out output")
// 	fmt.Println(string(out))
// 	assert.DirExists(t, filepath.Join(homeDir, ".codewind", "diagnostics"))
//  })

func TestWarnDG(t *testing.T) {
	warning := "test_warn"
	description := "test warning description"
	expectedConsoleOutput := warning + ": " + description + "\n"
	expectedJSONOutput := dgWarning{WarningType: warning, WarningDesc: description}
	t.Run("warnDG - console", func(t *testing.T) {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		warnDG(warning, description)
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = originalStdout
		assert.Equal(t, expectedConsoleOutput, string(out))
	})
	t.Run("warnDG - json", func(t *testing.T) {
		dgWarningArray = []dgWarning{}
		printAsJSON = true
		warnDG(warning, description)
		assert.Equal(t, expectedJSONOutput, dgWarningArray[0])
	})
}
