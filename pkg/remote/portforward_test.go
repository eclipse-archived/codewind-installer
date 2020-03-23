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
	"testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetProjectPodFromID(t *testing.T) {
	k8s := newTestSimpleK8s()
	k8s.clientset = fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Namespace: "test-ns",
			Labels: map[string]string{
				"projectID": "test",
			},
		},
	})
	want := &ProjectPod{
		ProjectName: "test",
		Namespace: "test-ns",
		ProjectID: "test",

	}
	got, err := k8s.GetProjectPodFromID("test")
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
