package main

import (
	"context"
	"flag"
	"github.com/google/uuid"
	"github.com/nickytd/leaderelection/leader"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"path/filepath"

	//need to run when oath token kubeconfig is supplied
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var kubeconfig, namespace, configmap, id string

func main() {
	klog.InitFlags(nil)
	flag.StringVar(
		&kubeconfig,
		"kubeconfig",
		defaultKubeconfig(),
		"path to kubeconfig")

	flag.StringVar(
		&namespace,
		"namespace",
		"default",
		"default namespace for creating the lease")

	flag.StringVar(
		&configmap,
		"configmap",
		"my-configmap",
		"default configmap name")

	flag.StringVar(
		&id,
		"id",
		uuid.New().String(),
		"leader identity")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	leader.SetupLeader(
		initClientSet(kubeconfig),
		ctx,
		cancel,
		namespace,
		configmap,
		id)
	defer cancel()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	klog.V(4).Infof("got signal %s", <-ch)
}

func initClientSet(kubeconfig string) *rest.Config {
	var config *rest.Config
	var err error

	klog.V(4).Infof("kubeconfig %s", kubeconfig)

	if config, err = rest.InClusterConfig(); err != nil && config == nil {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			klog.Errorf("error creating config %s", err.Error())
		}
	}
	return config
}

func defaultKubeconfig() string {
	fileName := os.Getenv("KUBECONFIG")
	if fileName != "" {
		return fileName
	}
	home, err := os.UserHomeDir()
	if err != nil {
		klog.Warningf("failed to get home directory: %s", err.Error())
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
