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

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	v1 "k8s.io/api/core/v1"
	apiextv1beta "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	labelTemplate = "%s/%s"
	// anyLabelValue is the value that will allow any value for a label when building the label query.
	anyLabelValue = ""
)

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

//nolint:gocritic // the arguments are needed
func DoAutoDiscover() (env configuration.TestParameters,
	testData configuration.TestConfiguration,
	pods,
	debugPods []v1.Pod, crds []*apiextv1beta.CustomResourceDefinition, namespaces []string) {
	env, err := configuration.LoadEnvironmentVariables()
	if err != nil {
		logrus.Fatalln("can't load environment variable")
	}
	testData, err = configuration.LoadConfiguration(env.ConfigurationPath)
	if err != nil {
		logrus.Fatalln("can't load configuration")
	}
	filenames := []string{}
	if env.Kubeconfig != "" {
		filenames = append(filenames, env.Kubeconfig)
	}
	if env.Home != "" {
		path := filepath.Join(env.Home, ".kube", "config")
		filenames = append(filenames, path)
	}
	oc := ocpclient.NewOcpClient(filenames...)
	namespaces = namespacesListToStringList(testData.TargetNameSpaces)
	pods = findPodsByLabel(oc.Coreclient, testData.TargetPodLabels, namespaces)

	debugLabel := configuration.Label{Prefix: debugLabelPrefix, Name: debugLabelName, Value: debugLabelValue}
	debugLabels := []configuration.Label{debugLabel}
	debugNS := []string{defaultNamespace}
	debugPods = findPodsByLabel(oc.Coreclient, debugLabels, debugNS)
	crds = FindTestCrdNames(testData.CrdFilters)
	return env, testData, pods, debugPods, crds, namespaces
}

func namespacesListToStringList(namespaceList []configuration.Namespace) (stringList []string) {
	for _, ns := range namespaceList {
		stringList = append(stringList, ns.Name)
	}
	return stringList
}
