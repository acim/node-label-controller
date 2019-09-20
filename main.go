package main

import (
	"flag"
	"time"

	"github.com/acim/node-label-controller/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/sample-controller/pkg/signals"
)

func main() {
	klog.InitFlags(nil)

	var masterURL, kubeConfig, os, labelKey, labelValue string
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&os, "os", "Container Linux", "Operating system name's prefix, i.e. Ubuntu, Fedora...")
	flag.StringVar(&labelKey, "label-key", "kubermatic.io/uses-container-linux", "Label key to label a matching node")
	flag.StringVar(&labelValue, "label-value", "true", "Label value to label a matching node")
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, time.Second*30)

	c := controller.NewController(kubeClient, kubeInformerFactory.Core().V1().Nodes(), controller.OSPrefix(os), labelKey, labelValue)

	kubeInformerFactory.Start(stopCh)

	if err = c.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %v", err)
	}
}
