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

package provider

import (
	"errors"
	"testing"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCsvToString(t *testing.T) {
	assert.Equal(t, "operator csv: test1 ns: testNS", CsvToString(&olmv1Alpha.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "testNS",
		},
	}))
}

func TestOperatorString(t *testing.T) {
	o := Operator{
		Name:             "test1",
		Namespace:        "testNS",
		SubscriptionName: "sub1",
	}
	assert.Equal(t, "csv: test1 ns:testNS subscription:sub1", o.String())
}

//nolint:funlen
func TestCreateOperators(t *testing.T) {
	// op1 in namespace ns1
	op1Ns1 := olmv1Alpha.ClusterServiceVersion{
		TypeMeta:   metav1.TypeMeta{Kind: "ClusterServiceVersion"},
		ObjectMeta: metav1.ObjectMeta{Name: "op1.v1.0.1", Namespace: "ns1"},
	}

	// op1 in namespace ns2
	op1Ns2 := olmv1Alpha.ClusterServiceVersion{
		TypeMeta:   metav1.TypeMeta{Kind: "ClusterServiceVersion"},
		ObjectMeta: metav1.ObjectMeta{Name: "op1.v1.0.1", Namespace: "ns2"},
	}

	// op2 in namespace ns2
	op2Ns2 := olmv1Alpha.ClusterServiceVersion{
		TypeMeta:   metav1.TypeMeta{Kind: "ClusterServiceVersion"},
		ObjectMeta: metav1.ObjectMeta{Name: "op2.v2.0.2", Namespace: "ns2"},
	}

	subscription1 := olmv1Alpha.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription"},
		ObjectMeta: metav1.ObjectMeta{Name: "subs1", Namespace: "ns1"},
		Spec:       &olmv1Alpha.SubscriptionSpec{Package: "op1", CatalogSource: "catalogSource1"},
		Status:     olmv1Alpha.SubscriptionStatus{InstalledCSV: "op1.v1.0.1"},
	}

	subscription2 := olmv1Alpha.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription"},
		ObjectMeta: metav1.ObjectMeta{Name: "subs2", Namespace: "ns2"},
		Spec:       &olmv1Alpha.SubscriptionSpec{Package: "op1", CatalogSource: "catalogSource2"},
		Status:     olmv1Alpha.SubscriptionStatus{InstalledCSV: "op1.v1.0.1"},
	}

	subscription3 := olmv1Alpha.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription"},
		ObjectMeta: metav1.ObjectMeta{Name: "subs3", Namespace: "ns2"},
		Spec:       &olmv1Alpha.SubscriptionSpec{Package: "op2", CatalogSource: "catalogSource3"},
		Status:     olmv1Alpha.SubscriptionStatus{InstalledCSV: "op2.v2.0.2"},
	}

	testCases := []struct {
		csvs              []olmv1Alpha.ClusterServiceVersion
		subscriptions     []olmv1Alpha.Subscription
		olmObjects        []runtime.Object
		expectedOperators []*Operator
		expectedErrorStr  string
	}{
		{
			csvs:              []olmv1Alpha.ClusterServiceVersion{},
			subscriptions:     []olmv1Alpha.Subscription{subscription1},
			expectedOperators: []*Operator{},
			expectedErrorStr:  "",
		},
		// ns1: csv1/subs1
		{
			csvs:          []olmv1Alpha.ClusterServiceVersion{op1Ns1},
			subscriptions: []olmv1Alpha.Subscription{subscription1},
			expectedOperators: []*Operator{
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns1",
					Csv:              &op1Ns1,
					SubscriptionName: "subs1",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns1Plan1",
							BundleImage: "lookuppath1",
							IndexImage:  "catalogSource1Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource1",
					Version: "v1.0.1",
				},
			},
		},
		// ns1: csv1/subs1 - installPlan not found.
		{
			csvs:              []olmv1Alpha.ClusterServiceVersion{op1Ns1},
			subscriptions:     []olmv1Alpha.Subscription{subscription1},
			olmObjects:        make([]runtime.Object, 0),
			expectedOperators: nil,
			expectedErrorStr:  "failed to get installPlans for csv op1.v1.0.1 (ns ns1), err: no installplans found for csv op1.v1.0.1 (ns ns1)",
		},
		// ns1: csv1/subs1 - bundleImage not found.
		{
			csvs:              []olmv1Alpha.ClusterServiceVersion{op1Ns1},
			subscriptions:     []olmv1Alpha.Subscription{subscription1},
			olmObjects:        []runtime.Object{&ns1InstallPlan1},
			expectedOperators: nil,
			expectedErrorStr:  "failed to get installPlan image index for csv op1.v1.0.1 (ns ns1) installPlan ns1Plan1, err: failed to get catalogsource: catalogsources.operators.coreos.com \"catalogSource1\" not found",
		},
		// ns1: csv1/subs1, ns2: csv2 (without subscription)
		{
			csvs:          []olmv1Alpha.ClusterServiceVersion{op1Ns1, op1Ns2},
			subscriptions: []olmv1Alpha.Subscription{subscription1},
			expectedOperators: []*Operator{
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns1",
					Csv:              &op1Ns1,
					SubscriptionName: "subs1",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns1Plan1",
							BundleImage: "lookuppath1",
							IndexImage:  "catalogSource1Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource1",
					Version: "v1.0.1",
				},
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns2",
					Csv:              &op1Ns2,
					SubscriptionName: "",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns2Plan1",
							BundleImage: "lookuppath2",
							IndexImage:  "catalogSource2Image",
						},
					},
					Package: "",
					Version: "v1.0.1",
				},
			},
		},
		// ns1: csv1/subs1, ns2: csv2/subs2
		{
			csvs:          []olmv1Alpha.ClusterServiceVersion{op1Ns1, op1Ns2},
			subscriptions: []olmv1Alpha.Subscription{subscription1, subscription2},
			expectedOperators: []*Operator{
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns1",
					Csv:              &op1Ns1,
					SubscriptionName: "subs1",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns1Plan1",
							BundleImage: "lookuppath1",
							IndexImage:  "catalogSource1Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource1",
					Version: "v1.0.1",
				},
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns2",
					Csv:              &op1Ns2,
					SubscriptionName: "subs2",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns2Plan1",
							BundleImage: "lookuppath2",
							IndexImage:  "catalogSource2Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource2",
					Version: "v1.0.1",
				},
			},
		},
		// ns1: csv1/subs1, ns2: csv2/subs2 + csv3/subs3
		{
			csvs:          []olmv1Alpha.ClusterServiceVersion{op1Ns1, op1Ns2, op2Ns2},
			subscriptions: []olmv1Alpha.Subscription{subscription1, subscription2, subscription3},
			expectedOperators: []*Operator{
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns1",
					Csv:              &op1Ns1,
					SubscriptionName: "subs1",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns1Plan1",
							BundleImage: "lookuppath1",
							IndexImage:  "catalogSource1Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource1",
					Version: "v1.0.1",
				},
				{
					Name:             "op1.v1.0.1",
					Namespace:        "ns2",
					Csv:              &op1Ns2,
					SubscriptionName: "subs2",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns2Plan1",
							BundleImage: "lookuppath2",
							IndexImage:  "catalogSource2Image",
						},
					},
					Package: "op1",
					Org:     "catalogSource2",
					Version: "v1.0.1",
				},
				{
					Name:             "op2.v2.0.2",
					Namespace:        "ns2",
					Csv:              &op2Ns2,
					SubscriptionName: "subs3",
					InstallPlans: []CsvInstallPlan{
						{
							Name:        "ns2Plan2",
							BundleImage: "lookuppath3",
							IndexImage:  "catalogSource3Image",
						},
					},
					Package: "op2",
					Org:     "catalogSource3",
					Version: "v2.0.2",
				},
			},
		},
	}

	for _, tc := range testCases {
		// In case the TC needs a particular set of olm runtime objects:
		if tc.olmObjects != nil {
			loadCustomOlmRuntimeTestObjects(tc.olmObjects...)
		} else {
			loadAllOlmRuntimeTestObjects()
		}

		ops, err := createOperators(tc.csvs, tc.subscriptions)
		assert.Equal(t, tc.expectedOperators, ops)
		if tc.expectedErrorStr == "" {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, errors.New(tc.expectedErrorStr), err)
		}
	}
}
