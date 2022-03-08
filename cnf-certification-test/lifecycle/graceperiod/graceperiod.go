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

package graceperiod

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1app "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultTerminationGracePeriod      = 30
	unconfiguredTerminationGracePeriod = -1
)

type lastAppliedConfigType struct {
	Spec struct {
		Template struct {
			Spec struct {
				TerminationGracePeriodSeconds int
			}
		}
	}
}

func getTerminationGracePeriodConfiguredInYaml(lastAppliedConfigString string) (terminationGracePeriodSeconds int, err error) {
	lastAppliedConfig := lastAppliedConfigType{}
	// Use -1 as default value, in case the param was not set.
	lastAppliedConfig.Spec.Template.Spec.TerminationGracePeriodSeconds = unconfiguredTerminationGracePeriod
	err = json.Unmarshal([]byte(lastAppliedConfigString), &lastAppliedConfig)
	if err != nil {
		return unconfiguredTerminationGracePeriod, err
	}
	return lastAppliedConfig.Spec.Template.Spec.TerminationGracePeriodSeconds, nil
}

func TestTerminationGracePeriodOnStatefulsets(env *provider.TestEnvironment) (badStatefulsets []*v1app.StatefulSet, curatedLogs loghelper.CuratedLogLines) { //nolint:dupl
	for _, sut := range env.SatetfulSets {
		aTerminationGracePeriodSeconds, err := getTerminationGracePeriodConfiguredInYaml(sut.Annotations[`kubectl.kubernetes.io/last-applied-configuration`])

		if err != nil {
			curatedLogs = curatedLogs.AddLogLine("Statefulset %s failed to get TerminationGracePeriodSeconds err: %s", provider.StatefulsetToString(sut), err)
			continue
		}

		if aTerminationGracePeriodSeconds == unconfiguredTerminationGracePeriod {
			curatedLogs = curatedLogs.AddLogLine("Statefulset %s does not have a terminationGracePeriodSeconds value set. Default value (%d) is used.",
				provider.StatefulsetToString(sut), defaultTerminationGracePeriod)
			badStatefulsets = append(badStatefulsets, sut)
		} else {
			logrus.Debugf("Statefulset %s last-applied-configuration's terminationGracePeriodSeconds: %d", provider.StatefulsetToString(sut), *sut.Spec.Template.Spec.TerminationGracePeriodSeconds)
		}
	}
	return badStatefulsets, curatedLogs
}

func TestTerminationGracePeriodOnDeployments(env *provider.TestEnvironment) (badDeployments []*v1app.Deployment, curatedLogs loghelper.CuratedLogLines) { //nolint:dupl
	for _, dut := range env.Deployments {
		aTerminationGracePeriodSeconds, err := getTerminationGracePeriodConfiguredInYaml(dut.Annotations[`kubectl.kubernetes.io/last-applied-configuration`])

		if err != nil {
			curatedLogs = curatedLogs.AddLogLine("Deployment %s failed to get TerminationGracePeriodSeconds err: %s", provider.DeploymentToString(dut), err)
			continue
		}

		if aTerminationGracePeriodSeconds == unconfiguredTerminationGracePeriod {
			curatedLogs = curatedLogs.AddLogLine("Deployment %s does not have a terminationGracePeriodSeconds value set. Default value (%d) is used.",
				provider.DeploymentToString(dut), defaultTerminationGracePeriod)
			badDeployments = append(badDeployments, dut)
		} else {
			logrus.Debugf("Deployment %s last-applied-configuration's terminationGracePeriodSeconds: %d", provider.DeploymentToString(dut), *dut.Spec.Template.Spec.TerminationGracePeriodSeconds)
		}
	}
	return badDeployments, curatedLogs
}

func TestTerminationGracePeriodOnPods(env *provider.TestEnvironment) (badPods []*v1.Pod, curatedLogs loghelper.CuratedLogLines) {
	numUnmanagedPods := 0
	for _, put := range env.Pods {
		// We'll process only "unmanaged" pods (not belonging to any deployment/statefulset) here.
		if len(put.OwnerReferences) != 0 {
			continue
		}
		numUnmanagedPods++
		aTerminationGracePeriodSeconds, err := getTerminationGracePeriodConfiguredInYaml(put.Annotations[`kubectl.kubernetes.io/last-applied-configuration`])

		if err != nil {
			curatedLogs = curatedLogs.AddLogLine("Pod %s failed to get TerminationGracePeriodSeconds err: %s", provider.PodToString(put), err)
			continue
		}

		if aTerminationGracePeriodSeconds == unconfiguredTerminationGracePeriod {
			curatedLogs = curatedLogs.AddLogLine("Pod %s does not have a terminationGracePeriodSeconds value set. Default value (%d) is used.",
				provider.PodToString(put), defaultTerminationGracePeriod)
			badPods = append(badPods, put)
		} else {
			logrus.Debugf("Pod %s last-applied-configuration's terminationGracePeriodSeconds: %d", provider.PodToString(put), *put.Spec.TerminationGracePeriodSeconds)
		}
	}
	logrus.Debugf("Number of unamanaged pods processed: %d", numUnmanagedPods)
	return badPods, curatedLogs
}
