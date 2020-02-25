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

package remote

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateCodewindTektonClusterRoles : create Codewind tekton cluster roles
func CreateCodewindTektonClusterRoles(deployOptions *DeployOptions) rbacv1.ClusterRole {
	ourRoles := []rbacv1.PolicyRule{

		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list"},
		},
	}
	return rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: CodewindTektonClusterRolesName,
		},
		Rules: ourRoles,
	}
}

// CreateCodewindRoles : create Codewind roles
func CreateCodewindRoles(deployOptions *DeployOptions) rbacv1.ClusterRole {
	ourRoles := []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"extensions", ""},
			Resources: []string{"ingresses", "ingresses/status", "podsecuritypolicies"},
			Verbs:     []string{"delete", "create", "patch", "get", "list", "update", "watch", "use"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
			Verbs:     []string{"delete", "create", "patch", "get", "list"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"pods", "pods/portforward", "pods/log", "pods/exec"},
			Verbs:     []string{"get", "list", "create", "delete", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "list", "create", "watch", "delete", "patch", "update"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"serviceaccounts"},
			Verbs:     []string{"get", "patch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list", "create", "delete", "patch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"get", "list", "create", "update", "delete", "patch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"persistentvolumeclaims", "persistentvolumeclaims/finalizers", "persistentvolumeclaims/status"},
			Verbs:     []string{"*"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"icp.ibm.com"},
			Resources: []string{"images"},
			Verbs:     []string{"get", "list", "create", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"apps", "extensions"},
			Resources: []string{"deployments", "deployments/finalizers"},
			Verbs:     []string{"watch", "get", "list", "create", "update", "delete", "patch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"extensions", "apps"},
			Resources: []string{"replicasets", "replicasets/finalizers"},
			Verbs:     []string{"get", "list", "update", "delete"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"rolebindings", "roles", "clusterroles"},
			Verbs:     []string{"create", "get", "patch", "list"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"events"},
			Verbs:     []string{"create", "patch", "update"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes", "routes/custom-host"},
			Verbs:     []string{"get", "list", "create", "delete", "watch", "patch", "update"},
		},
	}
	return rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deployOptions.Namespace,
			Name:      CodewindRolesName,
		},
		Rules: ourRoles,
	}
}

//CreateCodewindRoleBindings : create Codewind role bindings in the deployment namespace
func CreateCodewindRoleBindings(codewindInstance Codewind, deployOptions *DeployOptions, codewindRoleBindingName string) rbacv1.RoleBinding {
	labels := map[string]string{
		"codewindWorkspace": codewindInstance.WorkspaceID,
	}
	return rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   codewindRoleBindingName,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      codewindInstance.ServiceAccountName,
				Namespace: deployOptions.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     CodewindRolesName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

//CreateCodewindTektonClusterRoleBindings : create Codewind tekton cluster role bindings
func CreateCodewindTektonClusterRoleBindings(codewindInstance Codewind, deployOptions *DeployOptions, roleBindingName string) rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app":               CodewindTektonClusterRoleBindingName,
		"codewindWorkspace": codewindInstance.WorkspaceID,
	}
	return rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   roleBindingName,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      codewindInstance.ServiceAccountName,
				Namespace: deployOptions.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     CodewindTektonClusterRolesName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
