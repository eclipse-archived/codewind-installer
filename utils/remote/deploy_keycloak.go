package remote

import (
	"os"

	v1 "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	logr "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// DeployKeycloak : Deploy Keycloak instance
func DeployKeycloak(config *restclient.Config, clientset *kubernetes.Clientset, codewindInstance Codewind, deployOptions *DeployOptions, onOpenShift bool) error {
	// Deploy Keycloak
	keycloakSecrets := createKeycloakSecrets(codewindInstance, deployOptions)
	keycloakService := createKeycloakService(codewindInstance)
	keycloakDeploy := createKeycloakDeploy(codewindInstance)
	serverKey, serverCert, err := createCertificate(KeycloakPrefix+codewindInstance.Ingress, "Codewind Keycloak")
	keycloakTLSSecret := createKeycloakTLSSecret(codewindInstance, serverKey, serverCert)

	log.Infoln("Deploying Codewind Keycloak Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&keycloakSecrets)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Keycloak secrets: %v\n", err)
		return err
	}
	_, err = clientset.CoreV1().Services(deployOptions.Namespace).Create(&keycloakService)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Keycloak service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(deployOptions.Namespace).Create(&keycloakDeploy)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Keycloak deployment: %v\n", err)
		return err
	}

	log.Infoln("Deploying Codewind Keycloak TLS Secrets")
	_, err = clientset.CoreV1().Secrets(deployOptions.Namespace).Create(&keycloakTLSSecret)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Keycloak TLS secrets: %v\n", err)
		return err
	}

	// Expose Codewind over an ingress or route
	if onOpenShift {
		// Deploy a route on OpenShift
		route := createKeycloakRoute(codewindInstance)
		routev1client, err := routev1.NewForConfig(config)
		if err != nil {
			log.Printf("Error retrieving route client for OpenShift: %v\n", err)
			os.Exit(1)
		}
		_, err = routev1client.Routes(deployOptions.Namespace).Create(&route)
		if err != nil {
			log.Printf("Error: Unable to create route for Codewind: %v\n", err)
			os.Exit(1)
		}

	} else {
		logr.Infof("Deploying Codewind Keycloak Ingress")
		ingress := createIngressKeycloak(codewindInstance)
		_, err = clientset.ExtensionsV1beta1().Ingresses(deployOptions.Namespace).Create(&ingress)
		if err != nil {
			log.Printf("Error: Unable to create ingress for Codewind Keycloak: %v\n", err)
			os.Exit(1)
		}
	}
	return nil
}

func createKeycloakTLSSecret(codewind Codewind, pemPrivateKey string, pemPublicCert string) corev1.Secret {
	secrets := map[string]string{
		"tls.crt": pemPublicCert,
		"tls.key": pemPrivateKey,
	}
	name := "secret-keycloak-tls"
	return generateSecrets(codewind, name, secrets)
}

func createKeycloakSecrets(codewind Codewind, deployOptions *DeployOptions) corev1.Secret {
	secrets := map[string]string{
		"keycloak-admin-user":     deployOptions.KeycloakUser,
		"keycloak-admin-password": deployOptions.KeycloakPassword,
	}
	name := "secret-keycloak-user"
	return generateSecrets(codewind, name, secrets)
}

func createKeycloakDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	envVars := setKeycloakEnvVars(codewind)
	return generateDeployment(codewind, KeycloakPrefix, codewind.KeycloakImage, KeycloakContainerPort, volumes, volumeMounts, envVars, labels)
}

func createKeycloakService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               KeycloakPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, KeycloakPrefix, KeycloakContainerPort, labels)
}

// CreateIngressKeycloak returns a Kubernetes ingress for the Codewind Keycloak service
func createIngressKeycloak(codewind Codewind) extensionsv1.Ingress {
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

	// blockOwnerDeletion := true
	// controller := true

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

// createKeycloakRoute returns an OpenShift route for the Keycloak service
func createKeycloakRoute(codewind Codewind) v1.Route {
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
	}
}
