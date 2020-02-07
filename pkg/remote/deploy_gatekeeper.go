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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// DeployGatekeeper : Deploy gatekeepers
func DeployGatekeeper(config *restclient.Config, clientset *kubernetes.Clientset, codewindInstance Codewind, deployOptions *DeployOptions) error {

	logr.Infoln("Preparing Codewind Gatekeeper resources")

	gatekeeperSecrets := generateGatekeeperSecrets(codewindInstance, deployOptions)
	gatekeeperService := generateGatekeeperService(codewindInstance)
	gatekeeperDeploy := generateGatekeeperDeploy(codewindInstance, deployOptions)
	gatekeeperSessionSecret := generateGatekeeperSessionSecret(codewindInstance, deployOptions)

	serverKey, serverCert, _ := generateCertificate(GatekeeperPrefix+codewindInstance.Ingress, "Codewind Gatekeeper")
	gatekeeperTLSSecret := generateGatekeeperTLSSecret(codewindInstance, serverKey, serverCert)

	logr.Infoln("Deploying Codewind Gatekeeper Secrets")

	_, err := clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&gatekeeperSecrets)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Gatekeeper secrets: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Gatekeeper Session Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&gatekeeperSessionSecret)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind secrets: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Gatekeeper TLS Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&gatekeeperTLSSecret)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Gatekeeper TLS secrets: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Gatekeeper Deployment")
	_, err = clientset.AppsV1().Deployments(deployOptions.Namespace).Create(&gatekeeperDeploy)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Gatekeeper deployment: %v\n", err)
		return err
	}

	logr.Infoln("Deploying Codewind Gatekeeper Service")
	_, err = clientset.CoreV1().Services(deployOptions.Namespace).Create(&gatekeeperService)
	if err != nil {
		logr.Errorf("Error: Unable to create Codewind Gatekeeper service: %v\n", err)
		return err
	}

	// Expose Codewind over an ingress or route
	if codewindInstance.OnOpenShift {
		logr.Infof("Deploying Codewind Gatekeeper Route")
		// Deploy a route on OpenShift
		route := generateRouteGatekeeper(codewindInstance)
		routev1client, err := routev1.NewForConfig(config)
		if err != nil {
			logr.Printf("Error retrieving route client for OpenShift: %v\n", err)
			os.Exit(1)
		}
		_, err = routev1client.Routes(codewindInstance.Namespace).Create(&route)
		if err != nil {
			logr.Printf("Error: Unable to create route for Codewind: %v\n", err)
			os.Exit(1)
		}
	} else {
		logr.Infof("Deploying Codewind Gatekeeper Ingress")
		ingress := generateIngressGatekeeper(codewindInstance)
		_, err = clientset.ExtensionsV1beta1().Ingresses(codewindInstance.Namespace).Create(&ingress)
		if err != nil {
			logr.Printf("Error: Unable to create ingress for Codewind Gatekeeper: %v\n", err)
			os.Exit(1)
		}
	}
	return nil
}

func generateGatekeeperTLSSecret(codewind Codewind, pemPrivateKey string, pemPublicCert string) corev1.Secret {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	secrets := map[string]string{
		"tls.crt": pemPublicCert,
		"tls.key": pemPrivateKey,
	}
	name := "secret-codewind-tls"
	return generateSecrets(codewind, name, secrets, labels)
}

func generateGatekeeperSessionSecret(codewind Codewind, deployOptions *DeployOptions) corev1.Secret {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	secrets := map[string]string{
		"session_secret": deployOptions.CodewindSessionSecret,
	}
	name := "secret-codewind-session"
	return generateSecrets(codewind, name, secrets, labels)
}

func generateGatekeeperSecrets(codewind Codewind, deployOptions *DeployOptions) corev1.Secret {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	secrets := map[string]string{
		"client_secret": deployOptions.ClientSecret,
	}
	name := "secret-codewind-client"
	return generateSecrets(codewind, name, secrets, labels)
}

func generateGatekeeperDeploy(codewind Codewind, deployOptions *DeployOptions) appsv1.Deployment {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	envVars := setGatekeeperEnvVars(codewind, deployOptions)
	return generateDeployment(codewind, GatekeeperPrefix, codewind.GatekeeperImage, GatekeeperContainerPort, volumes, volumeMounts, envVars, labels, codewind.ServiceAccountName, false)
}

func generateGatekeeperService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, GatekeeperPrefix, GatekeeperContainerPort, labels)
}

// generateIngressGatekeeper returns a Kubernetes ingress for the Codewind Gatekeeper service
func generateIngressGatekeeper(codewind Codewind) extensionsv1.Ingress {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/rewrite-target":     "/",
		"ingress.bluemix.net/redirect-to-https":          "True",
		"ingress.bluemix.net/ssl-services":               "ssl-service=" + GatekeeperPrefix + "-" + codewind.WorkspaceID,
		"nginx.ingress.kubernetes.io/backend-protocol":   "HTTPS",
		"kubernetes.io/ingress.class":                    "nginx",
		"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
	}

	return extensionsv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        GatekeeperPrefix + "-" + codewind.WorkspaceID,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: extensionsv1.IngressSpec{
			TLS: []extensionsv1.IngressTLS{
				{
					Hosts:      []string{GatekeeperPrefix + codewind.Ingress},
					SecretName: "secret-codewind-tls" + "-" + codewind.WorkspaceID,
				},
			},
			Rules: []extensionsv1.IngressRule{
				{
					Host: GatekeeperPrefix + codewind.Ingress,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: GatekeeperPrefix + "-" + codewind.WorkspaceID,
										ServicePort: intstr.FromInt(GatekeeperContainerPort),
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

// generateRouteGatekeeper returns an OpenShift route for the gatekeeper service
func generateRouteGatekeeper(codewind Codewind) v1.Route {
	labels := map[string]string{
		"app":               GatekeeperPrefix,
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
			Name:   GatekeeperPrefix + "-" + codewind.WorkspaceID,
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
			Host: GatekeeperPrefix + codewind.Ingress,
			Port: &v1.RoutePort{
				TargetPort: intstr.FromInt(GatekeeperContainerPort),
			},
			TLS: &v1.TLSConfig{
				InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   v1.TLSTerminationPassthrough,
			},
			To: v1.RouteTargetReference{
				Kind:   "Service",
				Name:   GatekeeperPrefix + "-" + codewind.WorkspaceID,
				Weight: &weight,
			},
		},
	}
}

func setGatekeeperEnvVars(codewind Codewind, deployOptions *DeployOptions) []corev1.EnvVar {

	keycloakURL := KeycloakPrefix + codewind.Ingress

	if deployOptions.KeycloakTLSSecure {
		keycloakURL = "https://" + keycloakURL
	} else {
		keycloakURL = "http://" + keycloakURL
	}

	if deployOptions.KeycloakURL != "" {
		keycloakURL = deployOptions.KeycloakURL
	}

	return []corev1.EnvVar{
		{
			Name:  "AUTH_URL",
			Value: keycloakURL,
		},
		{
			Name:  "CLIENT_ID",
			Value: deployOptions.KeycloakClient,
		},
		{
			Name:  "ENABLE_AUTH",
			Value: "1",
		},
		{
			Name:  "GATEKEEPER_HOST",
			Value: GatekeeperPrefix + codewind.Ingress,
		},
		{
			Name:  "REALM",
			Value: deployOptions.KeycloakRealm,
		},
		{
			Name:  "WORKSPACE_SERVICE",
			Value: "CODEWIND_PFE_" + codewind.WorkspaceID,
		},
		{
			Name:  "WORKSPACE_ID",
			Value: codewind.WorkspaceID,
		},
		{
			Name:  "ACCESS_ROLE",
			Value: "codewind-" + codewind.WorkspaceID,
		},
		{
			Name: "CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-codewind-client" + "-" + codewind.WorkspaceID}, Key: "client_secret"}},
		},
		{
			Name: "SESSION_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-codewind-session" + "-" + codewind.WorkspaceID}, Key: "session_secret"}},
		},
		{
			Name:  "PORTAL_HTTPS",
			Value: "true",
		},
	}
}
