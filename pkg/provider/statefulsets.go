// Copyright (C) 2022 Red Hat, Inc.
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

package provider

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	appsv1 "k8s.io/api/apps/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type StatefulSet struct {
	*appsv1.StatefulSet
}

func (ss *StatefulSet) IsStatefulSetReady() bool {
	var replicas int32
	if ss.Spec.Replicas != nil {
		replicas = *(ss.Spec.Replicas)
	} else {
		replicas = 1
	}
	if ss.Status.ReadyReplicas != replicas ||
		ss.Status.CurrentReplicas != replicas ||
		ss.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}

func (ss *StatefulSet) ToString() string {
	return fmt.Sprintf("statefulset: %s ns: %s",
		ss.Name,
		ss.Namespace,
	)
}

func (ss *StatefulSet) AffinityRequired() bool {
	if val, ok := ss.Labels[AffinityRequiredKey]; ok {
		result, err := strconv.ParseBool(val)
		if err != nil {
			logrus.Warnf("failure to parse bool %v", val)
			return false
		}
		return result
	}
	return false
}

func (ss *StatefulSet) IsAffinityCompliant() (bool, error) {
	if ss.Spec.Template.Spec.Affinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding affinity rules", ss.String())
	}
	if ss.Spec.Template.Spec.Affinity.PodAntiAffinity != nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but has anti-affinity rules", ss.String())
	}
	if ss.Spec.Template.Spec.Affinity.PodAffinity == nil && ss.Spec.Template.Spec.Affinity.NodeAffinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding pod/node affinity rules", ss.String())
	}
	return true, nil
}

func GetUpdatedStatefulset(ac appv1client.AppsV1Interface, namespace, podName string) (*StatefulSet, error) {
	result, err := autodiscover.FindStatefulsetByNameByNamespace(ac, namespace, podName)
	return &StatefulSet{
		result,
	}, err
}
