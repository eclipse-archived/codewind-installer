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

package remote

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type (
	// ProjectPod : Relevant properities of a remote deployed project pod
	ProjectPod struct {
		Namespace   string
		ProjectID   string
		ProjectName string
	}

	// PortForwardPodRequest : The request made to forward the port from a remote pod to local
	PortForwardPodRequest struct {
		RestConfig *rest.Config
		Pod        v1.Pod
		LocalPort  int
		PodPort    int
		Streams    IOStreams
		StopCh     <-chan struct{}
		ReadyCh    chan struct{}
	}

	// PortForwarder : The channels required to create a new k8s port-forwarder
	PortForwarder struct {
		StopChannel  chan struct{}
		ReadyChannel chan struct{}
	}
)

// IOStreams provides the standard names for IOstreams, which are passed to kubernetes port-forward
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// HandlePortForward : Forwards port from remote pod to local
func HandlePortForward(projectID string, port int) *RemInstError {
	kubeConfig, err := GetKubeConfig()
	if err != nil {
		logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
		return &RemInstError{errOpNotFound, err, err.Error()}
	}

	client := K8sAPI{}
	client.clientset, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return &RemInstError{errOpNotFound, err, err.Error()}
	}

	podInfo, RemErr := client.GetProjectPodFromID(projectID)
	if RemErr != nil {
		return RemErr
	}

	portForwarder := PortForwarder{
		StopChannel:  make(chan struct{}, 1),
		ReadyChannel: make(chan struct{}),
	}

	stream := IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// If system interrupt, close the Stop Channel and hence finish port-forward
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		close(portForwarder.StopChannel)
	}()

	err = PortForwardPod(PortForwardPodRequest{
		RestConfig: kubeConfig,
		Pod: v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podInfo.ProjectName,
				Namespace: podInfo.Namespace,
			},
		},
		LocalPort: 9229,
		PodPort:   port,
		Streams:   stream,
		StopCh:    portForwarder.StopChannel,
		ReadyCh:   portForwarder.ReadyChannel,
	})
	if err != nil {
		return &RemInstError{errOpPortForward, err, err.Error()}
	}
	return nil
}

// PortForwardPod : Requests PortForward from remote pod to local
func PortForwardPod(req PortForwardPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// GetProjectPodFromID : Gets a project pod from its id
func (client K8sAPI) GetProjectPodFromID(projectID string) (*ProjectPod, *RemInstError) {
	// projectIDs are unique, so this should only return one deployment
	podList, err := client.clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "projectID=" + projectID,
	})
	if err != nil {
		return nil, &RemInstError{errOpGetPods, err, err.Error()}
	}

	if len(podList.Items) == 0 {
		err = errors.New("No remote pod with given projectID")
		return nil, &RemInstError{errOpGetPods, err, err.Error()}
	}

	if len(podList.Items) > 1 {
		err = errors.New("Multiple remote pods with given projectID")
		return nil, &RemInstError{errOpGetPods, err, err.Error()}
	}

	pod := podList.Items[0]
	return &ProjectPod{
		ProjectName: pod.GetName(),
		Namespace:   pod.GetNamespace(),
		ProjectID:   projectID,
	}, nil
}
