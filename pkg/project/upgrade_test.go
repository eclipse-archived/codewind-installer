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

package project

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UpgradeProjects(t *testing.T) {
	workspaceFolder := "../../resources/workspaces/"

	/*
		Currently unused, awaiting valid output testing
			validOuput := make(map[string]interface{})
			validOuput["migrated"] = []string{"valid-project"}
			validOuput["failed"] = make([]interface{}, 0)
	*/

	missingInfoOutput := make(map[string]interface{})
	missingInfoOutput["migrated"] = make([]string, 0)
	missingInfoOutput["failed"] = []interface{}{
		&map[string]string{
			"projectName": "missing-project-info",
			"error":       "Unable to upgrade project, failed to determine project details",
		},
	}

	missingDirectoryOutput := make(map[string]interface{})
	missingDirectoryOutput["migrated"] = make([]string, 0)
	missingDirectoryOutput["failed"] = []interface{}{
		&map[string]string{
			"projectName": "missing-project-dir",
			"error":       "stat ../../resources/workspaces/error-projects/missing-project-dir/missing-project-dir: no such file or directory",
		},
	}

	tests := map[string]struct {
		workspaceDir   string
		expectsErr     bool
		expectedOutput *map[string]interface{}
	}{
		"success case: missing directory should return bad path error": {
			workspaceDir:   "/does-not-exist/",
			expectsErr:     true,
			expectedOutput: nil,
		},
		"success case: missing .projects directory should return no projects error": {
			workspaceDir:   "/no-projects-dir/",
			expectsErr:     true,
			expectedOutput: nil,
		},
		"success case: empty .projects directory returns empty output and no errors": {
			workspaceDir:   "/empty/",
			expectsErr:     false,
			expectedOutput: &map[string]interface{}{"migrated": []string{}, "failed": []interface{}{}},
		},
		/*
			TODO: requires API call to bind project. Needs to be mocked.
				"success case: successfully upgrades a valid project": {
					workspaceDir:   "/valid-projects",
					expectsErr:     false,
					expectedOutput: &validOuput,
				},
		*/
		"success case: reports failed upgrade when project inf is missing information": {
			workspaceDir:   "/error-projects/missing-project-info",
			expectsErr:     false,
			expectedOutput: &missingInfoOutput,
		},
		"success case: reports failed upgrade when project directory is missing": {
			workspaceDir:   "/error-projects/missing-project-dir",
			expectsErr:     false,
			expectedOutput: &missingDirectoryOutput,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			path := path.Join(workspaceFolder, test.workspaceDir)
			response, err := UpgradeProjects(path)

			assert.Exactly(t, test.expectedOutput, response, "upgrade gave incorrect response")
			if test.expectsErr {
				assert.Error(t, err, "upgrade did not return an error when one was expected")
			} else {
				assert.Equal(t, (*ProjectError)(nil), err, "upgrade returned error %+v when none were expected", err)
			}
		})
	}

}
