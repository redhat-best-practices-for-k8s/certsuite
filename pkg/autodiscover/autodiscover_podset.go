// Copyright (C) 2020-2021 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
package autodiscover

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	v1 "k8s.io/api/apps/v1"
	v1scaling "k8s.io/api/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func FindDeploymentByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*v1.Deployment, error) {
	dpClient := appClient.Deployments(namespace)
	options := metav1.GetOptions{}
	dp, err := dpClient.Get(context.TODO(), name, options)
	if err != nil {
		logrus.Error("Can't retrieve deployment in ns=", namespace, " name=", name)
		return nil, err
	}
	return dp, nil
}
func FindStatefulsetByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*v1.StatefulSet, error) {
	ssClient := appClient.StatefulSets(namespace)
	ss, err := ssClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logrus.Error("Can't retrieve deployment in ns=", namespace, " name=", name)
		return nil, err
	}
	return ss, nil
}

//nolint:dupl
func findDeploymentByLabel(
	appClient appv1client.AppsV1Interface,
	labels []configuration.Label,
	namespaces []string,
) []v1.Deployment {
	deployments := []v1.Deployment{}
	for _, ns := range namespaces {
		dps, err := appClient.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorln("Failed to list deployments in ns=", ns, ", trying to proceed")
			continue
		}
		if len(dps.Items) == 0 {
			logrus.Warn("Did not find any deployments in ns=", ns)
		}

		for i := 0; i < len(dps.Items); i++ {
			for _, l := range labels {
				key, value := buildLabelKeyValue(l)
				logrus.Tracef("Searching pods in deployment %q found in ns %q using label %s=%s", dps.Items[i].Name, ns, key, value)
				if dps.Items[i].Spec.Template.ObjectMeta.Labels[key] == value {
					deployments = append(deployments, dps.Items[i])
					logrus.Info("Deployment ", dps.Items[i].Name, " found in ns ", ns)
				}
			}
		}
	}
	if len(deployments) == 0 {
		logrus.Warnf("Did not find any deployment in the configured namespaces %v", namespaces)
	}
	return deployments
}

//nolint:dupl
func findStatefulSetByLabel(
	appClient appv1client.AppsV1Interface,
	labels []configuration.Label,
	namespaces []string,
) []v1.StatefulSet {
	statefulsets := []v1.StatefulSet{}
	for _, ns := range namespaces {
		ss, err := appClient.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorln("Failed to list statefulsets in ns=", ns, ", trying to proceed")
			continue
		}
		if len(ss.Items) == 0 {
			logrus.Warn("Did not find any statefulSet in ns=", ns)
		}

		for i := 0; i < len(ss.Items); i++ {
			for _, l := range labels {
				key, value := buildLabelKeyValue(l)
				logrus.Tracef("Searching pods in statefulset %q found in ns %q using label %s=%s", ss.Items[i].Name, ns, key, value)
				if ss.Items[i].Spec.Template.ObjectMeta.Labels[key] == value {
					statefulsets = append(statefulsets, ss.Items[i])
					logrus.Info("StatefulSet ", ss.Items[i].Name, " found in ns ", ns)
				}
			}
		}
	}
	if len(statefulsets) == 0 {
		logrus.Warnf("Did not find any statefulset in the configured namespaces %v", namespaces)
	}
	return statefulsets
}

func findHpaControllers(cs kubernetes.Interface, namespaces []string) map[string]*v1scaling.HorizontalPodAutoscaler {
	m := make(map[string]*v1scaling.HorizontalPodAutoscaler)
	for _, ns := range namespaces {
		hpas, err := cs.AutoscalingV1().HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Error("can't list HorizontalPodAutoscalers on namespace ", ns, " err ", err)
			return m
		}
		for i := 0; i < len(hpas.Items); i++ {
			name := ns + hpas.Items[i].Name
			m[name] = &hpas.Items[i]
		}
	}
	if len(m) == 0 {
		logrus.Info("can't find any deployed HorizontalPodAutoscaler")
	}
	return m
}
