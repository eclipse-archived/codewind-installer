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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetProjectPodFromID(t *testing.T) {
	k8s := newTestSimpleK8s()
	testOptions := &ProjectPod{
		ProjectName: "testpod",
		Namespace:   "test-ns",
		ProjectID:   "a-test-project-id",
	}
	k8s.clientset = fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testOptions.ProjectName,
			Namespace: testOptions.Namespace,
			Labels: map[string]string{
				"projectID": testOptions.ProjectID,
			},
		},
	})
	want := &ProjectPod{
		ProjectName: testOptions.ProjectName,
		Namespace:   testOptions.Namespace,
		ProjectID:   testOptions.ProjectID,
	}
	t.Run("Success case - pod exists with given projectID", func(t *testing.T) {
		got, err := k8s.GetProjectPodFromID(testOptions.ProjectID)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("Error case - no pod exists with given projectID", func(t *testing.T) {
		_, err := k8s.GetProjectPodFromID("notvalidID")
		assert.NotNil(t, err)
	})
}
