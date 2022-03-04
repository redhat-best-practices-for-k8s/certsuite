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
	"fmt"
	"path/filepath"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	clientsholder "github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	v1apps "k8s.io/api/apps/v1"
	v1scaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	tnfCsvTargetLabelName  = "operator"
	tnfCsvTargetLabelValue = ""
	tnfLabelPrefix         = "test-network-function.com"
	labelTemplate          = "%s/%s"
	// anyLabelValue is the value that will allow any value for a label when building the label query.
	anyLabelValue = ""
)

type DiscoveredTestData struct {
	Env         configuration.TestParameters
	TestData    configuration.TestConfiguration
	Pods        []v1.Pod
	DebugPods   []v1.Pod
	Crds        []*apiextv1.CustomResourceDefinition
	Namespaces  []string
	Csvs        []olmv1Alpha.ClusterServiceVersion
	Deployments []v1apps.Deployment
	StatefulSet []v1apps.StatefulSet
	Hpas        map[string]*v1scaling.HorizontalPodAutoscaler
}

var data = DiscoveredTestData{}

func buildLabelName(labelPrefix, labelName string) string {
	if labelPrefix == "" {
		return labelName
	}
	return fmt.Sprintf(labelTemplate, labelPrefix, labelName)
}

func buildLabelQuery(label configuration.Label) string {
	fullLabelName := buildLabelName(label.Prefix, label.Name)
	if label.Value != anyLabelValue {
		return fmt.Sprintf("%s=%s", fullLabelName, label.Value)
	}
	return fullLabelName
}
func buildLabelKeyValue(label configuration.Label) (key, value string) {
	key = buildLabelName(label.Prefix, label.Name)
	value = label.Value
	return key, value
}

//nolint:funlen
// DoAutoDiscover finds objects under test
func DoAutoDiscover() DiscoveredTestData {
	var err error
	data.Env, err = configuration.LoadEnvironmentVariables()
	if err != nil {
		logrus.Fatalln("can't load environment variable")
	}
	data.TestData, err = configuration.LoadConfiguration(data.Env.ConfigurationPath)
	if err != nil {
		logrus.Fatalln("can't load configuration")
	}
	filenames := []string{}
	if data.Env.Kubeconfig != "" {
		filenames = append(filenames, data.Env.Kubeconfig)
	}
	if data.Env.Home != "" {
		path := filepath.Join(data.Env.Home, ".kube", "config")
		filenames = append(filenames, path)
	}
	oc := clientsholder.NewClientsHolder(filenames...)
	data.Namespaces = namespacesListToStringList(data.TestData.TargetNameSpaces)
	data.Pods = findPodsByLabel(oc.Coreclient, data.TestData.TargetPodLabels, data.Namespaces)

	debugLabel := configuration.Label{Prefix: debugLabelPrefix, Name: debugLabelName, Value: debugLabelValue}
	debugLabels := []configuration.Label{debugLabel}
	debugNS := []string{defaultNamespace}
	data.DebugPods = findPodsByLabel(oc.Coreclient, debugLabels, debugNS)
	data.Crds = FindTestCrdNames(data.TestData.CrdFilters)
	data.Csvs = findOperatorsByLabel(oc.OlmClient, []configuration.Label{{Name: tnfCsvTargetLabelName, Prefix: tnfLabelPrefix, Value: tnfCsvTargetLabelValue}}, data.TestData.TargetNameSpaces)
	data.Deployments = findDeploymentByLabel(oc.AppsClients, data.TestData.TargetPodLabels, data.Namespaces)
	data.StatefulSet = findStatefulSetByLabel(oc.AppsClients, data.TestData.TargetPodLabels, data.Namespaces)
	data.Hpas = findHpaControllers(oc.K8sClient, data.Namespaces)
	return data
}

func namespacesListToStringList(namespaceList []configuration.Namespace) (stringList []string) {
	for _, ns := range namespaceList {
		stringList = append(stringList, ns.Name)
	}
	return stringList
}
