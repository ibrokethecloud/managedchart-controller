//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"

	"github.com/harvester/managedchart-controller/pkg/crd"
	"github.com/harvester/managedchart-controller/pkg/server"
	"github.com/rancher/wrangler/v3/pkg/leader"
	"github.com/rancher/wrangler/v3/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	DefaultNamespace = "harvester-system"
)

func main() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := signals.SetupSignalContext()
	// setup CRDs
	if err := crd.Create(ctx, config); err != nil {
		logrus.Fatal(err)
	}

	s := server.NewServer(config)
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("error setting up k8sclient: %v", err)
	}
	manager := leader.NewManager(DefaultNamespace, "managedchart-controller", k8sClient)
	manager.OnLeader(func(ctx context.Context) error {
		return s.Start(ctx)
	})
	manager.Start(ctx)
	<-ctx.Done()
}
