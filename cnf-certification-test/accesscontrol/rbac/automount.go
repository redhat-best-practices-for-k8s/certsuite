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
	// Create a default SATS struct.
	saStatus := ServiceAccountTokenStatus{
		TokenSet:   false,
		TokenValue: false,
	}

	// Collect all of the service accounts that live in the target namespace.
	saList, saErr := at.ClientHolder.K8sClient.CoreV1().ServiceAccounts(at.podNamespace).List(context.TODO(), v1.ListOptions{})
	if saErr != nil {
		logrus.Errorf("executing serviceaccount command failed with error: %s", saErr)
		return saStatus, saErr
	}

	// Look through the list of service accounts for the desired SA and determine its status.
	for index := range saList.Items {
		if at.serviceAccountName == saList.Items[index].Name {
			return DetermineStatus(saList.Items[index].AutomountServiceAccountToken), nil
		}
	}

	return saStatus, nil
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
