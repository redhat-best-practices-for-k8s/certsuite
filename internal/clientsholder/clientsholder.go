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

package clientsholder

import (
	"fmt"
	"time"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	olmClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	olmFakeClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/sirupsen/logrus"

	apiextv1c "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"

	cncfNetworkAttachmentv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned/typed/k8s.cni.cncf.io/v1"
	ocpMachine "github.com/openshift/machine-config-operator/pkg/generated/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	k8sFakeClient "k8s.io/client-go/kubernetes/fake"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClientsHolder struct {
	RestConfig           *rest.Config
	DynamicClient        dynamic.Interface
	ScalingClient        scale.ScalesGetter
	APIExtClient         apiextv1.Interface
	OlmClient            olmClient.Interface
	OcpClient            clientconfigv1.ConfigV1Interface
	K8sClient            kubernetes.Interface
	K8sNetworkingClient  networkingv1.NetworkingV1Interface
	CNCFNetworkingClient cncfNetworkAttachmentv1.K8sCniCncfIoV1Interface
	MachineCfg           ocpMachine.Interface
	KubeConfig           []byte
	ready                bool
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
//
//nolint:funlen,gocyclo
func GetTestClientsHolder(k8sMockObjects []runtime.Object, filenames ...string) *ClientsHolder {
	// Build slices of different objects depending on what client
	// is supposed to expect them.
	var k8sClientObjects []runtime.Object
	var k8sExtClientObjects []runtime.Object

	for _, v := range k8sMockObjects {
		// Based on what type of object is, populate certain object slices
		// with what is supported by a certain client.
		// Add more items below if/when needed.
		switch v.(type) {
		// K8s Client Objects
		case *corev1.ServiceAccount:
			k8sClientObjects = append(k8sClientObjects, v)
		case *rbacv1.ClusterRole:
			k8sClientObjects = append(k8sClientObjects, v)
		case *rbacv1.ClusterRoleBinding:
			k8sClientObjects = append(k8sClientObjects, v)
		case *rbacv1.Role:
			k8sClientObjects = append(k8sClientObjects, v)
		case *rbacv1.RoleBinding:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.Pod:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.Service:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.Node:
			k8sClientObjects = append(k8sClientObjects, v)
		case *appsv1.Deployment:
			k8sClientObjects = append(k8sClientObjects, v)
		case *appsv1.StatefulSet:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.ResourceQuota:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.PersistentVolume:
			k8sClientObjects = append(k8sClientObjects, v)
		case *corev1.PersistentVolumeClaim:
			k8sClientObjects = append(k8sClientObjects, v)
		case *policyv1.PodDisruptionBudget:
			k8sClientObjects = append(k8sClientObjects, v)
		case *scalingv1.HorizontalPodAutoscaler:
			k8sClientObjects = append(k8sClientObjects, v)

		// K8s Extension Client Objects
		case *apiextv1c.CustomResourceDefinition:
			k8sExtClientObjects = append(k8sExtClientObjects, v)
		}
	}

	// Add the objects to their corresponding API Clients
	clientsHolder.K8sClient = k8sFakeClient.NewSimpleClientset(k8sClientObjects...)
	clientsHolder.APIExtClient = apiextv1fake.NewSimpleClientset(k8sExtClientObjects...)

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

func createByteArrayKubeConfig(kubeConfig *clientcmdapi.Config) ([]byte, error) {
	yamlBytes, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate yaml bytes from kubeconfig: %w", err)
	}
	return yamlBytes, nil
}

// GetClientsHolder instantiate an ocp client
func newClientsHolder(filenames ...string) (*ClientsHolder, error) { //nolint:funlen,gocyclo // this is a special function with lots of assignments
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
		return nil, fmt.Errorf("cannot instantiate rest config: %s", err)
	}

	// Save merged config to temporary kubeconfig file.
	kubeRawConfig, err := kubeconfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kube raw config: %w", err)
	}

	clientsHolder.KubeConfig, err = createByteArrayKubeConfig(&kubeRawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to byte array kube config reference: %w", err)
	}

	DefaultTimeout := 10 * time.Second
	clientsHolder.RestConfig.Timeout = DefaultTimeout

	clientsHolder.DynamicClient, err = dynamic.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate dynamic client (unstructured/dynamic): %s", err)
	}
	clientsHolder.APIExtClient, err = apiextv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate apiextv1: %s", err)
	}
	clientsHolder.OlmClient, err = olmClient.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate olm clientset: %s", err)
	}
	clientsHolder.K8sClient, err = kubernetes.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate k8sclient: %s", err)
	}
	// create the oc client
	clientsHolder.OcpClient, err = clientconfigv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate ocClient: %s", err)
	}
	clientsHolder.MachineCfg, err = ocpMachine.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate MachineCfg client: %s", err)
	}
	clientsHolder.K8sNetworkingClient, err = networkingv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate k8s networking client: %s", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate discoveryClient: %s", err)
	}
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	gr, err := restmapper.GetAPIGroupResources(clientsHolder.K8sClient.Discovery())
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate GetAPIGroupResources: %s", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	clientsHolder.ScalingClient, err = scale.NewForConfig(clientsHolder.RestConfig, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate ScalesGetter: %s", err)
	}

	clientsHolder.CNCFNetworkingClient, err = cncfNetworkAttachmentv1.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate CNCF networking client")
	}

	clientsHolder.ready = true
	return &clientsHolder, nil
}

type Context struct {
	namespace     string
	podName       string
	containerName string
}

func NewContext(namespace, podName, containerName string) Context {
	return Context{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
	}
}

func (c *Context) GetNamespace() string {
	return c.namespace
}

func (c *Context) GetPodName() string {
	return c.podName
}

func (c *Context) GetContainerName() string {
	return c.containerName
}
