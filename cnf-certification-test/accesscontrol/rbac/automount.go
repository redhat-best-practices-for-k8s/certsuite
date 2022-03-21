// Copyright (C) 2022 Red Hat, Inc.
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

package rbac

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AutomountTokenFuncs interface {
	AutomountServiceAccountSetOnSA(serviceAccountName, podNamespace string) (*bool, error)
}
type AutomountToken struct {
	ch *clientsholder.ClientsHolder
}

func NewAutomountTokenTester(ch *clientsholder.ClientsHolder) *AutomountToken {
	return &AutomountToken{
		ch: ch,
	}
}

func (at *AutomountToken) AutomountServiceAccountSetOnSA(serviceAccountName, podNamespace string) (*bool, error) {
	sa, err := at.ch.K8sClient.CoreV1().ServiceAccounts(podNamespace).Get(context.TODO(), serviceAccountName, v1.GetOptions{})
	if err != nil {
		logrus.Errorf("executing serviceaccount command failed with error: %s", err)
		return nil, err
	}
	return sa.AutomountServiceAccountToken, nil
}
