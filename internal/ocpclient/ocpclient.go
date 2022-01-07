package ocpclient

import (
	"time"

	"github.com/sirupsen/logrus"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type OcpClient struct {
	Coreclient *corev1client.CoreV1Client
	ready      bool
}

// NewOcpClient instantiate an ocp client
func NewOcpClient(filenames ...string) OcpClient {
	var ocpClient = OcpClient{}

	if ocpClient.ready {
		return ocpClient
	}
	ocpClient.ready = true

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
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	DefaultTimeout := 10 * time.Second
	restconfig.Timeout = DefaultTimeout
	ocpClient.Coreclient, err = corev1client.NewForConfig(restconfig)
	if err != nil {
		logrus.Panic("can't instantiate corev1client", err)
	}
	return ocpClient
}
