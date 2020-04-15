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
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

// DockerClient requires all the functions called on the docker client
type DockerClient interface {
	ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error)
	ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error)
	ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)
	DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error)
}

// NewDockerClient creates a new client for the docker API
func NewDockerClient() (DockerClient, *DockerError) {
	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.30"))
	if err != nil {
		return nil, &DockerError{errOpClientCreate, err, err.Error()}
	}

	return dockerClient, nil
}
