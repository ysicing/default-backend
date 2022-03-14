package k8s

import (
	"context"

	"github.com/ergoapi/util/exmap"
	"github.com/ergoapi/util/ptr"
	"github.com/ergoapi/zlog"
	"github.com/ysicing/default-backend/internal/cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	NotExistCode = 0
	ExistRunning = 1
	ExistStart   = 2
	Exist        = 99
)

type Meta struct {
	Host      string
	Name      string
	Namespace string
	Appid     string
	Status    int
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
				m.Status = NotExistCode
			} else {
				m.Status = Exist
			}
			return &m
		}
	}
	m.Status = NotExistCode
	return &m
}

func StartDeploy(host string) int {
	m := GetIngress(host)
	if m.Status == NotExistCode {
		return m.Status
	}
	dg, err := Client.AppsV1().Deployments(m.Namespace).Get(context.TODO(), m.Name, metav1.GetOptions{})
	if err != nil {
		return NotExistCode
	}
	replicas := *dg.Spec.Replicas
	if replicas > 0 {
		if cache.Check(host) {
			return ExistStart
		}
		return ExistRunning
	}
	dg.Spec.Replicas = ptr.Int32Ptr(1)
	_, err = Client.AppsV1().Deployments(m.Namespace).Update(context.TODO(), dg, metav1.UpdateOptions{})
	if err != nil {
		zlog.Error("update deployment: %v", err)
	}
	cache.Set(host)
	return ExistStart
}
