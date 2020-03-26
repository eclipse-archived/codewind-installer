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

package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestPullImage(t *testing.T) {
	t.Run("does not error when docker ImagePull succeeds", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		err := PullImage(client, "dummyImage", true)
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ImagePull errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		err := PullImage(client, "dummyImage", true)
		wantErr := &DockerError{errOpImagePull, errImagePull, errImagePull.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetImageList(t *testing.T) {
	t.Run("gets the image list that is returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageList, err := GetImageList(client)
		assert.Nil(t, err)
		assert.Equal(t, imageList, mockImageSummaryWithCwImages)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetImageList(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainerList(t *testing.T) {
	t.Run("gets the container list that is returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		containerList, err := GetContainerList(client)
		assert.Nil(t, err)
		assert.Equal(t, containerList, mockContainerListWithCwContainers)
	})

	t.Run("returns error when docker ContainerList returns error", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetContainerList(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestCheckImageStatus(t *testing.T) {
	t.Run("returns true when correct images are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageStatus, err := CheckImageStatus(client)
		assert.Nil(t, err)
		assert.True(t, imageStatus)
	})

	t.Run("returns false when codewind images are not returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithoutCw{}

		imageStatus, err := CheckImageStatus(client)
		assert.Nil(t, err)
		assert.False(t, imageStatus)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := CheckImageStatus(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestCheckContainerStatus(t *testing.T) {
	t.Run("returns true when correct containers are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		containerStatus, err := CheckContainerStatus(client)
		assert.Nil(t, err)
		assert.True(t, containerStatus)
	})

	t.Run("returns false when correct codewind containers are not returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithoutCw{}

		containerStatus, err := CheckContainerStatus(client)
		assert.Nil(t, err)
		assert.False(t, containerStatus)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := CheckContainerStatus(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetImageTags(t *testing.T) {
	t.Run("returns the image tags set in the ImageList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageTags, err := GetImageTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetImageTags(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainerTags(t *testing.T) {
	t.Run("returns the container tags set in the ContainerList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageTags, err := GetContainerTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetContainerTags(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, err, wantErr)
	})
}

func TestGetPFEHostAndPort(t *testing.T) {
	t.Run("returns the PFE host and port set in the ContainerList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		host, port, err := GetPFEHostAndPort(client)
		assert.Nil(t, err)
		assert.Equal(t, "pfe", host)
		assert.Equal(t, "1000", port)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, _, err := GetPFEHostAndPort(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestValidateImageDigest(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		_, err := ValidateImageDigest(client, "test:0.0.9")
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := ValidateImageDigest(client, "test:0.0.9")
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetAutoRemovePolicy(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		autoremovePolicy, err := getContainerAutoRemovePolicy(client, "pfe")
		assert.Nil(t, err)
		assert.True(t, autoremovePolicy)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := getContainerAutoRemovePolicy(client, "pfe")
		wantErr := &DockerError{errOpContainerInspect, errContainerInspect, errContainerInspect.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestStopContainer(t *testing.T) {
	t.Run("no error returned when container is stopped", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		err := StopContainer(client, types.Container{
			Names: []string{"/codewind-pfe"},
			ID:    "pfe",
			Image: "eclipse/codewind-pfe:0.0.9",
			Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}})
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ContainerStop errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		err := StopContainer(client, types.Container{})
		containerInspectErr := &DockerError{errOpContainerInspect, errContainerInspect, errContainerInspect.Error()}
		wantErr := &DockerError{errOpStopContainer, containerInspectErr, containerInspectErr.Desc}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainersToRemove(t *testing.T) {
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
			containersToRemove := GetContainersToRemove(test.containerList)
			assert.Equal(t, len(test.expectedContainers), len(containersToRemove))
			for _, container := range containersToRemove {
				assert.Contains(t, test.expectedContainers, container.Names[0])
			}
		})
	}
}
