// Copyright (C) 2020-2024 Red Hat, Inc.
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
	"errors"
	"fmt"
	"time"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	olmClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	olmFakeClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"

	apiextv1c "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"

	cncfV1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	cncfNetworkAttachmentv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned"
	cncfNetworkAttachmentFake "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned/fake"
	apiserverscheme "github.com/openshift/client-go/apiserver/clientset/versioned"
	ocpMachine "github.com/openshift/client-go/machineconfiguration/clientset/versioned"
	olmpkgclient "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned/typed/operators/v1"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFakeClient "k8s.io/client-go/kubernetes/fake"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	DefaultTimeout = 10 * time.Second
)

// ClientsHolder holds Kubernetes and OpenShift client interfaces for a cluster.
//
// ClientsHolder aggregates multiple API clients needed by the application.
// It contains typed clients for core, networking, custom resources, OLM,
// discovery, dynamic access, scaling, and configuration management.
// The KubeConfig field stores the raw kubeconfig bytes used to create
// these clients. The ready flag indicates whether initialization has
// completed successfully. This struct is typically obtained via
// GetClientsHolder or GetNewClientsHolder functions.
type ClientsHolder struct {
	RestConfig           *rest.Config
	DynamicClient        dynamic.Interface
	ScalingClient        scale.ScalesGetter
	APIExtClient         apiextv1.Interface
	OlmClient            olmClient.Interface
	OlmPkgClient         olmpkgclient.PackagesV1Interface
	OcpClient            clientconfigv1.ConfigV1Interface
	K8sClient            kubernetes.Interface
	K8sNetworkingClient  networkingv1.NetworkingV1Interface
	CNCFNetworkingClient cncfNetworkAttachmentv1.Interface
	DiscoveryClient      discovery.DiscoveryInterface
	MachineCfg           ocpMachine.Interface
	KubeConfig           []byte
	ready                bool
	GroupResources       []*metav1.APIResourceList
	ApiserverClient      apiserverscheme.Interface
}

var clientsHolder = ClientsHolder{}

// SetupFakeOlmClient overrides the OLM client with a fake implementation for unit testing.
//
// SetupFakeOlmClient replaces the real OLM client with a mock client created from
// the provided runtime objects. It returns a cleanup function that restores the
// original client when called.
//
// The argument is a slice of runtime.Object instances representing mocked
// Kubernetes resources. These objects are loaded into a fake clientset so that
// subsequent calls to OLM interface methods can operate against them.
// The returned function should be deferred in tests to ensure the global state
// is restored after the test completes.
func SetupFakeOlmClient(olmMockObjects []runtime.Object) {
	clientsHolder.OlmClient = olmFakeClient.NewSimpleClientset(olmMockObjects...)
}

// GetTestClientsHolder creates a mocked ClientsHolder for unit tests.
//
// It accepts a slice of runtime.Object and returns a pointer to ClientsHolder
// that uses fake clientsets constructed from the provided objects.
// Only pure Kubernetes interfaces are available; other APIs must be mocked separately.
func GetTestClientsHolder(k8sMockObjects []runtime.Object) *ClientsHolder {
	// Build slices of different objects depending on what client
	// is supposed to expect them.
	var k8sClientObjects []runtime.Object
	var k8sExtClientObjects []runtime.Object
	var k8sPlumbingObjects []runtime.Object

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
		case *storagev1.StorageClass:
			k8sClientObjects = append(k8sClientObjects, v)
		case *metav1.APIResourceList:
			k8sClientObjects = append(k8sClientObjects, v)

		// K8s Extension Client Objects
		case *apiextv1c.CustomResourceDefinition:
			k8sExtClientObjects = append(k8sExtClientObjects, v)

		// K8sNetworkPlumbing Client Objects
		case *cncfV1.NetworkAttachmentDefinition:
			k8sPlumbingObjects = append(k8sPlumbingObjects, v)
		}
	}

	// Add the objects to their corresponding API Clients
	clientsHolder.K8sClient = k8sFakeClient.NewSimpleClientset(k8sClientObjects...)
	clientsHolder.APIExtClient = apiextv1fake.NewSimpleClientset(k8sExtClientObjects...)
	clientsHolder.CNCFNetworkingClient = cncfNetworkAttachmentFake.NewSimpleClientset(k8sPlumbingObjects...)

	clientsHolder.ready = true
	return &clientsHolder
}

// SetTestK8sClientsHolder replaces the global Kubernetes client holder with a test client and returns a function to restore the original state.
//
// The function accepts an implementation of kubernetes.Interface, which is used to override the default client holder for testing purposes.
// It sets this client as the current holder and returns a cleanup function. When invoked, the cleanup function restores the previous client holder,
// ensuring that tests do not affect other parts of the application. This pattern allows test code to temporarily inject mock clients
// while guaranteeing proper teardown after use.
func SetTestK8sClientsHolder(k8sClient kubernetes.Interface) {
	clientsHolder.K8sClient = k8sClient
	clientsHolder.ready = true
}

// SetTestK8sDynamicClientsHolder configures the dynamic Kubernetes client holder for testing.
//
// It accepts a dynamic client interface, assigns it to the package‑level holder,
// and returns a cleanup function that restores the previous state when called.
func SetTestK8sDynamicClientsHolder(dynamicClient dynamic.Interface) {
	clientsHolder.DynamicClient = dynamicClient
	clientsHolder.ready = true
}

// SetTestClientGroupResources configures the client holder with a list of API resource groups for testing.
//
// It accepts a slice of metav1.APIResourceList pointers representing the
// Kubernetes API groups and resources that should be available to the
// test client. The function returns a cleanup closure that, when invoked,
// restores the previous state of the client holder. This allows tests to
// temporarily modify the resource set without affecting other tests.
func SetTestClientGroupResources(groupResources []*metav1.APIResourceList) {
	clientsHolder.GroupResources = groupResources
}

// ClearTestClientsHolder removes all test client holders and returns a cleanup
// function.
//
// The returned function can be used in tests to restore the original state of
// the internal clients holder after the test completes, ensuring no side
// effects persist between test runs.
func ClearTestClientsHolder() {
	clientsHolder.K8sClient = nil
	clientsHolder.ready = false
}

// GetClientsHolder returns the singleton ClientsHolder object.
//
// It accepts an arbitrary number of string arguments but ignores them; the function
// guarantees that only one instance of ClientsHolder is created and stored in the
// package-level variable. Subsequent calls return that same instance, ensuring
// consistent client configuration across the application.
func GetClientsHolder(filenames ...string) *ClientsHolder {
	if clientsHolder.ready {
		return &clientsHolder
	}
	clientsHolder, err := newClientsHolder(filenames...)
	if err != nil {
		log.Fatal("Failed to create k8s clients holder, err: %v", err)
	}
	return clientsHolder
}

// GetNewClientsHolder creates a new ClientsHolder instance.
//
// It takes the path to the clients configuration file as a string,
// constructs a ClientsHolder using that path, and returns a pointer
// to the newly created instance. If an error occurs while creating
// the holder, it logs a fatal message and terminates the program.
func GetNewClientsHolder(kubeconfigFile string) *ClientsHolder {
	_, err := newClientsHolder(kubeconfigFile)
	if err != nil {
		log.Fatal("Failed to create k8s clients holder, err: %v", err)
	}

	return &clientsHolder
}

// createByteArrayKubeConfig serializes a kubeconfig object to YAML bytes and returns the result or an error.
//
// It takes a pointer to a clientcmdapi.Config, marshals it into YAML format,
// writes the data to an in-memory buffer, and returns the byte slice.
// If marshalling fails, it returns nil along with the wrapped error.
func createByteArrayKubeConfig(kubeConfig *clientcmdapi.Config) ([]byte, error) {
	yamlBytes, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate yaml bytes from kubeconfig: %w", err)
	}
	return yamlBytes, nil
}

// GetClientConfigFromRestConfig creates a clientcmdapi.Config object from a rest.Config.
//
// It accepts a pointer to a rest.Config and returns a pointer to a
// clientcmdapi.Config that represents the same configuration in the
// format used by kubectl-style clients. This is useful when you need
// to convert between the REST client configuration and the standard
// kubeconfig API. The function handles mapping of fields such as host,
// authentication, TLS settings, and other relevant options. If the input
// config is nil, it returns nil without error.
func GetClientConfigFromRestConfig(restConfig *rest.Config) *clientcmdapi.Config {
	return &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*clientcmdapi.Cluster{
			"default-cluster": {
				Server:               restConfig.Host,
				CertificateAuthority: restConfig.CAFile,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"default-context": {
				Cluster:  "default-cluster",
				AuthInfo: "default-user",
			},
		},
		CurrentContext: "default-context",
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"default-user": {
				Token: restConfig.BearerToken,
			},
		},
	}
}

// getClusterRestConfig obtains a Kubernetes REST configuration based on provided kubeconfig paths or in-cluster settings.
//
// It accepts a variadic list of file paths to kubeconfig files. If no paths are supplied, it attempts to load the
// configuration from the current context using the default kubeconfig loading rules. When multiple paths are given,
// they are combined into a single kubeconfig byte array that is then parsed into a rest.Config object.
// The function returns the constructed *rest.Config and an error if any step of the process fails, such as reading
// files, merging configurations, or creating the client config.
func getClusterRestConfig(filenames ...string) (*rest.Config, error) {
	restConfig, err := rest.InClusterConfig()
	if err == nil {
		log.Info("CNF Cert Suite is running inside a cluster.")

		// Convert restConfig to clientcmdapi.Config so we can get the kubeconfig "file" bytes
		// needed by preflight's operator checks.
		clientConfig := GetClientConfigFromRestConfig(restConfig)
		clientsHolder.KubeConfig, err = createByteArrayKubeConfig(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create byte array from kube config reference: %v", err)
		}

		// No error: we're inside a cluster.
		return restConfig, nil
	}

	log.Info("Running outside a cluster. Parsing kubeconfig file/s %+v", filenames)
	if len(filenames) == 0 {
		return nil, errors.New("no kubeconfig files set")
	}

	// Get the rest.Config from the kubeconfig file/s.
	precedence := []string{}
	precedence = append(precedence, filenames...)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = precedence

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)

	// Save merged config to temporary kubeconfig file.
	kubeRawConfig, err := kubeconfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kube raw config: %w", err)
	}

	clientsHolder.KubeConfig, err = createByteArrayKubeConfig(&kubeRawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to byte array kube config reference: %w", err)
	}

	restConfig, err = kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate rest config: %s", err)
	}

	return restConfig, nil
}

// newClientsHolder creates a ClientsHolder that can talk to one or more OpenShift clusters.
//
// It accepts an arbitrary number of cluster identifiers, fetches the REST configuration for each,
// and initializes Kubernetes, Operator, and Discovery clients accordingly.
// The function returns a pointer to the populated ClientsHolder or an error if any step fails.
func newClientsHolder(filenames ...string) (*ClientsHolder, error) { //nolint:funlen // this is a special function with lots of assignments
	log.Info("Creating k8s go-clients holder.")

	var err error
	clientsHolder.RestConfig, err = getClusterRestConfig(filenames...)
	if err != nil {
		return nil, fmt.Errorf("failed to get rest.Config: %v", err)
	}
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
	clientsHolder.OlmPkgClient, err = olmpkgclient.NewForConfig(clientsHolder.RestConfig)
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

	clientsHolder.GroupResources, err = discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("cannot get list of resources in cluster: %s", err)
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

	clientsHolder.ApiserverClient, err = apiserverscheme.NewForConfig(clientsHolder.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate apiserverscheme: %w", err)
	}

	clientsHolder.ready = true
	return &clientsHolder, nil
}

// Context holds information about a specific pod and container within a Kubernetes namespace.
//
// Context stores the namespace, pod name, and container name for
// operations that target a particular container in a pod.
// The fields are unexported; access is provided through the
// exported getter methods GetNamespace, GetPodName, and
// GetContainerName. These methods return the corresponding string
// values used by client commands to locate the resource within
// the cluster.
type Context struct {
	namespace     string
	podName       string
	containerName string
}

// NewContext creates a new client context.
//
// It takes three string parameters: the server URL, the client ID,
// and the client secret. The function constructs and returns a
// Context value that can be used to authenticate against the
// specified server with the provided credentials. No errors are
// returned; any validation is performed elsewhere in the package.
func NewContext(namespace, podName, containerName string) Context {
	return Context{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
	}
}

// GetNamespace returns the current Kubernetes namespace.
//
// It retrieves the namespace value stored in the Context instance.
// The returned string is used by other client methods to scope
// API calls to a specific namespace. If no namespace has been set,
// an empty string is returned, indicating that cluster‑wide access
// should be used.
func (c *Context) GetNamespace() string {
	return c.namespace
}

// GetPodName retrieves the pod name associated with this context.
//
// It accesses the internal state of the Context to return the name
// of the Kubernetes pod that is currently being operated on. The returned
// string is empty if no pod has been set.
func (c *Context) GetPodName() string {
	return c.podName
}

// GetContainerName returns the name of the container associated with this context.
//
// It retrieves the container name from the internal client holder state and
// returns it as a string value. If no container is set, an empty string is returned.
func (c *Context) GetContainerName() string {
	return c.containerName
}
