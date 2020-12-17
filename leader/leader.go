package leader

import (
	"context"
	"github.com/google/uuid"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"time"
)

func SetupLeader(config *rest.Config, ctx context.Context, cancel context.CancelFunc, namespace string, configmap string) {

	var clientset *kubernetes.Clientset
	var err error

	if clientset, err = kubernetes.NewForConfig(config); err != nil {
		klog.Errorf("error creating clientset %s", err.Error())
		return
	}

	uid := uuid.New().String()

	var configMapLock = &resourcelock.ConfigMapLock{
		Client: clientset.CoreV1(),
		ConfigMapMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      configmap,
		},
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: uid,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            configMapLock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.V(2).Infof("%s start leading", uid)
				createConfigMap(
					configMapLock.Client,
					namespace,
					configmap,
					uid,
				)
			},
			OnStoppedLeading: func() {
				klog.V(2).Infof("%s stop leading", uid)
				cancel()
			},
			OnNewLeader: func(identity string) {
				if identity == uid {
					return
				}
				klog.V(2).Infof("new leader %s", identity)
			},
		},
	})

}

func createConfigMap(clientset v12.ConfigMapsGetter, namespace string, name string, uid string) {

	_, err := clientset.ConfigMaps(namespace).Get(
		context.TODO(),
		name,
		v1.GetOptions{},
	)
	if err != nil {
		klog.V(2).Infof("create configmap %s", name)
		_, err = clientset.ConfigMaps(namespace).Create(
			context.TODO(),
			&v13.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string]string{"lock": uid},
			},
			v1.CreateOptions{},
		)
		if err != nil {
			klog.Errorf("error creating configmap %s", err.Error())
			return
		}
	} else {
		klog.V(2).Infof("update configmap %s", name)
		clientset.ConfigMaps(namespace).Update(
			context.TODO(),
			&v13.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string]string{"lock": uid},
			},
			v1.UpdateOptions{},
		)
	}
}
