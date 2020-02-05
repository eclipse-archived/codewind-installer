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

package utils

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

var mockImageSummary = []types.ImageSummary{
	types.ImageSummary{
		ID:          "pfe",
		RepoDigests: []string{"eclipse/codewind-pfe", "sha256:7173b809", "test:0.0.9"},
		RepoTags:    []string{"test:0.0.9"},
	},
	types.ImageSummary{
		ID:          "performance",
		RepoDigests: []string{"eclipse/codewind-performance", "sha256:7173b809", "test:0.0.9"},
		RepoTags:    []string{"test:0.0.9"},
	},
}

var mockContainerList = []types.Container{
	types.Container{
		Names: []string{"/codewind-pfe"},
		ID:    "pfe",
		Image: "eclipse/codewind-pfe:0.0.9",
		Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}},
	types.Container{
		Names: []string{"/codewind-performance"},
		Image: "eclipse/codewind-performance:0.0.9"},
}

type mockDockerClient struct {
}

func (m *mockDockerClient) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return mockImageSummary, nil
}

func (m *mockDockerClient) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClient) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return mockContainerList, nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return nil
}

func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return nil
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				AutoRemove: true,
			},
		},
	}, nil
}

func (m *mockDockerClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{
		Descriptor: v1.Descriptor{
			Digest: "sha256:7173b809",
		},
	}, nil
}

func TestPullImage(t *testing.T) {
	t.Run("does not error", func(t *testing.T) {
		client := &mockDockerClient{}
		err := PullImage(client, "dummyImage", true)
		assert.Nil(t, err)
	})
}

func TestGetImageList(t *testing.T) {
	t.Run("gets an image list", func(t *testing.T) {
		client := &mockDockerClient{}

		imageList, err := GetImageList(client)
		assert.Nil(t, err)
		assert.Equal(t, imageList, mockImageSummary)
	})
}

func TestGetContainerList(t *testing.T) {
	t.Run("gets a container list", func(t *testing.T) {
		client := &mockDockerClient{}

		containerList, err := GetContainerList(client)
		assert.Nil(t, err)
		assert.Equal(t, containerList, mockContainerList)
	})
}

func TestGetContainersToRemove(t *testing.T) {
	t.Run("gets correct containers to remove", func(t *testing.T) {
		containerList := []types.Container{
			types.Container{Names: []string{"/cw-test"}},
			types.Container{Names: []string{"/test"}},
		}
		containersToRemove := GetContainersToRemove(containerList)
		wantContainersToRemove := []types.Container{
			types.Container{Names: []string{"/cw-test"}},
		}
		assert.Equal(t, wantContainersToRemove, containersToRemove)

	})
}

func TestCheckImageStatus(t *testing.T) {
	t.Run("returns true when correct images are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClient{}

		imageStatus, err := CheckImageStatus(client)
		assert.Nil(t, err)
		assert.True(t, imageStatus)
	})
}

func TestCheckContainerStatus(t *testing.T) {
	t.Run("returns true when correct containers are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClient{}

		containerStatus, err := CheckContainerStatus(client)
		assert.Nil(t, err)
		assert.True(t, containerStatus)
	})
}

func TestGetImageTags(t *testing.T) {
	t.Run("returns the image tags set in the imageList mock", func(t *testing.T) {
		client := &mockDockerClient{}

		imageTags, err := GetImageTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})
}

func TestGetContainerTags(t *testing.T) {
	t.Run("returns the container tags set in the containerList mock", func(t *testing.T) {
		client := &mockDockerClient{}

		imageTags, err := GetContainerTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})
}

func TestGetPFEHostAndPort(t *testing.T) {
	t.Run("returns the PFE host and port set in the containerList mock", func(t *testing.T) {
		client := &mockDockerClient{}

		host, port, err := GetPFEHostAndPort(client)
		assert.Nil(t, err)
		assert.Equal(t, "pfe", host)
		assert.Equal(t, "1000", port)
	})
}

func TestValidateImageDigest(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClient{}
		_, err := ValidateImageDigest(client, "test:0.0.9")
		assert.Nil(t, err)
	})
}

func TestGetAutoRemovePolicy(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClient{}
		autoremovePolicy, err := getContainerAutoRemovePolicy(client, "pfe")
		assert.Nil(t, err)
		assert.True(t, autoremovePolicy)
	})
}

func TestStopContainer(t *testing.T) {
	t.Run("no error returned when container is stopped", func(t *testing.T) {
		client := &mockDockerClient{}
		err := StopContainer(client, types.Container{
			Names: []string{"/codewind-pfe"},
			ID:    "pfe",
			Image: "eclipse/codewind-pfe:0.0.9",
			Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}})
		assert.Nil(t, err)
	})
}
