package vodmodule_stats

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func podsClient(namespace string) (v1.PodInterface, error) {
	k8sCfg, err := rest.InClusterConfig()
	if err != nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		if k8sCfg, err = kubeConfig.ClientConfig(); err != nil {
			return nil, fmt.Errorf("failed in-cluster and file k8s cfgs: %w", err)
		}
	}

	cs, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		return nil, fmt.Errorf("creating client from cfg %#v: %w", k8sCfg, err)
	}

	return cs.CoreV1().Pods(namespace), nil
}
