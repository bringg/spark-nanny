package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeClient struct {
	clientset *kubernetes.Clientset
}

// newKubeClient creates a new kuberentes client
func newKubeClient() *kubeClient {
	log.Info().Msg("creating kubernetes client")

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msg("creating kubernetes clientset")

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return &kubeClient{clientset: clientset}
}

// getDriverPod returns the driver pod for the given name and namespace
func (k *kubeClient) getDriverPod(name, ns string) (*corev1.Pod, error) {
	return k.clientset.CoreV1().Pods(ns).Get(context.TODO(), fmt.Sprintf("%s-driver", name), metav1.GetOptions{})
}

// deleteDriverPod deletes the driver pod for the given name and namespace
func (k *kubeClient) deleteDriverPod(name, ns string) error {
	return k.clientset.CoreV1().Pods(ns).Delete(context.TODO(), fmt.Sprintf("%s-driver", name), metav1.DeleteOptions{})
}
