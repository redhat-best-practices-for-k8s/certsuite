package rbac

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AutomountToken struct {
	serviceAccountName string
	podNamespace       string
	ClientHolder       *clientsholder.ClientsHolder
}

type ServiceAccountTokenStatus struct {
	Name       string
	TokenSet   bool
	TokenValue bool
}

func NewAutomountTester(serviceAccountName, podNamespace string, ch *clientsholder.ClientsHolder) *AutomountToken {
	return &AutomountToken{
		serviceAccountName: serviceAccountName,
		podNamespace:       podNamespace,
		ClientHolder:       ch,
	}
}

func (at *AutomountToken) AutomountServiceAccountSetOnSA() (ServiceAccountTokenStatus, error) {
	sa, err := at.ClientHolder.K8sClient.CoreV1().ServiceAccounts(at.podNamespace).Get(context.TODO(), at.serviceAccountName, v1.GetOptions{})
	if err != nil {
		logrus.Errorf("executing serviceaccount command failed with error: %s", err)
		return ServiceAccountTokenStatus{}, err
	}
	return DetermineStatus(sa.AutomountServiceAccountToken), nil
}

func DetermineStatus(automountField *bool) ServiceAccountTokenStatus {
	saStatus := ServiceAccountTokenStatus{}
	if automountField == nil {
		saStatus.TokenSet = false
		saStatus.TokenValue = false
	} else {
		saStatus.TokenSet = true
		saStatus.TokenValue = *automountField
	}
	return saStatus
}
