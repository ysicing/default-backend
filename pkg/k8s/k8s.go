package k8s

import (
	"context"
	"fmt"

	"github.com/ergoapi/util/exmap"
	"github.com/ergoapi/zlog"
	"github.com/ysicing/default-backend/internal/kube"
	"github.com/ergoapi/util/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	ingressv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var Client kubernetes.Interface

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

type Meta struct {
	Host      string
	Name      string
	Namespace string
	Appid     string
	Status    string
}

func (c *Controlller) GetIngress(host string) *Meta {
	var m Meta
	ings, _ := c.ingressInformer.Lister().List(labels.Everything())
	for _, ing := range ings {
		if ing.Name == host {
			m.Host = host
			m.Appid = exmap.GetLabelValue(ing.Labels, "k8s.easycorp.work/appid")
			m.Name = exmap.GetLabelValue(ing.Labels, "k8s.easycorp.work/name")
			m.Namespace = ing.Namespace
			if len(m.Appid) == 0 {
				m.Status = "404"
			} else {
				m.Status = "200"
			}
			return &m
		}
	}
	m.Status = "404"
	return &m
}

func (c *Controlller) StartDeploy(host string) string {
	m := c.GetIngress(host)
	if m.Status == "404" {
		return m.Status
	}
	d, err := c.deploymentInformer.Lister().Deployments(m.Namespace).Get(m.Name)
	if err != nil {
		zlog.Error("get deployment: %v", err)
		return "404"
	}
	if *d.Spec.Replicas > 0 {
		return "200"
	}
	dg, _ := Client.AppsV1().Deployments(m.Namespace).Get(context.TODO(), m.Name, metav1.GetOptions{})
	dg.Spec.Replicas = ptr.Int32Ptr(1)
	_, err = Client.AppsV1().Deployments(m.Namespace).Update(context.TODO(), dg, metav1.UpdateOptions{})
	if err != nil {
		zlog.Error("update deployment: %v", err)
		return "404"
	}
	return "200"
}

func (c *Controlller) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.ingressInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync caches")
	}
	return nil
}

func NewControlller(i informers.SharedInformerFactory) *Controlller {
	c := &Controlller{
		informerFactory:    i,
		ingressInformer:    i.Networking().V1().Ingresses(),
		deploymentInformer: i.Apps().V1().Deployments(),
	}
	c.ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.create,
		DeleteFunc: c.delete,
		UpdateFunc: c.update,
	})
	return c
}
