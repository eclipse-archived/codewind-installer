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

package docker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var mockImageSummaryWithCwImages = []types.ImageSummary{
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

var mockContainerListWithCwContainers = []types.Container{
	types.Container{
		Names: []string{"/codewind-pfe"},
		ID:    "pfe",
		Image: "eclipse/codewind-pfe:0.0.9",
		Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}},
	types.Container{
		Names: []string{"/codewind-performance"},
		Image: "eclipse/codewind-performance:0.0.9"},
	types.Container{
		Names: []string{"/cw-testProject"},
		Image: "eclipse/codewind-performance:0.0.9"},
}

var mockContainerListWithOnlyPFEContainer = []types.Container{
	types.Container{
		Names: []string{"/codewind-pfe"},
		ID:    "pfe",
		Image: "eclipse/codewind-pfe:0.0.9",
		Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}},
}

var mockImageSummaryWithoutCwImages = []types.ImageSummary{
	types.ImageSummary{
		ID:       "golang",
		RepoTags: []string{"golang:0.0.9"},
	},
	types.ImageSummary{
		ID:       "registry",
		RepoTags: []string{"registry:0.0.9"},
	},
}

var mockContainerListWithoutCwContainers = []types.Container{
	types.Container{
		Names: []string{"/registry"},
		Image: "registry",
	},
	types.Container{
		Names: []string{"/go-test"},
		Image: "golang",
	},
}

//MockDockerClientWithCw - This mock client will return container and images lists, with Codewind items included
type MockDockerClientWithCw struct {
}

//ImagePull - returns empty ReadCloser
func (m *MockDockerClientWithCw) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

//ImageList - reutrns some mock images
func (m *MockDockerClientWithCw) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return mockImageSummaryWithCwImages, nil
}

//ContainerList - returns some mock containers
func (m *MockDockerClientWithCw) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return mockContainerListWithCwContainers, nil
}

//ContainerStop - returns no errors
func (m *MockDockerClientWithCw) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return nil
}

//ContainerRemove - returns no errors
func (m *MockDockerClientWithCw) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return nil
}

//ClientVersion - returns empty string
func (m *MockDockerClientWithCw) ClientVersion() string {
	return ""
}

//ContainerLogs - returns empty ReadCloser
func (m *MockDockerClientWithCw) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

//CopyFromContainer - returns empty ReadCloser, empty ContainerPathStat
func (m *MockDockerClientWithCw) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, types.ContainerPathStat{Name: "", Size: 0, Mode: 0, Mtime: time.Now(), LinkTarget: ""}, nil
}

//ServerVersion - returns empty Version struct
func (m *MockDockerClientWithCw) ServerVersion(ctx context.Context) (types.Version, error) {
	return types.Version{Platform: struct{ Name string }{""}, Components: []types.ComponentVersion{}, Version: "", APIVersion: "", MinAPIVersion: "", GitCommit: "", GoVersion: "", Os: "", Arch: "", KernelVersion: "", Experimental: true, BuildTime: ""}, nil
}

//ContainerInspect - returns basic ContainerJSON
func (m *MockDockerClientWithCw) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				AutoRemove: true,
			},
		},
	}, nil
}

//DaemonHost - returns empty string
func (m *MockDockerClientWithCw) DaemonHost() string {
	return ""
}

//DistributionInspect - returns a basic DistributionInspect
func (m *MockDockerClientWithCw) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{
		Descriptor: v1.Descriptor{
			Digest: "sha256:7173b809",
		},
	}, nil
}

//RegistryLogin - returns basic AuthenticateOKBody
func (m *MockDockerClientWithCw) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, nil
}

// This mock client will return container and images lists, with only a PFE container running
type mockDockerClientWithPFEContainerOnly struct {
}

func (m *mockDockerClientWithPFEContainerOnly) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClientWithPFEContainerOnly) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return mockImageSummaryWithCwImages, nil
}

func (m *mockDockerClientWithPFEContainerOnly) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return mockContainerListWithOnlyPFEContainer, nil
}

func (m *mockDockerClientWithPFEContainerOnly) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return nil
}

func (m *mockDockerClientWithPFEContainerOnly) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return nil
}

func (m *mockDockerClientWithPFEContainerOnly) ClientVersion() string {
	return ""
}

func (m *mockDockerClientWithPFEContainerOnly) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClientWithPFEContainerOnly) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, types.ContainerPathStat{Name: "", Size: 0, Mode: 0, Mtime: time.Now(), LinkTarget: ""}, nil
}

func (m *mockDockerClientWithPFEContainerOnly) ServerVersion(ctx context.Context) (types.Version, error) {
	return types.Version{Platform: struct{ Name string }{""}, Components: []types.ComponentVersion{}, Version: "", APIVersion: "", MinAPIVersion: "", GitCommit: "", GoVersion: "", Os: "", Arch: "", KernelVersion: "", Experimental: true, BuildTime: ""}, nil
}

func (m *mockDockerClientWithPFEContainerOnly) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				AutoRemove: true,
			},
		},
	}, nil
}

func (m *mockDockerClientWithPFEContainerOnly) DaemonHost() string {
	return ""
}

func (m *mockDockerClientWithPFEContainerOnly) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{
		Descriptor: v1.Descriptor{
			Digest: "sha256:7173b809",
		},
	}, nil
}

func (m *mockDockerClientWithPFEContainerOnly) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, nil
}

// This mock client will return valid image and containers lists, without Codewind items
type mockDockerClientWithoutCw struct {
}

func (m *mockDockerClientWithoutCw) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClientWithoutCw) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return mockImageSummaryWithoutCwImages, nil
}

func (m *mockDockerClientWithoutCw) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return mockContainerListWithoutCwContainers, nil
}

func (m *mockDockerClientWithoutCw) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return nil
}

func (m *mockDockerClientWithoutCw) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return nil
}

func (m *mockDockerClientWithoutCw) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				AutoRemove: true,
			},
		},
	}, nil
}

func (m *mockDockerClientWithoutCw) DaemonHost() string {
	return ""
}

func (m *mockDockerClientWithoutCw) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{
		Descriptor: v1.Descriptor{
			Digest: "sha256:7173b809",
		},
	}, nil
}

func (m *mockDockerClientWithoutCw) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, nil
}

func (m *mockDockerClientWithoutCw) ClientVersion() string {
	return ""
}

func (m *mockDockerClientWithoutCw) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClientWithoutCw) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, types.ContainerPathStat{Name: "", Size: 0, Mode: 0, Mtime: time.Now(), LinkTarget: ""}, nil
}

func (m *mockDockerClientWithoutCw) ServerVersion(ctx context.Context) (types.Version, error) {
	return types.Version{Platform: struct{ Name string }{""}, Components: []types.ComponentVersion{}, Version: "", APIVersion: "", MinAPIVersion: "", GitCommit: "", GoVersion: "", Os: "", Arch: "", KernelVersion: "", Experimental: true, BuildTime: ""}, nil
}

//MockDockerErrorClient - This mock client will return errors for each call to a docker function
type MockDockerErrorClient struct {
}

var errImagePull = errors.New("error pulling image")
var errImageList = errors.New("error listing images")

//ErrContainerList - exported for testing purposes
var ErrContainerList = errors.New("error listing containers")
var errContainerStop = errors.New("error stopping container")
var errContainerRemove = errors.New("error removing container")

//ErrContainerInspect - exported for testing purposes
var ErrContainerInspect = errors.New("error inspecting container")
var errDistributionInspect = errors.New("error inspecting distribution")

//ErrContainerLogs - exported for testing purposes
var ErrContainerLogs = errors.New("error getting container logs")

//ErrCopyFromContainer - exported for testing purposes
var ErrCopyFromContainer = errors.New("error copying files from container")

//ErrServerVersion - exported for testing purposes
var ErrServerVersion = errors.New("error getting server version")

//ImageList - returns an error
func (m *MockDockerErrorClient) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return nil, errImageList
}

//ImagePull - returns an error
func (m *MockDockerErrorClient) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, errImagePull
}

//ContainerList - returns an error
func (m *MockDockerErrorClient) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return []types.Container{}, ErrContainerList
}

//ContainerStop - returns an error
func (m *MockDockerErrorClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return errContainerStop
}

//ContainerRemove - returns an error
func (m *MockDockerErrorClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return errContainerRemove
}

//ContainerInspect - returns an error
func (m *MockDockerErrorClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, ErrContainerInspect
}

//DaemonHost - returns an empty string
func (m *MockDockerErrorClient) DaemonHost() string {
	return ""
}

//DistributionInspect - returns an error
func (m *MockDockerErrorClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{}, errDistributionInspect
}

//RegistryLogin - returns an error
func (m *MockDockerErrorClient) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, nil
}

//ClientVersion - returns an empty string
func (m *MockDockerErrorClient) ClientVersion() string {
	return ""
}

//ContainerLogs - returns an error
func (m *MockDockerErrorClient) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, ErrContainerLogs
}

//CopyFromContainer - returns an error
func (m *MockDockerErrorClient) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, types.ContainerPathStat{Name: "", Size: 0, Mode: 0, Mtime: time.Now(), LinkTarget: ""}, ErrCopyFromContainer
}

//ServerVersion - returns an error
func (m *MockDockerErrorClient) ServerVersion(ctx context.Context) (types.Version, error) {
	return types.Version{Platform: struct{ Name string }{""}, Components: []types.ComponentVersion{}, Version: "", APIVersion: "", MinAPIVersion: "", GitCommit: "", GoVersion: "", Os: "", Arch: "", KernelVersion: "", Experimental: true, BuildTime: ""}, ErrServerVersion
}
