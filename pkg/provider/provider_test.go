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

package provider

import (
	"errors"
	"os"
	"reflect"
	"testing"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		TypeMeta: metav1.TypeMeta{
			Kind: "InstallPlan",
		}, ObjectMeta: metav1.ObjectMeta{Name: "ns1Plan1", Namespace: "ns1"},
		Spec: olmv1Alpha.InstallPlanSpec{
			CatalogSource:          "catalogSource1",
			CatalogSourceNamespace: "ns1",
			ClusterServiceVersionNames: []string{
				"op1.v1.0.1",
			},
			Approval: olmv1Alpha.ApprovalAutomatic,
			Approved: true,
		},
		Status: olmv1Alpha.InstallPlanStatus{
			BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath1",
				CatalogSourceRef: &corev1.ObjectReference{
					Name:      "catalogSource1",
					Namespace: "ns1",
				}}},
		},
	}

	ns2InstallPlan1 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{
			Kind: "InstallPlan",
		}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan1", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{
			CatalogSource:          "catalogSource2",
			CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{
				"op1.v1.0.1",
			},
			Approval: olmv1Alpha.ApprovalAutomatic,
			Approved: true,
		},
		Status: olmv1Alpha.InstallPlanStatus{
			BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath2",
				CatalogSourceRef: &corev1.ObjectReference{
					Name:      "catalogSource2",
					Namespace: "ns2",
				}}},
		},
	}

	ns2InstallPlan2 = olmv1Alpha.InstallPlan{
		TypeMeta: metav1.TypeMeta{
			Kind: "InstallPlan",
		}, ObjectMeta: metav1.ObjectMeta{Name: "ns2Plan2", Namespace: "ns2"},
		Spec: olmv1Alpha.InstallPlanSpec{
			CatalogSource:          "catalogSource3",
			CatalogSourceNamespace: "ns2",
			ClusterServiceVersionNames: []string{
				"op2.v2.0.2",
			},
			Approval: olmv1Alpha.ApprovalAutomatic,
			Approved: true,
		},
		Status: olmv1Alpha.InstallPlanStatus{
			BundleLookups: []olmv1Alpha.BundleLookup{{Path: "lookuppath3",
				CatalogSourceRef: &corev1.ObjectReference{
					Name:      "catalogSource3",
					Namespace: "ns2",
				}}},
		},
	}
)

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
		c := NewContainer()
		c.Status.ContainerID = tc.testCID
		uid, err := c.GetUID()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedUID, uid)
	}
}

func TestConvertArrayPods(t *testing.T) {
	testCases := []struct {
		testPods     []*corev1.Pod
		expectedPods []*Pod
	}{
		{ // Test Case 1 - No containers
			testPods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
				},
			},
			expectedPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
				},
			},
		},
		{ // Test Case 2 - Containers
			testPods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testpod1",
						Namespace: "testnamespace1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "testcontainer1",
							},
						},
					},
				},
			},
			expectedPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testpod1",
							Namespace: "testnamespace1",
						},
					},
					Containers: []*Container{
						{
							Container: &corev1.Container{
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
		assert.Equal(t, tc.expectedPods[0].Name, convertedArray[0].Name)
		assert.Equal(t, tc.expectedPods[0].Namespace, convertedArray[0].Namespace)
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
			Container: &corev1.Container{
				Name: tc.name,
			},
			Status: corev1.ContainerStatus{
				ContainerID: tc.containerID,
			},
			Runtime: tc.runtime,
		}
		assert.Equal(t, tc.expectedStringLongOutput, c.StringLong())
		assert.Equal(t, tc.expectedStringOutput, c.String())
	}
}

func TestIsWorkerNode(t *testing.T) {
	testCases := []struct {
		node           *corev1.Node
		expectedResult bool
	}{
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/master": ""},
				},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/worker": ""},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/worker": "blahblah"},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":                         "fakeValue1",
						"node-role.kubernetes.io/worker": "",
					},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":                         "fakeValue1",
						"node-role.kubernetes.io/worker": "",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		node := Node{Data: tc.node}
		assert.Equal(t, tc.expectedResult, node.IsWorkerNode())
	}
}

func TestIsMasterNode(t *testing.T) {
	testCases := []struct {
		node           *corev1.Node
		expectedResult bool
	}{
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label1": "fakeValue1"}},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/worker": ""},
				},
			},
			expectedResult: false,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/master": ""},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/master": "blahblah"},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": "blablah"},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":                         "fakeValue1",
						"node-role.kubernetes.io/master": "",
					},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":                                "fakeValue1",
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
			expectedResult: true,
		},
		{
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":                                "fakeValue1",
						"node-role.kubernetes.io/master":        "",
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		node := Node{Data: tc.node}
		assert.Equal(t, tc.expectedResult, node.IsMasterNode())
	}
}

func TestGetNodeCount(t *testing.T) {
	generateEnv := func(isMaster bool) *TestEnvironment {
		key := "node-role.kubernetes.io/worker"
		if isMaster {
			key = "node-role.kubernetes.io/master"
		}

		return &TestEnvironment{
			Nodes: map[string]Node{
				"node1": {
					Data: &corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{key: ""},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testIsMaster bool
	}{
		{
			testIsMaster: true,
		},
		{
			testIsMaster: false,
		},
	}

	for _, tc := range testCases {
		tEnv := generateEnv(tc.testIsMaster)

		if tc.testIsMaster {
			assert.Equal(t, 1, tEnv.GetMasterCount())
			assert.Equal(t, 0, tEnv.GetWorkerCount())
		} else {
			assert.Equal(t, 1, tEnv.GetWorkerCount())
			assert.Equal(t, 0, tEnv.GetMasterCount())
		}
	}
}

func TestIsRTKernel(t *testing.T) {
	generateNode := func(kernel string) *Node {
		return &Node{
			Data: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						KernelVersion: kernel,
					},
				},
			},
		}
	}

	testCases := []struct {
		testKernel     string
		expectedOutput bool
	}{
		{ // Test Case #1 - Kernel is RT
			testKernel:     "3.10.0-1127.10.1.rt56.1106.el7",
			expectedOutput: true,
		},
		{ // Test Case #2 - Kernel is standard
			testKernel:     "3.10.0-1127.10.1.1106.el7",
			expectedOutput: false,
		},
		{ // Test Case #3 - Kernel string empty
			testKernel:     "",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		n := generateNode(tc.testKernel)
		assert.Equal(t, n.IsRTKernel(), tc.expectedOutput)
	}
}

func TestIsRHCOS(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput bool
	}{
		{
			testImageName:  "Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			expectedOutput: true,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: false,
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		assert.Equal(t, tc.expectedOutput, node.IsRHCOS())
	}
}

func TestIsRHEL(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput bool
	}{
		{
			testImageName:  "Red Hat Enterprise Linux 8.5 (Ootpa)",
			expectedOutput: true,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: false,
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		assert.Equal(t, tc.expectedOutput, node.IsRHEL())
	}
}

func TestGetRHCOSVersion(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput string
		expectedErr    error
	}{
		{
			testImageName:  "Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			expectedOutput: "4.10.14",
			expectedErr:    nil,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 20.04"),
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 21.10"),
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}

		origValue := rhcosRelativePath
		rhcosRelativePath = "%s/../../cnf-certification-test/platform/operatingsystem/files/rhcos_version_map" // for testing only
		result, err := node.GetRHCOSVersion()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
		rhcosRelativePath = origValue
	}
}

func TestGetCSCOSVersion(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput string
		expectedErr    error
	}{
		{
			testImageName:  "CentOS Stream CoreOS 413.92.202303061740-0 (Plow)",
			expectedOutput: "413.92.202303061740-0",
			expectedErr:    nil,
		},
		{
			testImageName:  "Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			expectedOutput: "",
			expectedErr: errors.New(
				"invalid OS type: Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)",
			),
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 20.04"),
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 21.10"),
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}

		result, err := node.GetCSCOSVersion()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestGetRHELVersion(t *testing.T) {
	testCases := []struct {
		testImageName  string
		expectedOutput string
		expectedErr    error
	}{
		{
			testImageName:  "Red Hat Enterprise Linux 8.5 (Ootpa)",
			expectedOutput: "8.5",
			expectedErr:    nil,
		},
		{
			testImageName:  "Ubuntu 20.04",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 20.04"),
		},
		{
			testImageName:  "Ubuntu 21.10",
			expectedOutput: "",
			expectedErr:    errors.New("invalid OS type: Ubuntu 21.10"),
		},
	}

	for _, tc := range testCases {
		node := Node{
			Data: &corev1.Node{
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						OSImage: tc.testImageName,
					},
				},
			},
		}
		result, err := node.GetRHELVersion()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestBuildImageWithVersion(t *testing.T) {
	testCases := []struct {
		repoVar         string
		supportImageVar string
		expectedOutput  string
	}{
		{
			repoVar:         "test1",
			supportImageVar: "image1",
			expectedOutput:  "test1/image1",
		},
		{
			repoVar:         "",
			supportImageVar: "",
			expectedOutput:  "quay.io/testnetworkfunction/debug-partner:4.5.5",
		},
	}

	defer func() {
		os.Unsetenv("TNF_PARTNER_REPO")
		os.Unsetenv("SUPPORT_IMAGE")
	}()

	for _, tc := range testCases {
		os.Setenv("TNF_PARTNER_REPO", tc.repoVar)
		os.Setenv("SUPPORT_IMAGE", tc.supportImageVar)
		assert.Equal(t, tc.expectedOutput, buildImageWithVersion())
	}
}

func Test_buildContainerImageSource(t *testing.T) {
	type args struct {
		urlImage   string
		urlImageID string
	}
	tests := []struct {
		name       string
		args       args
		wantSource ContainerImageIdentifier
	}{
		{name: "image has tag and registry",
			args: args{
				urlImage:   "quay.io/testnetworkfunction/cnf-test-partner:latest",
				urlImageID: "quay.io/testnetworkfunction/cnf-test-partner@sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
			wantSource: ContainerImageIdentifier{
				Registry:   "quay.io",
				Repository: "testnetworkfunction/cnf-test-partner",
				Tag:        "latest",
				Digest:     "sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
		},
		{name: "digest in image and imageID do not match and no tag",
			args: args{
				urlImage:   "quay.io/testnetworkfunction/cnf-test-partner@sha256:2341c96eba68e2dbf9498a2fe7b96465665465465a2d0d811168667a919345",
				urlImageID: "quay.io/testnetworkfunction/cnf-test-partner@sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
			wantSource: ContainerImageIdentifier{
				Registry:   "",
				Repository: "",
				Tag:        "",
				Digest:     "sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
		},
		{name: "image with no tag and no registry",
			args: args{
				urlImage:   "httpd:2.4.57",
				urlImageID: "quay.io/httpd:2.4.57@sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
			wantSource: ContainerImageIdentifier{
				Registry:   "",
				Repository: "httpd",
				Tag:        "2.4.57",
				Digest:     "sha256:2341c96eba68e2dbf9498a2fe7b95e6f9b84f6ac15fa2d0d811168667a919a49",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSource := buildContainerImageSource(tt.args.urlImage, tt.args.urlImageID); !reflect.DeepEqual(
				gotSource,
				tt.wantSource,
			) {
				t.Errorf("buildContainerImageSource() = %v, want %v", gotSource, tt.wantSource)
			}
		})
	}
}

func TestGetBaremetalNodes(t *testing.T) {
	type fields struct {
		Nodes map[string]Node
	}
	tests := []struct {
		name   string
		fields fields
		want   []Node
	}{
		{
			name: "test1",
			fields: fields{
				Nodes: map[string]Node{
					"node1": {
						Data: &corev1.Node{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
							Spec: corev1.NodeSpec{
								ProviderID: "baremetalhost://aaaaaa",
							},
						},
					},
				},
			},
			want: []Node{
				{
					Data: &corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Spec: corev1.NodeSpec{
							ProviderID: "baremetalhost://aaaaaa",
						},
					},
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				Nodes: map[string]Node{
					"node1": {
						Data: &corev1.Node{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
							Spec: corev1.NodeSpec{
								ProviderID: "Virtual://aaaaaa",
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &TestEnvironment{
				Nodes: tt.fields.Nodes,
			}
			if got := env.GetBaremetalNodes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestEnvironment.GetBaremetalNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPodContainers(t *testing.T) {
	pod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-k8s-0",
			Namespace: "openshift-monitoring",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "prometheus",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
				},
				{
					Name:  "config-reloader",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
				},
				{
					Name:  "thanos-sidecar",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
				},
				{
					Name:  "prometheus-proxy",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
				},
				{
					Name:  "kube-rbac-proxy",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
				},
				{
					Name:  "kube-rbac-proxy-thanos",
					Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
				},
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "config-reloader",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
					ContainerID:  "cri-o://57c22c7e3eb0c906dd517bab91059db671b6ac03b70a44f2839ac1ece3c03db3",
				},
				{
					Name:         "kube-rbac-proxy",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					ContainerID:  "cri-o://33ed1ea7f25bea164cde6a196ada613101c0958c1c4907502d62008f245fbf35",
				},
				{
					Name:         "kube-rbac-proxy-thanos",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					ContainerID:  "cri-o://5ed2c50da83c19bf8aab38178af5c44438c1abc9bd6478480b9be0282b1a1cd6",
				},
				{
					Name:         "prometheus",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
					ContainerID:  "cri-o://2c0ffc1f7c522edfa786ff4fc04c9d8f487926e6159dc2c5c708d2bfdc2619a7",
				},
				{
					Name:         "prometheus-proxy",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
					ContainerID:  "cri-o://a54fbb740338618de653a20596571490caa93e28be9d582314d898bf8e163d22",
				},
				{
					Name:         "thanos-sidecar",
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Time{}}},
					Ready:        true,
					RestartCount: 2,
					Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
					ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
					ContainerID:  "cri-o://1c90c74031e6565fbbab74ef1bdc6cb9ca9b48ca728c42d7041f1e4e71a95ed0",
				},
			},
		},
	}

	type args struct {
		aPod          *corev1.Pod
		useIgnoreList bool
	}
	tests := []struct {
		name              string
		args              args
		wantContainerList []*Container
	}{
		{
			name: "pod1normal",
			args: args{aPod: &pod1, useIgnoreList: false},
			wantContainerList: []*Container{
				{
					Container: &corev1.Container{
						Name:  "prometheus",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
					},
					Status: corev1.ContainerStatus{
						Name: "prometheus",
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0",
						ContainerID:  "cri-o://2c0ffc1f7c522edfa786ff4fc04c9d8f487926e6159dc2c5c708d2bfdc2619a7",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "2c0ffc1f7c522edfa786ff4fc04c9d8f487926e6159dc2c5c708d2bfdc2619a7",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:064300d031bcc2423e2dc5eb32c9606c869942aa41f239b9c561c0b038d3d8f0"},
				},
				{
					Container: &corev1.Container{
						Name:  "config-reloader",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
					},
					Status: corev1.ContainerStatus{
						Name:         "config-reloader",
						State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918",
						ContainerID:  "cri-o://57c22c7e3eb0c906dd517bab91059db671b6ac03b70a44f2839ac1ece3c03db3",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "57c22c7e3eb0c906dd517bab91059db671b6ac03b70a44f2839ac1ece3c03db3",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:55fdb6cdbcb7c25d8206eba68ef8676fc86949ad6965ce5f9bc1afe18e0c6918"},
				},
				{
					Container: &corev1.Container{
						Name:  "thanos-sidecar",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
					},
					Status: corev1.ContainerStatus{
						Name:         "thanos-sidecar",
						State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f",
						ContainerID:  "cri-o://1c90c74031e6565fbbab74ef1bdc6cb9ca9b48ca728c42d7041f1e4e71a95ed0",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "1c90c74031e6565fbbab74ef1bdc6cb9ca9b48ca728c42d7041f1e4e71a95ed0",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:ac3558b2758c283c355f30b1255793f1363b86c199569de55a6e599a39135b1f"},
				},
				{
					Container: &corev1.Container{
						Name:  "prometheus-proxy",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
					},
					Status: corev1.ContainerStatus{
						Name:         "prometheus-proxy",
						State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3",
						ContainerID:  "cri-o://a54fbb740338618de653a20596571490caa93e28be9d582314d898bf8e163d22",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "a54fbb740338618de653a20596571490caa93e28be9d582314d898bf8e163d22",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:df80d3297a5530801baf25e2b4e2e265fe094c43fe1fa959f83e380b56a3f0c3"},
				},
				{
					Container: &corev1.Container{
						Name:  "kube-rbac-proxy",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					},
					Status: corev1.ContainerStatus{
						Name:         "kube-rbac-proxy",
						State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
						ContainerID:  "cri-o://33ed1ea7f25bea164cde6a196ada613101c0958c1c4907502d62008f245fbf35",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "33ed1ea7f25bea164cde6a196ada613101c0958c1c4907502d62008f245fbf35",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52"},
				},
				{
					Container: &corev1.Container{
						Name:  "kube-rbac-proxy-thanos",
						Image: "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
					},
					Status: corev1.ContainerStatus{
						Name:         "kube-rbac-proxy-thanos",
						State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						Ready:        true,
						RestartCount: 2,
						Image:        "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
						ImageID:      "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52",
						ContainerID:  "cri-o://5ed2c50da83c19bf8aab38178af5c44438c1abc9bd6478480b9be0282b1a1cd6",
					},
					Namespace:                "openshift-monitoring",
					Podname:                  "prometheus-k8s-0",
					Runtime:                  "cri-o",
					UID:                      "5ed2c50da83c19bf8aab38178af5c44438c1abc9bd6478480b9be0282b1a1cd6",
					ContainerImageIdentifier: ContainerImageIdentifier{Digest: "sha256:e2b2c89aedaa44964e4cf003ef94963da2e773ace08e601592078adefa482b52"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotContainerList := getPodContainers(tt.args.aPod, tt.args.useIgnoreList); !reflect.DeepEqual(gotContainerList, tt.wantContainerList) {
				t.Errorf("getPodContainers() = %v, want %v", gotContainerList, tt.wantContainerList)
			}
		})
	}
}
