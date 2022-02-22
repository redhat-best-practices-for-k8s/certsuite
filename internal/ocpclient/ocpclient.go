package ocpclient

import (
	"time"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	clientOlm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type OcpClient struct {
	Coreclient    *corev1client.CoreV1Client
	ClientConfig  clientconfigv1.ConfigV1Interface
	DynamicClient dynamic.Interface
	APIExtClient  apiextensionsv1beta1.ApiextensionsV1beta1Interface
	OlmClient     *clientOlm.Clientset
	RestConfig    *rest.Config

	ready bool
}

var ocpClient = OcpClient{}

// NewOcpClient instantiate an ocp client
func NewOcpClient(filenames ...string) OcpClient { //nolint:funlen // this is a special function with lots of assignments
	if ocpClient.ready {
		return ocpClient
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
	ocpClient.RestConfig, err = kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	DefaultTimeout := 10 * time.Second
	ocpClient.RestConfig.Timeout = DefaultTimeout

	ocpClient.Coreclient, err = corev1client.NewForConfig(ocpClient.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate corev1client: ", err)
	}
	ocpClient.ClientConfig, err = clientconfigv1.NewForConfig(ocpClient.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate corev1client: ", err)
	}
	ocpClient.DynamicClient, err = dynamic.NewForConfig(ocpClient.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate dynamic client (unstructured/dynamic): ", err)
	}
	ocpClient.APIExtClient, err = apiextensionsv1beta1.NewForConfig(ocpClient.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate dynamic client (unstructured/dynamic): ", err)
	}
	ocpClient.OlmClient, err = clientOlm.NewForConfig(ocpClient.RestConfig)
	if err != nil {
		logrus.Panic("can't instantiate olm clientset: ", err)
	}

	ocpClient.ready = true
	return ocpClient
}
