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
	"os"

	v1 "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	logr "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// DeployKeycloak : Deploy Keycloak instance
func DeployKeycloak(config *restclient.Config, clientset *kubernetes.Clientset, codewindInstance Codewind, deployOptions *DeployOptions, onOpenShift bool) error {
	// Deploy Keycloak
	keycloakSecrets := generateKeycloakSecrets(codewindInstance, deployOptions)
	keycloakService := generateKeycloakService(codewindInstance)
	keycloakDeploy := generateKeycloakDeploy(codewindInstance)
	serverKey, serverCert, _ := generateCertificate(KeycloakPrefix+codewindInstance.Ingress, "Codewind Keycloak")
	keycloakTLSSecret := generateKeycloakTLSSecret(codewindInstance, serverKey, serverCert)
	keycloakPVC := generateKeycloakPVC(codewindInstance, deployOptions, "")

	logr.Infoln("Creating Codewind Keycloak PVC")
	_, err := clientset.CoreV1().PersistentVolumeClaims(deployOptions.Namespace).Create(&keycloakPVC)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Keycloak PVC: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Keycloak Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&keycloakSecrets)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Keycloak secrets: %v\n", err)
		return err
	}
	_, err = clientset.CoreV1().Services(deployOptions.Namespace).Create(&keycloakService)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Keycloak service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(deployOptions.Namespace).Create(&keycloakDeploy)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Keycloak deployment: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Keycloak TLS Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&keycloakTLSSecret)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Keycloak TLS secrets: %v\n", err)
		return err
	}

	// Expose Codewind over an ingress or route
	if onOpenShift {
		// Deploy a route on OpenShift
		route := generateKeycloakRoute(codewindInstance)
		routev1client, err := routev1.NewForConfig(config)
		if err != nil {
			logr.Printf("Error retrieving route client for OpenShift: %v\n", err)
			os.Exit(1)
		}
		_, err = routev1client.Routes(deployOptions.Namespace).Create(&route)
		if err != nil {
			logr.Printf("Error: Unable to create route for Codewind: %v\n", err)
			os.Exit(1)
		}

	} else {
		logr.Infof("Deploying Codewind Keycloak Ingress")
		ingress := generateIngressKeycloak(codewindInstance)
		_, err = clientset.ExtensionsV1beta1().Ingresses(deployOptions.Namespace).Create(&ingress)
		if err != nil {
			logr.Printf("Error: Unable to create ingress for Codewind Keycloak: %v\n", err)
			os.Exit(1)
		}
	}
	return nil
}

func generateKeycloakTLSSecret(codewind Codewind, pemPrivateKey string, pemPublicCert string) corev1.Secret {
	secrets := map[string]string{
		"tls.crt": pemPublicCert,
		"tls.key": pemPrivateKey,
	}
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	name := "secret-keycloak-tls"
	return generateSecrets(codewind, name, secrets, labels)
}

func generateKeycloakSecrets(codewind Codewind, deployOptions *DeployOptions) corev1.Secret {
	secrets := map[string]string{
		"keycloak-admin-user":     deployOptions.KeycloakUser,
		"keycloak-admin-password": deployOptions.KeycloakPassword,
	}
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	name := "secret-keycloak-user"
	return generateSecrets(codewind, name, secrets, labels)
}

func generateKeycloakDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	volumes, volumeMounts := setKeycloakVolumes(codewind)
	envVars := setKeycloakEnvVars(codewind)
	return generateDeployment(codewind, KeycloakPrefix, codewind.KeycloakImage, KeycloakContainerPort, volumes, volumeMounts, envVars, labels, codewind.ServiceAccountKC, false)
}

func generateKeycloakService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, KeycloakPrefix, KeycloakContainerPort, labels)
}

// generateIngressKeycloak returns a Kubernetes ingress for the Codewind Keycloak service
func generateIngressKeycloak(codewind Codewind) extensionsv1.Ingress {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/rewrite-target":     "/",
		"nginx.ingress.kubernetes.io/backend-protocol":   "HTTP",
		"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
		"kubernetes.io/ingress.class":                    "nginx",
	}

	return extensionsv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        KeycloakPrefix + "-" + codewind.WorkspaceID,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: extensionsv1.IngressSpec{
			TLS: []extensionsv1.IngressTLS{
				{
					Hosts:      []string{KeycloakPrefix + codewind.Ingress},
					SecretName: "secret-keycloak-tls" + "-" + codewind.WorkspaceID,
				},
			},
			Rules: []extensionsv1.IngressRule{
				{
					Host: KeycloakPrefix + codewind.Ingress,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: KeycloakPrefix + "-" + codewind.WorkspaceID,
										ServicePort: intstr.FromInt(KeycloakContainerPort),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// generateKeycloakRoute returns an OpenShift route for the Keycloak service
func generateKeycloakRoute(codewind Codewind) v1.Route {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	weight := int32(100)
	// blockOwnerDeletion := true
	// controller := true

	return v1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   KeycloakPrefix + "-" + codewind.WorkspaceID,
			Labels: labels,
			// OwnerReferences: []metav1.OwnerReference{
			// 	{
			// 		APIVersion:         "apps/v1",
			// 		BlockOwnerDeletion: &blockOwnerDeletion,
			// 		Controller:         &controller,
			// 		Kind:               "ReplicaSet",
			// 		Name:               codewind.OwnerReferenceName,
			// 		UID:                codewind.OwnerReferenceUID,
			// 	},
			// },
		},
		Spec: v1.RouteSpec{
			Host: KeycloakPrefix + codewind.Ingress,
			Port: &v1.RoutePort{
				TargetPort: intstr.FromInt(KeycloakContainerPort),
			},
			TLS: &v1.TLSConfig{
				InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   v1.TLSTerminationEdge,
			},
			To: v1.RouteTargetReference{
				Kind:   "Service",
				Name:   KeycloakPrefix + "-" + codewind.WorkspaceID,
				Weight: &weight,
			},
		},
	}
}

func setKeycloakEnvVars(codewind Codewind) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "KEYCLOAK_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-keycloak-user" + "-" + codewind.WorkspaceID}, Key: "keycloak-admin-user"}},
		},
		{
			Name: "KEYCLOAK_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-keycloak-user" + "-" + codewind.WorkspaceID}, Key: "keycloak-admin-password"}},
		},
		{
			Name:  "PROXY_ADDRESS_FORWARDING",
			Value: "true",
		},
		{
			Name:  "DB_VENDOR",
			Value: "h2",
		},
	}
}

func generateKeycloakPVC(codewind Codewind, deployOptions *DeployOptions, storageClass string) corev1.PersistentVolumeClaim {

	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	pvc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   KeycloakPrefix + "-pvc-" + codewind.WorkspaceID,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	// If a storage class was passed in, set it in the PVC
	if storageClass != "" {
		pvc.Spec.StorageClassName = &storageClass
	}

	return pvc
}

// setKeycloakVolumes returns a volumes & corresponding volume mount required by the Keycloak container:
func setKeycloakVolumes(codewind Codewind) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{
		{
			Name: "keycloak-data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: KeycloakPrefix + "-pvc-" + codewind.WorkspaceID,
				},
			},
		},
	}
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "keycloak-data",
			MountPath: "/opt/jboss/keycloak/standalone/data",
		},
	}
	return volumes, volumeMounts
}
