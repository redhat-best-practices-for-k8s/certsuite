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

// ClientsHolder Holds configured Kubernetes API clients for cluster interaction
//
// This structure aggregates multiple client interfaces, including core,
// dynamic, extension, networking, and OLM clients, along with configuration
// data such as the REST config and kubeconfig bytes. It provides a single point
// from which tests or utilities can execute commands inside pods, query
// resources, or manipulate cluster objects. The ready flag indicates whether
// the holder has been fully initialized.
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

// SetupFakeOlmClient Replaces the real OLM client with a fake for testing
//
// This function takes a slice of Kubernetes objects that represent mocked OLM
// resources. It constructs a new fake client set containing those objects and
// assigns it to the package's client holder, enabling tests to interact with
// OLM APIs without contacting an actual cluster.
func SetupFakeOlmClient(olmMockObjects []runtime.Object) {
	clientsHolder.OlmClient = olmFakeClient.NewSimpleClientset(olmMockObjects...)
}

// GetTestClientsHolder Creates a mocked client holder for unit tests
//
// This function accepts a slice of runtime objects that represent Kubernetes
// resources and builds separate slices for each supported client type. It then
// initializes fake clients with these objects, marks the holder as ready, and
// returns it for use in testing scenarios.
//
//nolint:funlen,gocyclo
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

// SetTestK8sClientsHolder Stores a Kubernetes client for test usage
//
// This function assigns the provided Kubernetes interface to an internal holder
// and marks it as ready. It is intended for tests that require a mock or real
// client without interacting with a live cluster. After execution, other
// components can retrieve the stored client from the holder.
func SetTestK8sClientsHolder(k8sClient kubernetes.Interface) {
	clientsHolder.K8sClient = k8sClient
	clientsHolder.ready = true
}

// SetTestK8sDynamicClientsHolder Assigns a test Kubernetes dynamic client to the internal holder
//
// This function stores the provided dynamic client instance in an internal
// structure used by tests, marking the holder as ready for use. It replaces any
// existing client reference and enables subsequent code that relies on the
// dynamic client to operate against this test instance.
func SetTestK8sDynamicClientsHolder(dynamicClient dynamic.Interface) {
	clientsHolder.DynamicClient = dynamicClient
	clientsHolder.ready = true
}

// SetTestClientGroupResources Stores a list of API resource group definitions
//
// This function receives an array of API resource lists and assigns it to the
// internal holder used by the client package. It updates the shared state that
// other components reference when interacting with Kubernetes groups. No value
// is returned, and the operation replaces any previously stored resources.
func SetTestClientGroupResources(groupResources []*metav1.APIResourceList) {
	clientsHolder.GroupResources = groupResources
}

// ClearTestClientsHolder Resets the Kubernetes client and marks holder as not ready
//
// This function clears the stored Kubernetes client reference, setting it to
// nil, and updates an internal flag to indicate that the holder is no longer
// ready for use. It does not return a value and has no parameters. After
// calling this, any attempt to access the client will need reinitialization.
func ClearTestClientsHolder() {
	clientsHolder.K8sClient = nil
	clientsHolder.ready = false
}

// GetClientsHolder Returns a cached instance of the Kubernetes clients holder
//
// This function checks whether the global ClientsHolder has already been
// initialized and ready; if so, it returns that instance immediately. If not,
// it attempts to create a new holder by calling the internal constructor with
// any provided configuration filenames. Errors during creation are logged as
// fatal, terminating the program. The resulting holder is returned for use by
// other parts of the application.
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

// GetNewClientsHolder Creates a Kubernetes clients holder from the provided kubeconfig
//
// The function takes a file path to a kubeconfig, uses an internal constructor
// to instantiate a ClientsHolder with all necessary API clients, and logs a
// fatal error if construction fails. On success it returns a pointer to the
// fully initialized holder for use by other components.
func GetNewClientsHolder(kubeconfigFile string) *ClientsHolder {
	_, err := newClientsHolder(kubeconfigFile)
	if err != nil {
		log.Fatal("Failed to create k8s clients holder, err: %v", err)
	}

	return &clientsHolder
}

// createByteArrayKubeConfig Converts a Kubernetes configuration into YAML byte array
//
// The function takes a pointer to a client configuration structure and
// serializes it into its YAML representation using the client-go library. It
// returns the resulting bytes along with any error that occurs during
// serialization, allowing callers to use the data as a kubeconfig file in
// memory.
func createByteArrayKubeConfig(kubeConfig *clientcmdapi.Config) ([]byte, error) {
	yamlBytes, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate yaml bytes from kubeconfig: %w", err)
	}
	return yamlBytes, nil
}

// GetClientConfigFromRestConfig Creates a kubeconfig configuration from a REST client
//
// It accepts a Kubernetes rest.Config pointer and builds an equivalent
// clientcmdapi.Config structure containing cluster, context, and authentication
// information. The resulting config includes the server URL, certificate
// authority path, bearer token, and sets a default cluster and context for use
// by other components.
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

// getClusterRestConfig Retrieves a Kubernetes REST configuration from in‑cluster or kubeconfig files
//
// The function first attempts to obtain an in‑cluster configuration; if
// successful it converts that config into a kubeconfig byte array for
// downstream use and returns the rest.Config. If not running inside a cluster,
// it requires one or more kubeconfig file paths, merges them with precedence
// rules, creates a temporary kubeconfig representation, extracts the REST
// client configuration from it, and returns that configuration along with any
// error encountered.
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

// newClientsHolder Creates a holder of Kubernetes client interfaces based on provided kubeconfig files
//
// It loads a rest configuration from the given kubeconfig paths or in-cluster
// settings, then initializes numerous typed and dynamic clients for API
// extensions, OLM, OpenShift, networking, scaling, and CNCF networking. The
// function also retrieves cluster resource listings and prepares a REST mapper
// for scale operations. Upon successful setup, it marks the holder as ready and
// returns it; otherwise an error is returned.
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

// Context Represents a target container within a pod
//
// This structure holds the namespace, pod name, and container name used when
// executing commands inside Kubernetes pods. It provides accessor methods to
// retrieve each field value. The context is typically created with NewContext
// and passed to command execution functions.
type Context struct {
	namespace     string
	podName       string
	containerName string
}

// NewContext Creates a context for running commands inside a specific pod container
//
// This function takes the namespace, pod name, and container name of a probe
// pod and returns a Context object that holds those values. The returned
// Context is used by other components to target the correct container when
// executing shell commands via the client holder. No additional processing or
// validation occurs; it simply packages the identifiers into the struct.
func NewContext(namespace, podName, containerName string) Context {
	return Context{
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
	}
}

// Context.GetNamespace retrieves the namespace from the context
//
// This method accesses the internal namespace field of a Context instance and
// returns it as a string. It does not modify any state or perform additional
// logic, simply exposing the value stored during context creation.
func (c *Context) GetNamespace() string {
	return c.namespace
}

// Context.GetPodName returns the pod name stored in the context
//
// This method retrieves and returns the podName field from a Context instance.
// It takes no arguments and always yields a string representing the current pod
// identifier used for Kubernetes API calls.
func (c *Context) GetPodName() string {
	return c.podName
}

// Context.GetContainerName Returns the current pod's container name
//
// This method retrieves the container name stored in the Context object. It
// accesses an internal field that holds the name of the container to which
// commands will be executed or operations will target. The returned string is
// used by other components when interacting with Kubernetes pods.
func (c *Context) GetContainerName() string {
	return c.containerName
}
