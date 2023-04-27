// Copyright (C) 2020-2023 Red Hat, Inc.
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
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/scale"
)

func FindDeploymentByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.Deployment, error) {
	dp, err := appClient.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logrus.Error("Cannot retrieve deployment in ns=", namespace, " name=", name)
		return nil, err
	}
	return dp, nil
}
func FindStatefulsetByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.StatefulSet, error) {
	ss, err := appClient.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logrus.Error("Cannot retrieve deployment in ns=", namespace, " name=", name)
		return nil, err
	}
	return ss, nil
}

func FindCrObjectByNameByNamespace(scalesGetter scale.ScalesGetter, ns, name string, groupResourceSchema schema.GroupResource) (*scalingv1.Scale, error) {
	crScale, err := scalesGetter.Scales(ns).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
	if err != nil {
		logrus.Error("Cannot retrieve deployment in ns=", ns, " name=", name)
		return nil, err
	}
	return crScale, nil
}

//nolint:dupl
func findDeploymentByLabel(
	appClient appv1client.AppsV1Interface,
	labels []configuration.LabelObject,
	namespaces []string,
) []appsv1.Deployment {
	deployments := []appsv1.Deployment{}
	for _, ns := range namespaces {
		dps, err := appClient.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list deployments in ns=%s, err: %v . Trying to proceed.", ns, err)
			continue
		}
		if len(dps.Items) == 0 {
			logrus.Warn("Did not find any deployments in ns=", ns)
		}

		for i := 0; i < len(dps.Items); i++ {
			for _, aLabelObject := range labels {
				logrus.Tracef("Searching pods in deployment %q found in ns %q using label %s=%s", dps.Items[i].Name, ns, aLabelObject.LabelKey, aLabelObject.LabelValue)
				if dps.Items[i].Spec.Template.ObjectMeta.Labels[aLabelObject.LabelKey] == aLabelObject.LabelValue {
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
	labels []configuration.LabelObject,
	namespaces []string,
) []appsv1.StatefulSet {
	statefulsets := []appsv1.StatefulSet{}
	for _, ns := range namespaces {
		ss, err := appClient.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorf("Failed to list statefulsets in ns=%s, err: %v . Trying to proceed.", ns, err)
			continue
		}
		if len(ss.Items) == 0 {
			logrus.Warn("Did not find any statefulSet in ns=", ns)
		}

		for i := 0; i < len(ss.Items); i++ {
			for _, aLabelObject := range labels {
				logrus.Tracef("Searching pods in statefulset %q found in ns %q using label %s=%s", ss.Items[i].Name, ns, aLabelObject.LabelKey, aLabelObject.LabelValue)
				if ss.Items[i].Spec.Template.ObjectMeta.Labels[aLabelObject.LabelKey] == aLabelObject.LabelValue {
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

func findHpaControllers(cs kubernetes.Interface, namespaces []string) []*scalingv1.HorizontalPodAutoscaler {
	var m []*scalingv1.HorizontalPodAutoscaler
	for _, ns := range namespaces {
		hpas, err := cs.AutoscalingV1().HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Error("Cannot list HorizontalPodAutoscalers on namespace ", ns, " err ", err)
			return m
		}
		for i := 0; i < len(hpas.Items); i++ {
			m = append(m, &hpas.Items[i])
		}
	}
	if len(m) == 0 {
		logrus.Info("Cannot find any deployed HorizontalPodAutoscaler")
	}
	return m
}
