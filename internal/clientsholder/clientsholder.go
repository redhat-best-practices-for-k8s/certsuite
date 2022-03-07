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
	"time"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	clientOlm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientsHolder struct {
	RestConfig    *rest.Config
	Coreclient    *corev1client.CoreV1Client
	ClientConfig  clientconfigv1.ConfigV1Interface
	DynamicClient dynamic.Interface
	APIExtClient  apiextv1.ApiextensionsV1Interface
	OlmClient     *clientOlm.Clientset
	AppsClients   *appv1client.AppsV1Client
	K8sClient     *kubernetes.Clientset
	OClient       *clientconfigv1.ConfigV1Client

	ready bool
}

var clientsHolder = ClientsHolder{}

// NewClientsHolder instantiate an ocp client
func NewClientsHolder(filenames ...string) *ClientsHolder { //nolint:funlen // this is a special function with lots of assignments
	if clientsHolder.ready {
		return &clientsHolder
	}

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
		panic(err)
	}
	DefaultTimeout := 10 * time.Second
	clientsHolder.RestConfig.Timeout = DefaultTimeout

	clientsHolder.Coreclient, err = corev1client.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate corev1client: ", err)
	}
	clientsHolder.ClientConfig, err = clientconfigv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate corev1client: ", err)
	}
	clientsHolder.DynamicClient, err = dynamic.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate dynamic client (unstructured/dynamic): ", err)
	}
	clientsHolder.APIExtClient, err = apiextv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate dynamic client (unstructured/dynamic): ", err)
	}
	clientsHolder.OlmClient, err = clientOlm.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate olm clientset: ", err)
	}
	clientsHolder.AppsClients, err = appv1client.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate appv1client", err)
	}
	// create the k8sclient
	clientsHolder.K8sClient, err = kubernetes.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate k8sclient", err)
	}
	// create the oc client
	clientsHolder.OClient, err = clientconfigv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate ocClient", err)
	}

	clientsHolder.ready = true
	return &clientsHolder
}
