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

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/eclipse/codewind-installer/pkg/security"
	"github.com/eclipse/codewind-installer/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cwctlName = "cwctl_test"
const cwctl = "./" + cwctlName
const testDir = "./testDir"

func TestCwctl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	err := exec.Command("go", "build", "-o", cwctlName).Run()
	require.Nil(t, err)

	testBasicUsage(t)
	testUseInsecureKeyring(t)
	testCreateProjectFromTemplate(t)

	os.Remove(cwctlName)
}

func testBasicUsage(t *testing.T) {
	t.Run("cwctl", func(t *testing.T) {
		out, err := exec.Command(cwctl).Output()
		assert.Nil(t, err)
		require.NotNil(t, out)
	})
}

func testUseInsecureKeyring(t *testing.T) {
	t.Run("cwctl --insecureKeyring seckeyring update", func(t *testing.T) {
		os.Remove(security.GetPathToInsecureKeyring())

		cmd := exec.Command(cwctl, "--insecureKeyring", "seckeyring", "update",
			"--conid=local",
			"--username=testuser",
			"--password=seCretphrase",
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		require.Equal(t, "{\"status\":\"OK\"}\n", string(out))

		file, readErr := ioutil.ReadFile(security.GetPathToInsecureKeyring())
		assert.Nil(t, readErr)
		require.NotNil(t, file)

		secrets := []security.KeyringSecret{}
		unmarshalErr := json.Unmarshal([]byte(file), &secrets)
		assert.Nil(t, unmarshalErr)
		require.Len(t, secrets, 1)

		secret := secrets[0]
		assert.Equal(t, "testuser", string(secret.Username))
		require.Equal(t, "seCretphrase", string(secret.Password))

		os.Remove(security.GetPathToInsecureKeyring())
	})
	t.Run("INSECURE_KEYRING=true cwctl seckeyring update", func(t *testing.T) {
		os.Remove(security.GetPathToInsecureKeyring())

		cmd := exec.Command(cwctl, "seckeyring", "update",
			"--conid=local",
			"--username=testuser",
			"--password=seCretphrase",
		)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "INSECURE_KEYRING=true")
		out, err := cmd.Output()
		assert.Nil(t, err)
		require.Equal(t, "{\"status\":\"OK\"}\n", string(out))

		file, readErr := ioutil.ReadFile(security.GetPathToInsecureKeyring())
		assert.Nil(t, readErr)
		require.NotNil(t, file)

		secrets := []security.KeyringSecret{}
		unmarshalErr := json.Unmarshal([]byte(file), &secrets)
		assert.Nil(t, unmarshalErr)
		require.Len(t, secrets, 1)

		secret := secrets[0]
		assert.Equal(t, "testuser", string(secret.Username))
		require.Equal(t, "seCretphrase", string(secret.Password))

		os.Remove(security.GetPathToInsecureKeyring())
	})
}

func testCreateProjectFromTemplate(t *testing.T) {
	t.Run("cwctl project create --url <insecureTemplateRepo> --path <testDir>", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "project", "create",
			"--url="+test.PublicGHRepoURL,
			"--path="+testDir,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Equal(t, "{\"status\":\"success\",\"projectPath\":\"./testDir\",\"result\":{\"language\":\"javascript\",\"projectType\":\"nodejs\"}}\n", string(out))
	})
	t.Run("cwctl project create --url <secureTemplateRepo> --path <testDir> --username <test.GHEUsername> --password <test.GHEPassword>", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--username="+test.GHEUsername,
			"--password="+test.GHEPassword,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Equal(t, "{\"status\":\"success\",\"projectPath\":\"./testDir\",\"result\":{\"language\":\"unknown\",\"projectType\":\"docker\"}}\n", string(out))
	})
	t.Run("cwctl project create --url <secureTemplateRepo> --path <testDir> --username <goodUsername> --password <badPassword>", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--username="+test.GHEUsername,
			"--password=badpassword",
		)
		out, err := cmd.Output()
		assert.NotNil(t, err)
		assert.Equal(t, "", string(out))
	})
}
