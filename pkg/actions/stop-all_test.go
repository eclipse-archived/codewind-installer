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
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func Test_getContainersToRemove(t *testing.T) {
	tests := map[string]struct {
		containerList      []types.Container
		expectedContainers []string
	}{
		"Returns project containers (cw-)": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-nodejsexpress"},
				},
				types.Container{
					Names: []string{"/cw-springboot"},
				},
			},
			expectedContainers: []string{
				"/cw-nodejsexpress",
				"/cw-springboot",
			},
		},
		"Ignores a non-codewind container": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-valid-container"},
				},
				types.Container{
					Names: []string{"invalid-container"},
				},
			},
			expectedContainers: []string{
				"/cw-valid-container",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			containersToRemove := getContainersToRemove(test.containerList)
			assert.Equal(t, len(test.expectedContainers), len(containersToRemove))
			for _, container := range containersToRemove {
				assert.Contains(t, test.expectedContainers, container.Names[0])
			}
		})
	}
}
