package k8s

import (
	"fmt"

	"github.com/ysicing/default-backend/internal/kube"
	"k8s.io/client-go/informers"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	ingressv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var Client kubernetes.Interface
var IngressInformer ingressv1.IngressInformer

func init() {
	var err error
	kubecfg := &kube.ClientConfig{}
	Client, err = kube.New(kubecfg)
	if err != nil {
		panic(err)
	}
}

type Controlller struct {
	informerFactory    informers.SharedInformerFactory
	ingressInformer    ingressv1.IngressInformer
	deploymentInformer appsinformer.DeploymentInformer
}

func (c *Controlller) create(obj interface{}) {

}

func (c *Controlller) delete(obj interface{}) {

}

func (c *Controlller) update(oldobj, newobj interface{}) {

}

func (c *Controlller) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.ingressInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync caches")
	}
	return nil
}

func NewControlller(i informers.SharedInformerFactory) *Controlller {
	IngressInformer = i.Networking().V1().Ingresses()
	c := &Controlller{
		informerFactory:    i,
		ingressInformer:    IngressInformer,
		deploymentInformer: i.Apps().V1().Deployments(),
	}
	c.ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.create,
		DeleteFunc: c.delete,
		UpdateFunc: c.update,
	})
	return c
}
