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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	logr "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetImages returns the images that are to be used for PFE and the Performance dashboard in Codewind
// If environment vars are set (such as $PFE_IMAGE, $PFE_TAG, $PERFORMANCE_IMAGE, or $PERFORMANCE_TAG), it will use those,
// otherwise it defaults to the constants defined in constants/default.go
func GetImages() (string, string, string, string) {
	var pfeImage, performanceImage, keycloakImage, gatekeeperImage string
	var pfeTag, performanceTag, keycloakTag, gatekeeperTag string

	if pfeImage = os.Getenv("PFE_IMAGE"); pfeImage == "" {
		pfeImage = PFEImage
	}
	if performanceImage = os.Getenv("PERFORMANCE_IMAGE"); performanceImage == "" {
		performanceImage = PerformanceImage
	}
	if keycloakImage = os.Getenv("KEYCLOAK_IMAGE"); keycloakImage == "" {
		keycloakImage = KeycloakImage
	}
	if gatekeeperImage = os.Getenv("GATEKEEPER_IMAGE"); gatekeeperImage == "" {
		gatekeeperImage = GatekeeperImage
	}
	if pfeTag = os.Getenv("PFE_TAG"); pfeTag == "" {
		pfeTag = PFEImageTag
	}
	if performanceTag = os.Getenv("PERFORMANCE_TAG"); performanceTag == "" {
		performanceTag = PerformanceTag
	}
	if keycloakTag = os.Getenv("KEYCLOAK_TAG"); keycloakTag == "" {
		keycloakTag = KeycloakImageTag
	}
	if gatekeeperTag = os.Getenv("GATEKEEPER_TAG"); gatekeeperTag == "" {
		gatekeeperTag = GatekeeperImageTag
	}
	return pfeImage + ":" + pfeTag, performanceImage + ":" + performanceTag, keycloakImage + ":" + keycloakTag, gatekeeperImage + ":" + gatekeeperTag
}

// Get kubeconfig
func getKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	// Use KUBECONFIG environment variable if set
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if ok && kubeconfig != "" {
		// If multiple files provided choose first.
		kubeconfig = filepath.SplitList(kubeconfig)[0]
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
			return nil, &RemInstError{errOpNotFound, err, err.Error()}
		}
		return config, nil
	}

	homeDir := getHomeDir()

	kubeconfig = filepath.Join(homeDir, ".kube", "config")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		inClusterConfig, inClusterConfigErr := rest.InClusterConfig()
		if inClusterConfigErr != nil {
			logr.Infof("Unable to retrieve Kubernetes Config %v\n", err)
			return nil, &RemInstError{errOpNotFound, err, err.Error()}
		}
		return inClusterConfig, nil
	}

	return config, nil
}

// Get home directory
func getHomeDir() string {
	homeDir := ""
	const GOOS string = runtime.GOOS
	if GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}

// generateDeployment returns a Kubernetes deployment object with the given name for the given image.
// Additionally, volume/volumemounts and env vars can be specified.
func generateDeployment(codewind Codewind, name string, image string, port int, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, envVars []corev1.EnvVar, labels map[string]string, serviceAccountName string, privileged bool) appsv1.Deployment {

	//blockOwnerDeletion := true
	//controller := true
	replicas := int32(1)
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + codewind.WorkspaceID,
			Namespace: codewind.Namespace,
			Labels:    labels,
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
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Volumes:            volumes,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: ImagePullPolicy,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: volumeMounts,
							Env:          envVars,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(port),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func generateSecrets(codewind Codewind, name string, secrets map[string]string, labels map[string]string) corev1.Secret {
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + codewind.WorkspaceID,
			Namespace: codewind.Namespace,
			Labels:    labels,
		},
		StringData: secrets,
	}
	return secret
}

// generateService returns a Kubernetes service object with the given name, exposed over the specified port
// for the container with the given labels.
func generateService(codewind Codewind, name string, port int, labels map[string]string) corev1.Service {
	//blockOwnerDeletion := true
	//controller := true
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + codewind.WorkspaceID,
			Namespace: codewind.Namespace,
			Labels:    labels,
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
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(port),
					Name: name + "-http",
				},
			},
			Selector: labels,
		},
	}
	return service
}

func generateCertificate(dnsName string, certTitle string) (string, string, error) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{certTitle},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{dnsName},
	}

	logr.Println("Creating " + dnsName + " server Key")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logr.Errorln("Unable to create server key")
		return "", "", err
	}

	logr.Println("Creating " + dnsName + " server certificate")
	certDerBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		logr.Errorf("Failed to create certificate: %s\n", err)
		return "", "", err
	}

	out := &bytes.Buffer{}
	pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})
	pemPublicCert := out.String()
	out.Reset()
	pem.Encode(out, pemBlockForKey(privateKey))
	pemPrivateKey := out.String()
	out.Reset()

	return pemPrivateKey, pemPublicCert, nil
}

func publicKey(privateKey interface{}) interface{} {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(privateKey interface{}) *pem.Block {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	default:
		return nil
	}
}

// WaitForPodReady : Wait for pod to enter the running phase
func WaitForPodReady(clientset *kubernetes.Clientset, codewindInstance Codewind, labelSelector string, podName string) bool {

	logr.Infof("Waiting for pod: %v", podName)
	var waitTime int64 = 30

	watcher, err := clientset.CoreV1().Pods(codewindInstance.Namespace).Watch(metav1.ListOptions{
		LabelSelector:  labelSelector,
		TimeoutSeconds: &waitTime,
	})
	if err != nil {
		logr.Errorln("Unable to attach watcher")
		os.Exit(1)
	}
	lastPhase := ""
	lastReason := ""
	changeEvent := watcher.ResultChan()
	for {
		event, channelOk := <-changeEvent
		if !channelOk {
			// Channel closed, no event value. (Watch probably timed out.)
			logr.Warnf("Watch for Pod %v ended. Timeout or connection error.", podName)
			return false
		}
		foundPod, castOk := event.Object.(*corev1.Pod)
		if castOk == false {
			logr.Errorf("Expected a Pod but found: %T\n", event.Object)
			os.Exit(1)
		}
		status := foundPod.Status
		for _, condition := range foundPod.Status.Conditions {
			currentReason := string(condition.Reason)
			if lastPhase != string(status.Phase) || (lastReason != currentReason && currentReason != "") {
				logr.Printf("%v, phase: %v %v \n", podName, status.Phase, currentReason)
			}
			lastPhase = string(status.Phase)
			lastReason = currentReason
			if status.Phase == corev1.PodRunning {
				watcher.Stop()
				return true
			}
		}
	}
}
