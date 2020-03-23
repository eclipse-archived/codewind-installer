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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ProjectPod : Relevant properities of a remote deploted project pod
type ProjectPod struct {
	Namespace   string
	ProjectID   string
	ProjectName string
}

// PortForwardPodRequest : The request made to forward the port from a remote pod to local
type PortForwardPodRequest struct {
	RestConfig *rest.Config
	Pod        v1.Pod
	LocalPort  int
	PodPort    int
	Streams    genericclioptions.IOStreams
	StopCh     <-chan struct{}
	ReadyCh    chan struct{}
}

// HandlePortForward : Forwards port from remote pod to local
func HandlePortForward(projectID string, namespace string) error {
	// wg will wait for goroutines to finish
	var wg sync.WaitGroup
	wg.Add(1)

	config, err := GetKubeConfig()
	if err != nil {
		return err
	}

	client := K8sAPI{}
	client.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	podInfo, err := client.GetProjectPodFromID(projectID)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	// stream controls where portforward sends its output,
	// and where to expect its input from
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		close(stopCh)
		fmt.Println(sig)
		wg.Done()
	}()

	go func() error {
		err := PortForwardPod(PortForwardPodRequest{
			RestConfig: config,
			Pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podInfo.ProjectName,
					Namespace: podInfo.Namespace,
				},
			},
			LocalPort: 9229,
			PodPort:   9229,
			Streams:   stream,
			StopCh:    stopCh,
			ReadyCh:   readyCh,
		})
		if err != nil {
			return err
		}
		return nil
	}()

	select {
	case <-readyCh:
		break
	case <-stopCh:
		println("stopped!")
		break
	}
	wg.Wait()
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
func (client K8sAPI) GetProjectPodFromID(projectID string) (*ProjectPod, error) {
	// projectIDs are unique, so this should only return one deployment
	podList, err := client.clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "projectID=" + projectID,
	})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, err
	}

	pod := podList.Items[0]
	return &ProjectPod{
		ProjectName: pod.GetName(),
		Namespace:   pod.GetNamespace(),
		ProjectID:   projectID,
	}, nil
}
