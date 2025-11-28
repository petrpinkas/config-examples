package kubernetes

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	k8sClient client.Client
	once      sync.Once
)

// GetClient returns a singleton Kubernetes client
func GetClient() (client.Client, error) {
	var err error
	once.Do(func() {
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))

		cfg, cfgErr := config.GetConfig()
		if cfgErr != nil {
			err = cfgErr
			return
		}

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	})
	return k8sClient, err
}

