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
	defer os.Remove(cwctlName)

	testBasicUsage(t)
	testUseInsecureKeyring(t)
	testCreateProjectFromTemplate(t)
	testSuccessfulAddAndRemoveTemplateRepos(t)
	testFailToAddTemplateRepo(t)
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
	t.Run("success case: create project"+
		"\ncwctl project create --url <insecureTemplateRepo> --path <testDir>", func(t *testing.T) {
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
	t.Run("success case: create project and output JSON"+
		"\ncwctl --json project create --url <insecureTemplateRepo> --path <testDir>", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "--json", "project", "create",
			"--url="+test.PublicGHRepoURL,
			"--path="+testDir,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Equal(t, "{\"status\":\"success\",\"projectPath\":\"./testDir\",\"result\":{\"language\":\"javascript\",\"projectType\":\"nodejs\"}}\n", string(out))
	})
	t.Run("success case: create GHE project with good username and password"+
		"\ncwctl project create --url <secureTemplateRepo> --path <testDir> --username <test.GHEUsername> --password <test.GHEPassword>", func(t *testing.T) {
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
	t.Run("success case: create GHE project with good personalAccessToken"+
		"\ncwctl project create --url <secureTemplateRepo> --path <testDir> --personalAccessToken <test.GHEPersonalAccessToken>", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--personalAccessToken="+test.GHEPersonalAccessToken,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Equal(t, "{\"status\":\"success\",\"projectPath\":\"./testDir\",\"result\":{\"language\":\"unknown\",\"projectType\":\"docker\"}}\n", string(out))
	})
	t.Run("fail case: create GHE project with good username but bad password"+
		"\ncwctl project create --url <secureTemplateRepo> --path <testDir> --username <goodUsername> --password <badPassword>", func(t *testing.T) {
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--username="+test.GHEUsername,
			"--password=badpassword",
		)
		out, err := cmd.CombinedOutput()
		assert.NotNil(t, err)
		assert.Contains(t, string(out), "401 Unauthorized")
	})
}

func testSuccessfulAddAndRemoveTemplateRepos(t *testing.T) {
	t.Run("success case: add and remove template repo"+
		"\ncwctl templates repos add --url <publicGHDevfile>"+
		"\ncwctl templates repos remove --url <publicGHDevfile>", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.PublicGHDevfileURL,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.PublicGHDevfileURL)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.PublicGHDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.PublicGHDevfileURL)
	})
	t.Run("success case: add template repo with name and description"+
		"\ncwctl templates repos add --url --name --description"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.PublicGHDevfileURL,
			"--name=publicGHDevfile",
			"--description=publicGHDevfile",
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.PublicGHDevfileURL)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.PublicGHDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.PublicGHDevfileURL)
	})
	t.Run("success case: add GHE template repo using username and password"+
		"\ncwctl templates repos add --url <GHEDevfile> --username --password"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--username="+test.GHEUsername,
			"--password="+test.GHEPassword,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
	t.Run("success case: add GHE template repo using personal access token"+
		"\ncwctl templates repos add --url <GHEDevfile> --personalAccessToken"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--personalAccessToken="+test.GHEPersonalAccessToken,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
	t.Run("success case: create GHE template project using stored GHE username-password"+
		"\ncwctl templates repos add --url <GHEDevfile> --username <goodUsername> --password <goodPassword>"+
		"\ncwctl project create --url <GHETemplateRepo>"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--username="+test.GHEUsername,
			"--password="+test.GHEPassword,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		createCmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
		)
		createOut, createErr := createCmd.Output()
		assert.Nil(t, createErr)
		assert.Contains(t, string(createOut), "success")
		assert.Contains(t, string(createOut), testDir)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
	t.Run("success case: create GHE template project using stored GHE personalAccessToken"+
		"\ncwctl templates repos add --url <GHEDevfile> --personalAccessToken <goodToken>"+
		"\ncwctl project create --url <GHETemplateRepo>"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--personalAccessToken="+test.GHEPersonalAccessToken,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		createCmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
		)
		createOut, createErr := createCmd.Output()
		assert.Nil(t, createErr)
		assert.Contains(t, string(createOut), "success")
		assert.Contains(t, string(createOut), testDir)

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
	t.Run("fail case: create GHE template project with bad password, overriding good stored GHE creds"+
		"\ncwctl templates repos add --url <GHEDevfile> --username --password"+
		"\ncwctl project create --url <GHETemplateRepo> --username <goodUsername> --password <badPassword>"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--username="+test.GHEUsername,
			"--password="+test.GHEPassword,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		createCmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--username="+test.GHEUsername,
			"--password=badpassword",
		)
		createOut, createErr := createCmd.CombinedOutput()
		assert.NotNil(t, createErr)
		assert.Contains(t, string(createOut), "401 Unauthorized")

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
	t.Run("fail case: create GHE template project with bad personalAccessToken, overriding good stored GHE creds"+
		"\ncwctl templates repos add --url <GHEDevfile> --personalAccessToken <goodToken>"+
		"\ncwctl project create --url <GHETemplateRepo> --personalAccessToken <badToken>"+
		"\ncwctl templates repos remove --url", func(t *testing.T) {
		if !test.UsingOwnGHECredentials {
			t.Skip("skipping this test because you haven't set GitHub credentials needed for this test")
		}

		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--personalAccessToken="+test.GHEPersonalAccessToken,
		)
		out, err := cmd.Output()
		assert.Nil(t, err)
		assert.Contains(t, string(out), test.GHEDevfileURL)

		createCmd := exec.Command(cwctl, "project", "create",
			"--url="+test.GHERepoURL,
			"--path="+testDir,
			"--personalAccessToken=badtoken",
		)
		createOut, createErr := createCmd.CombinedOutput()
		assert.NotNil(t, createErr)
		assert.Contains(t, string(createOut), "401 Unauthorized")

		removeCmd := exec.Command(cwctl, "templates", "repos", "remove",
			"--url="+test.GHEDevfileURL,
		)
		removeOut, removeErr := removeCmd.Output()
		assert.Nil(t, removeErr)
		assert.NotContains(t, string(removeOut), test.GHEDevfileURL)
	})
}

func testFailToAddTemplateRepo(t *testing.T) {
	t.Run("cwctl templates repos add --url <badURL>", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url=https://raw.githubusercontent.com/kabanero-io/codewind-templates/aad4bafc14e1a295fb8e462c20fe8627248609a3/devfiles/NOT_INDEX_JSON",
		)
		out, err := cmd.CombinedOutput()
		assert.Nil(t, err)
		assert.Contains(t, string(out), "does not point to a JSON file of the correct form")
	})
	t.Run("cwctl templates repos add --url <GHEDevfile> --personalAccessToken --username", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--personalAccessToken=validPersonalAccessToken",
			"--username=validUsername",
		)
		out, err := cmd.CombinedOutput()
		assert.Nil(t, err)
		assert.Contains(t, string(out), "received credentials for multiple authentication methods")
	})
	t.Run("cwctl templates repos add --url <GHEDevfile> --username", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--username=validUsername",
		)
		out, err := cmd.CombinedOutput()
		assert.Nil(t, err)
		assert.Contains(t, string(out), "received username but no password")
	})
	t.Run("cwctl templates repos add --url <GHEDevfile> --password", func(t *testing.T) {
		cmd := exec.Command(cwctl, "templates", "repos", "add",
			"--url="+test.GHEDevfileURL,
			"--password=validPassword",
		)
		out, err := cmd.CombinedOutput()
		assert.Nil(t, err)
		assert.Contains(t, string(out), "received password but no username")
	})
}
