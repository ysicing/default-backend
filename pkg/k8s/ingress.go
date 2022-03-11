package k8s

import (
	"context"

	"github.com/ergoapi/util/exmap"
	"github.com/ergoapi/util/ptr"
	"github.com/ergoapi/zlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Meta struct {
	Host      string
	Name      string
	Namespace string
	Appid     string
	Status    string
}

func GetIngress(host string) *Meta {
	zlog.Debug("check host: %s", host)
	var m Meta
	ings, _ := IngressInformer.Lister().List(labels.Everything())
	for _, ing := range ings {
		if ing.Name == host {
			m.Host = host
			m.Appid = exmap.GetLabelValue(ing.Labels, "k8s.easycorp.work/appid")
			m.Name = exmap.GetLabelValue(ing.Labels, "k8s.easycorp.work/name")
			m.Namespace = ing.Namespace
			if len(m.Appid) == 0 || len(m.Name) == 0 {
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

func StartDeploy(host string) string {
	m := GetIngress(host)
	if m.Status == "404" {
		return m.Status
	}
	zlog.Debug("check deploy: %s", host)
	dg, err := Client.AppsV1().Deployments(m.Namespace).Get(context.TODO(), m.Name, metav1.GetOptions{})
	if err != nil {
		return "404"
	}
	if *dg.Spec.Replicas > 0 {
		return "503"
	}
	dg.Spec.Replicas = ptr.Int32Ptr(1)
	_, err = Client.AppsV1().Deployments(m.Namespace).Update(context.TODO(), dg, metav1.UpdateOptions{})
	if err != nil {
		zlog.Error("update deployment: %v", err)
		return "404"
	}
	return "200"
}
