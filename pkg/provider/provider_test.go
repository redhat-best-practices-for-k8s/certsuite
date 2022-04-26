// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"fmt"
	"testing"

	"errors"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmFakeClient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sTesting "k8s.io/client-go/testing"
)

var (
	// All this catalogSources and installPlans are used by more than one unit test, so make sure
	// you fully understand them before changing these values.
	// They define runtime objects for 2 CSVs "op1.v1.0.1" and "op2.v2.0.2" that are installed in
	// namespaces ns1 (op1) and ns2 (op1 + op2). So there's one catalogSource + installPlan for
	// each installation. Subscriptions and CSVs are only needed by TestCreateOperators, so they're
	// defined there only.
	catalogSource1 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource1", Namespace: "ns1"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource1Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	catalogSource2 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource2", Namespace: "ns2"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource2Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	catalogSource3 = olmv1Alpha.CatalogSource{
		TypeMeta:   metav1.TypeMeta{Kind: "CatalogSource"},
		ObjectMeta: metav1.ObjectMeta{Name: "catalogSource3", Namespace: "ns2"},
		Spec:       olmv1Alpha.CatalogSourceSpec{Image: "catalogSource3Image"},
		Status:     olmv1Alpha.CatalogSourceStatus{},
	}

	ns1InstallPlan1 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns1Plan1", Namespace: "ns1"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource1", CatalogSourceNamespace: "ns1",
			ClusterServiceVersionNames: []string{"op1.v1.0.1"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath1",
			CatalogSourceRef: &v1.ObjectReference{Name: "catalogSource1", Namespace: "ns1"}}}},
	}

	ns2InstallPlan1 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan1", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource2", CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{"op1.v1.0.1"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath2",
			CatalogSourceRef: &v1.ObjectReference{Name: "catalogSource2", Namespace: "ns2"}}}},
	}

	ns2InstallPlan2 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{Kind: "InstallPlan"}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan2", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{CatalogSource: "catalogSource3", CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{"op2.v2.0.2"}, Approval: olmv1Alpha.ApprovalAutomatic, Approved: true},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath3",
			CatalogSourceRef: &v1.ObjectReference{Name: "catalogSource3", Namespace: "ns2"}}}},
	}
)

func Test_isDaemonSetReady(t *testing.T) {
	type args struct {
		status *appsv1.DaemonSetStatus
	}
	tests := []struct {
		name        string
		args        args
		wantIsReady bool
	}{
		{name: "daemonsetReady",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: true,
		},
		{name: "daemonsetNotReady1",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 3, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
		{name: "daemonsetNotReady2",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 1, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
		{name: "daemonsetNotReady3",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 3, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsReady := isDaemonSetReady(tt.args.status); gotIsReady != tt.wantIsReady {
				t.Errorf("isDaemonSetReady() = %v, want %v", gotIsReady, tt.wantIsReady)
			}
		})
	}
}

func TestGetUID(t *testing.T) {
	testCases := []struct {
		testCID     string
		expectedErr error
		expectedUID string
	}{
		{
			testCID:     "cid://testing",
			expectedErr: nil,
			expectedUID: "testing",
		},
		{
			testCID:     "cid://",
			expectedErr: errors.New("cannot determine container UID"),
			expectedUID: "",
		},
	}

	for _, tc := range testCases {
		c := GetContainer()
		c.Data = &v1.Container{}
		c.Status.ContainerID = tc.testCID
		uid, err := c.GetUID()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedUID, uid)
	}
}

func loadAllOlmRuntimeTestObjects() {
	loadCustomOlmRuntimeTestObjects(&ns1InstallPlan1, &ns2InstallPlan1, &ns2InstallPlan2,
		&catalogSource1, &catalogSource2, &catalogSource3)
}

func loadCustomOlmRuntimeTestObjects(olmObjects ...runtime.Object) {
	_ = clientsholder.GetTestClientsHolder(nil)

	clientsholder.SetupFakeOlmClient(olmObjects)
}

//nolint:funlen
func TestGetInstallPlansInNamespace(t *testing.T) {
	fakeErrorReactionFn := func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("fake error")
	}

	testCases := []struct {
		namespace                   string
		useErrorReactionFn          bool
		expectedNsInstallplans      []olmv1Alpha.InstallPlan
		expectedClusterInstallPlans map[string][]olmv1Alpha.InstallPlan
		expectedErrorStr            string
	}{
		{
			namespace:              "ns1",
			expectedNsInstallplans: []olmv1Alpha.InstallPlan{ns1InstallPlan1},
			expectedClusterInstallPlans: map[string][]olmv1Alpha.InstallPlan{
				"ns1": {ns1InstallPlan1},
			},
		},
		{
			namespace:              "ns2",
			expectedNsInstallplans: []olmv1Alpha.InstallPlan{ns2InstallPlan1, ns2InstallPlan2},
			expectedClusterInstallPlans: map[string][]olmv1Alpha.InstallPlan{
				"ns1": {ns1InstallPlan1},
				"ns2": {ns2InstallPlan1, ns2InstallPlan2},
			},
		},
		{
			namespace:              "ns3",
			useErrorReactionFn:     true,
			expectedNsInstallplans: []olmv1Alpha.InstallPlan{},
			expectedClusterInstallPlans: map[string][]olmv1Alpha.InstallPlan{
				"ns1": {ns1InstallPlan1},
				"ns2": {ns2InstallPlan1, ns2InstallPlan2},
			},
			expectedErrorStr: "unable get installplans in namespace ns3, err: fake error",
		},
	}

	clusterInstallPlans := map[string][]olmv1Alpha.InstallPlan{}
	for _, tc := range testCases {
		loadAllOlmRuntimeTestObjects()
		// In case the TC needs a particular set of olm runtime objects:
		if tc.useErrorReactionFn {
			oc := clientsholder.GetClientsHolder()
			oc.OlmClient.(*olmFakeClient.Clientset).PrependReactor("list", "installplans", fakeErrorReactionFn)
		}

		installPlans, err := getInstallPlansInNamespace(tc.namespace, clusterInstallPlans)
		if tc.expectedErrorStr == "" {
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedNsInstallplans, installPlans)
			assert.Equal(t, tc.expectedClusterInstallPlans, clusterInstallPlans)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, errors.New(tc.expectedErrorStr), err)
		}
	}
}

//nolint:funlen
func TestGetCsvInstallPlans(t *testing.T) {
	loadAllOlmRuntimeTestObjects()

	op4InstallPlan1 := olmv1Alpha.InstallPlan{
		TypeMeta:   metav1.TypeMeta{Kind: "InstallPlan"},
		ObjectMeta: metav1.ObjectMeta{Name: "installPlan1", Namespace: "ns4"},
		Spec:       olmv1Alpha.InstallPlanSpec{ClusterServiceVersionNames: []string{"op4.v4.0.4"}},
		Status: olmv1Alpha.InstallPlanStatus{BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath1",
			CatalogSourceRef: &v1.ObjectReference{Name: "catalogSource1", Namespace: "ns4"}}}},
	}
	op4InstallPlan2 := op4InstallPlan1
	op4InstallPlan2.ObjectMeta.Name = "installPlan2"

	testCases := []struct {
		namespace            string
		csv                  string
		olmObjects           []runtime.Object
		expectedInstallPlans []*olmv1Alpha.InstallPlan
		expectedErrorFmt     string
	}{
		{
			namespace:            "ns1",
			csv:                  "op1.v1.0.1",
			expectedInstallPlans: []*olmv1Alpha.InstallPlan{&ns1InstallPlan1},
		},
		{
			namespace:            "ns2",
			csv:                  "op1.v1.0.1",
			expectedInstallPlans: []*olmv1Alpha.InstallPlan{&ns2InstallPlan1},
		},
		{
			namespace:            "ns2",
			csv:                  "op2.v2.0.2",
			expectedInstallPlans: []*olmv1Alpha.InstallPlan{&ns2InstallPlan2},
		},
		{
			namespace:            "ns1",
			csv:                  "csv2",
			expectedInstallPlans: nil,
			expectedErrorFmt:     "no installplans found for csv %s (ns %s)",
		},
		// Operator with a "bad" install plan.
		{
			namespace: "ns3",
			csv:       "op3.v3.0.3",
			olmObjects: []runtime.Object{
				&olmv1Alpha.InstallPlan{
					TypeMeta:   metav1.TypeMeta{Kind: "InstallPlan"},
					ObjectMeta: metav1.ObjectMeta{Name: "badInstallPlan", Namespace: "ns3"},
					Spec:       olmv1Alpha.InstallPlanSpec{ClusterServiceVersionNames: []string{"op3.v3.0.3"}},
					// This installPlan won't be retrieved as it lacks of the bundle lookups info in the status field.
					Status: olmv1Alpha.InstallPlanStatus{},
				},
			},
			expectedInstallPlans: nil,
			expectedErrorFmt:     "no installplans found for csv %s (ns %s)",
		},
		// Two intallPlans for the same csv, in a new namespace.
		{
			namespace:            "ns4",
			csv:                  "op4.v4.0.4",
			olmObjects:           []runtime.Object{&op4InstallPlan1, &op4InstallPlan2},
			expectedInstallPlans: []*olmv1Alpha.InstallPlan{&op4InstallPlan1, &op4InstallPlan2},
			expectedErrorFmt:     "",
		},
	}

	clusterInstallPlans := map[string][]olmv1Alpha.InstallPlan{}
	for _, tc := range testCases {
		// In case the TC needs a particular set of olm runtime objects:
		if tc.olmObjects != nil {
			loadCustomOlmRuntimeTestObjects(tc.olmObjects...)
		} else {
			loadAllOlmRuntimeTestObjects()
		}
		installPlans, err := getCsvInstallPlans(tc.namespace, tc.csv, clusterInstallPlans)
		assert.Equal(t, tc.expectedInstallPlans, installPlans)
		if tc.expectedErrorFmt == "" {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, err, fmt.Errorf(tc.expectedErrorFmt, tc.csv, tc.namespace))
		}
	}
}

func TestGetCatalogSourceImageIndexFromInstallPlan(t *testing.T) {
	loadAllOlmRuntimeTestObjects()

	testCases := []struct {
		installPlan        *olmv1Alpha.InstallPlan
		expectedImageIndex string
		expectedErrorStr   string
	}{
		{
			installPlan:        &ns1InstallPlan1,
			expectedImageIndex: "catalogSource1Image",
		},
		{
			installPlan:        &ns2InstallPlan1,
			expectedImageIndex: "catalogSource2Image",
		},
		{
			installPlan:        &ns2InstallPlan2,
			expectedImageIndex: "catalogSource3Image",
		},
		{
			installPlan: &olmv1Alpha.InstallPlan{
				Status: olmv1Alpha.InstallPlanStatus{
					BundleLookups: []olmv1Alpha.BundleLookup{
						{Path: "path", CatalogSourceRef: &v1.ObjectReference{Name: "catalogName", Namespace: "notExistingNamespace"}}}},
			},
			expectedImageIndex: "",
			expectedErrorStr:   "failed to get catalogsource: catalogsources.operators.coreos.com \"catalogName\" not found",
		},
	}

	for _, tc := range testCases {
		imageIndex, err := getCatalogSourceImageIndexFromInstallPlan(tc.installPlan)
		assert.Equal(t, tc.expectedImageIndex, imageIndex)
		if tc.expectedErrorStr == "" {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, err, errors.New(tc.expectedErrorStr))
		}
	}
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
		expectedOperators []Operator
		expectedErrorStr  string
	}{
		{
			csvs:              []olmv1Alpha.ClusterServiceVersion{},
			subscriptions:     []olmv1Alpha.Subscription{subscription1},
			expectedOperators: []Operator{},
			expectedErrorStr:  "",
		},
		// ns1: csv1/subs1
		{
			csvs:          []olmv1Alpha.ClusterServiceVersion{op1Ns1},
			subscriptions: []olmv1Alpha.Subscription{subscription1},
			expectedOperators: []Operator{
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
			expectedOperators: []Operator{
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
			expectedOperators: []Operator{
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
			expectedOperators: []Operator{
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

//nolint:funlen
func TestConvertArrayPods(t *testing.T) {
	testCases := []struct {
		testPods     []*v1.Pod
		expectedPods []*Pod
	}{
		{ // Test Case 1 - No containers
			testPods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
				},
			},
			expectedPods: []*Pod{
				{
					Data: &v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
				},
			},
		},
		{ // Test Case 2 - Containers
			testPods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "testcontainer1",
							},
						},
					},
				},
			},
			expectedPods: []*Pod{
				{
					Data: &v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
					Containers: []*Container{
						{
							Data: &v1.Container{
								Name: "testcontainer1",
							},
							Namespace: "testnamespace1",
							Podname:   "testpod1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		convertedArray := ConvertArrayPods(tc.testPods)
		assert.Equal(t, tc.expectedPods[0].Containers, convertedArray[0].Containers)
		assert.Equal(t, tc.expectedPods[0].Data.Name, convertedArray[0].Data.Name)
		assert.Equal(t, tc.expectedPods[0].Data.Namespace, convertedArray[0].Data.Namespace)
	}
}

func TestIsSkipHelmChart(t *testing.T) {
	testCases := []struct {
		testHelmName   string
		testList       []configuration.SkipHelmChartList
		expectedOutput bool
	}{
		{ // Test Case #1 - Helm Chart names match, skipping
			testHelmName: "test1",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: true,
		},
		{ // Test Case #2 - Helm Chart names mismatch, not skipping
			testHelmName: "test2",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: false,
		},
		{ // Test Case #3 - Empty list
			testHelmName:   "test3",
			testList:       []configuration.SkipHelmChartList{},
			expectedOutput: false,
		},
		{ // Test Case #4 - Empty list, helm name empty
			testHelmName:   "",
			testList:       []configuration.SkipHelmChartList{},
			expectedOutput: false,
		},
		{ // Test Case #5 - Helm Chart name missing
			testHelmName: "",
			testList: []configuration.SkipHelmChartList{
				{
					Name: "test1",
				},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, isSkipHelmChart(tc.testHelmName, tc.testList))
	}
}

func TestContainerStringFuncs(t *testing.T) {
	testCases := []struct {
		nodename    string
		namespace   string
		podname     string
		name        string
		containerID string
		runtime     string

		expectedStringOutput     string
		expectedStringLongOutput string
	}{
		{
			nodename:                 "testnode",
			namespace:                "testnamespace",
			podname:                  "testpod",
			name:                     "name1",
			containerID:              "cID1",
			runtime:                  "runtime1",
			expectedStringLongOutput: "node: testnode ns: testnamespace podName: testpod containerName: name1 containerUID: cID1 containerRuntime: runtime1",
			expectedStringOutput:     "container: name1 pod: testpod ns: testnamespace",
		},
		{
			nodename:                 "testnode",
			namespace:                "testnamespace",
			podname:                  "testpod",
			name:                     "name2",
			containerID:              "cID2",
			runtime:                  "runtime2",
			expectedStringLongOutput: "node: testnode ns: testnamespace podName: testpod containerName: name2 containerUID: cID2 containerRuntime: runtime2",
			expectedStringOutput:     "container: name2 pod: testpod ns: testnamespace",
		},
	}

	for _, tc := range testCases {
		c := &Container{
			NodeName:  tc.nodename,
			Namespace: tc.namespace,
			Podname:   tc.podname,
			Data: &v1.Container{
				Name: tc.name,
			},
			Status: v1.ContainerStatus{
				ContainerID: tc.containerID,
			},
			Runtime: tc.runtime,
		}
		assert.Equal(t, tc.expectedStringLongOutput, c.StringLong())
		assert.Equal(t, tc.expectedStringOutput, c.String())
	}
}

func TestDeploymentToString(t *testing.T) {
	assert.Equal(t, "deployment: test1 ns: testNS", DeploymentToString(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "testNS",
		},
	}))
}

func TestStatefulsetToString(t *testing.T) {
	assert.Equal(t, "statefulset: test1 ns: testNS", StatefulsetToString(&appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "testNS",
		},
	}))
}

func TestCsvToString(t *testing.T) {
	assert.Equal(t, "operator csv: test1 ns: testNS", CsvToString(&olmv1Alpha.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "testNS",
		},
	}))
}

func TestPodString(t *testing.T) {
	p := Pod{
		Data: &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "testNS",
			},
		},
	}
	assert.Equal(t, "pod: test1 ns: testNS", p.String())
}

func TestOperatorString(t *testing.T) {
	o := Operator{
		Name:             "test1",
		Namespace:        "testNS",
		SubscriptionName: "sub1",
	}
	assert.Equal(t, "csv: test1 ns:testNS subscription:sub1", o.String())
}

// WorkerLabels = []string{"node-role.kubernetes.io/worker"}
// MasterLabels = []string{"node-role.kubernetes.io/master", "node-role.kubernetes.io/control-plane"}
func TestIsWorkerNode(t *testing.T) {
	testCases := []struct {
		node           *v1.Node
		expectedResult bool
	}{
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": ""}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": "blahblah"}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsWorkerNode(tc.node))
	}
}

func TestIsMasterNode(t *testing.T) {
	testCases := []struct {
		node           *v1.Node
		expectedResult bool
	}{
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": ""}},
			},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/master": "blahblah"}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/control-plane": "blablah"}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/master": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1", "node-role.kubernetes.io/master": "", "node-role.kubernetes.io/control-plane": ""}},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsMasterNode(tc.node))
	}
}
