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

package clientsholder

import (
	"fmt"
	"time"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	olmClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	olmFakeClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/sirupsen/logrus"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	k8sFakeClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientsHolder struct {
	RestConfig    *rest.Config
	DynamicClient dynamic.Interface
	APIExtClient  apiextv1.ApiextensionsV1Interface
	OlmClient     olmClient.Interface
	OcpClient     clientconfigv1.ConfigV1Interface
	K8sClient     kubernetes.Interface

	ready bool
}

var clientsHolder = ClientsHolder{}

// SetupFakeOlmClient Overrides the OLM client with the fake interface object for unit testing. Loads
// the mocking objects so olmv interface methods can find them.
func SetupFakeOlmClient(olmMockObjects []runtime.Object) {
	clientsHolder.OlmClient = olmFakeClient.NewSimpleClientset(olmMockObjects...)
}

// GetTestClientHolder Overwrites the existing clientholders with a mocked version for unit testing.
// Only pure k8s interfaces will be available. The runtime objects must be pure k8s ones.
// For other (OLM, )
// runtime mocking objects loading, use the proper clientset mocking function.
func GetTestClientsHolder(k8sMockObjects []runtime.Object, filenames ...string) *ClientsHolder {
	clientsHolder.K8sClient = k8sFakeClient.NewSimpleClientset(k8sMockObjects...)

	clientsHolder.ready = true
	return &clientsHolder
}

func ClearTestClientsHolder() {
	clientsHolder.K8sClient = nil
	clientsHolder.ready = false
}

// GetClientsHolder returns the singleton ClientsHolder object.
func GetClientsHolder(filenames ...string) *ClientsHolder {
	if clientsHolder.ready {
		return &clientsHolder
	}

	clientsHolder, err := newClientsHolder(filenames...)
	if err != nil {
		logrus.Panic("Failed to create k8s clients holder: ", err)
	}
	return clientsHolder
}

// GetClientsHolder instantiate an ocp client
func newClientsHolder(filenames ...string) (*ClientsHolder, error) { //nolint:funlen // this is a special function with lots of assignments
	logrus.Infof("Creating k8s go-clients holder.")
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	precedence := []string{}
	if len(filenames) > 0 {
		precedence = append(precedence, filenames...)
	}

	loadingRules.Precedence = precedence
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	var err error
	clientsHolder.RestConfig, err = kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("can't instantiate rest config: %s", err)
	}
	DefaultTimeout := 10 * time.Second
	clientsHolder.RestConfig.Timeout = DefaultTimeout

	clientsHolder.DynamicClient, err = dynamic.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate dynamic client (unstructured/dynamic): %s", err)
	}
	clientsHolder.APIExtClient, err = apiextv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate apiextv1: %s", err)
	}
	clientsHolder.OlmClient, err = olmClient.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate olm clientset: %s", err)
	}
	clientsHolder.K8sClient, err = kubernetes.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate k8sclient: %s", err)
	}
	// create the oc client
	clientsHolder.OcpClient, err = clientconfigv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("can't instantiate ocClient: %s", err)
	}

	clientsHolder.ready = true
	return &clientsHolder, nil
}
