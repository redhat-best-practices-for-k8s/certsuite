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

package testhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewContainerReportObject(t *testing.T) {
	testCases := []struct {
		testNamespace     string
		testPodName       string
		testContainerName string
		testReason        string
		testIsCompliant   bool
		expectedOutput    *ReportObject
	}{
		{
			testNamespace:     "testNamespace",
			testPodName:       "testPodName",
			testContainerName: "testContainerName",
			testReason:        "testReason",
			testIsCompliant:   true,
			expectedOutput: &ReportObject{
				ObjectType: "Container",
				ObjectFieldsKeys: []string{
					Namespace,
					PodName,
					ContainerName,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testNamespace",
					"testPodName",
					"testContainerName",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewContainerReportObject(testCase.testNamespace, testCase.testPodName, testCase.testContainerName, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewCertifiedContainerReportObject(t *testing.T) {
	reportObj := NewCertifiedContainerReportObject(provider.ContainerImageIdentifier{
		Registry:   "testRegistry",
		Repository: "testRepository",
		Tag:        "testTag",
		Digest:     "testDigest",
	}, "testReason", true)

	assert.Equal(t, ContainerImageType, reportObj.ObjectType)
	for _, reportKey := range reportObj.ObjectFieldsKeys {
		assert.Contains(t, []string{ImageRegistry, ImageRepo, ImageTag, ImageDigest, ReasonForCompliance}, reportKey)
	}

	for _, reportValue := range reportObj.ObjectFieldsValues {
		assert.Contains(t, []string{"testRegistry", "testRepository", "testTag", "testDigest", "testReason"}, reportValue)
	}
}

func TestNewNodeReportObject(t *testing.T) {
	testCases := []struct {
		testNodeName    string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testNodeName:    "testNodeName",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: NodeType,
				ObjectFieldsKeys: []string{
					Name,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testNodeName",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewNodeReportObject(testCase.testNodeName, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewClusterVersionReportObject(t *testing.T) {
	testCases := []struct {
		testVersion     string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testVersion:     "testVersion",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: OCPClusterType,
				ObjectFieldsKeys: []string{
					OCPClusterVersionType,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testVersion",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewClusterVersionReportObject(testCase.testVersion, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewPodReportObject(t *testing.T) {
	testCases := []struct {
		testNamespace   string
		testPodName     string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testNamespace:   "testNamespace",
			testPodName:     "testPodName",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: PodType,
				ObjectFieldsKeys: []string{
					Namespace,
					PodName,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testNamespace",
					"testPodName",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewPodReportObject(testCase.testNamespace, testCase.testPodName, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewTaintReportObject(t *testing.T) {
	testCases := []struct {
		testTaint       string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testTaint:       "testTaint",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: TaintType,
				ObjectFieldsKeys: []string{
					NodeType,
					ReasonForCompliance,
					TaintBit,
				},
				ObjectFieldsValues: []string{
					"node1",
					"testTaint",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewTaintReportObject(testCase.testTaint, "node1", testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewHelmChartReportObject(t *testing.T) {
	testCases := []struct {
		testChart       string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testChart:       "testChart",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: HelmType,
				ObjectFieldsKeys: []string{
					Name,
					Namespace,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testChart",
					"testReason",
					"helm1",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewHelmChartReportObject(testCase.testChart, "helm1", testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewOperatorReportObject(t *testing.T) {
	testCases := []struct {
		testOperator    string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testOperator:    "testOperator",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: OperatorType,
				ObjectFieldsKeys: []string{
					Name,
					Namespace,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testOperator",
					"testReason",
					"operator1",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewOperatorReportObject(testCase.testOperator, "operator1", testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewDeploymentReportObject(t *testing.T) {
	testCases := []struct {
		testNamespace   string
		testDeployment  string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testNamespace:   "testNamespace",
			testDeployment:  "testDeployment",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: DeploymentType,
				ObjectFieldsKeys: []string{
					Namespace,
					DeploymentName,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testNamespace",
					"testDeployment",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewDeploymentReportObject(testCase.testNamespace, testCase.testDeployment, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewStatefulSetReportObject(t *testing.T) {
	testCases := []struct {
		testNamespace   string
		testStatefulSet string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testNamespace:   "testNamespace",
			testStatefulSet: "testStatefulSet",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: StatefulSetType,
				ObjectFieldsKeys: []string{
					Namespace,
					StatefulSetName,
					ReasonForCompliance,
				},
				ObjectFieldsValues: []string{
					"testNamespace",
					"testStatefulSet",
					"testReason",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewStatefulSetReportObject(testCase.testNamespace, testCase.testStatefulSet, testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestNewCrdReportObject(t *testing.T) {
	testCases := []struct {
		testCrd         string
		testReason      string
		testIsCompliant bool
		expectedOutput  *ReportObject
	}{
		{
			testCrd:         "testCrd",
			testReason:      "testReason",
			testIsCompliant: true,
			expectedOutput: &ReportObject{
				ObjectType: CustomResourceDefinitionType,
				ObjectFieldsKeys: []string{
					CustomResourceDefinitionName,
					ReasonForCompliance,
					CustomResourceDefinitionVersion,
				},
				ObjectFieldsValues: []string{
					"testCrd",
					"testReason",
					"version1",
				},
			},
		},
	}

	for _, testCase := range testCases {
		reportObj := NewCrdReportObject(testCase.testCrd, "version1", testCase.testReason, testCase.testIsCompliant)

		assert.Equal(t, testCase.expectedOutput.ObjectType, reportObj.ObjectType)
		for _, reportKey := range reportObj.ObjectFieldsKeys {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsKeys, reportKey)
		}

		for _, reportValue := range reportObj.ObjectFieldsValues {
			assert.Contains(t, testCase.expectedOutput.ObjectFieldsValues, reportValue)
		}
	}
}

func TestSetContainerProcessValues(t *testing.T) {
	reportObj := NewContainerReportObject("namespace1", "pod1", "container1", "reason1", true)
	reportObj.SetContainerProcessValues("policy1", "priority1", "command1")
	assert.Equal(t, ContainerProcessType, reportObj.ObjectType)
}

func TestResultToString(t *testing.T) {
	testCases := []struct {
		input          int
		expectedResult string
	}{
		{input: SUCCESS, expectedResult: "SUCCESS"},
		{input: FAILURE, expectedResult: "FAILURE"},
		{input: ERROR, expectedResult: "ERROR"},
		{input: 1337, expectedResult: ""},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, ResultToString(tc.input))
	}
}

func TestEqual(t *testing.T) {
	testCases := []struct {
		testSlice1     []*ReportObject
		testSlice2     []*ReportObject
		expectedResult bool
	}{
		{
			testSlice1:     []*ReportObject{{ObjectType: "test1"}},
			testSlice2:     []*ReportObject{{ObjectType: "test1"}},
			expectedResult: true,
		},
		{
			testSlice1:     []*ReportObject{{ObjectType: "test1"}},
			testSlice2:     []*ReportObject{{ObjectType: "test2"}},
			expectedResult: false,
		},
		{
			testSlice1:     []*ReportObject{{ObjectType: "test1"}},
			testSlice2:     []*ReportObject{{ObjectType: "test1"}, {ObjectType: "test2"}},
			expectedResult: false,
		},
		{
			testSlice1:     []*ReportObject{{ObjectType: "test1"}, {ObjectType: "test2"}},
			testSlice2:     nil,
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expectedResult, Equal(testCase.testSlice1, testCase.testSlice2))
	}
}

func TestGetNoServicesUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Services: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Services: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoServicesUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetDaemonSetFailedToSpawnSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{DaemonsetFailedToSpawn: true},
			expectedResult: true,
		},
		{
			testEnv:        &provider.TestEnvironment{DaemonsetFailedToSpawn: false},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetDaemonSetFailedToSpawnSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetSharedProcessNamespacePodsSkipFn(t *testing.T) {
	newProviderPod := func(shareProcessNamespace *bool) *provider.Pod {
		return &provider.Pod{
			Pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					ShareProcessNamespace: shareProcessNamespace,
				},
			},
		}
	}

	trueVar := true
	falseVar := false

	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				newProviderPod(nil),
			}},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				newProviderPod(&trueVar),
			}},
			expectedResult: false,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				newProviderPod(&falseVar),
			}},
			expectedResult: true,
		},
	}

	for _, testCase := range testCases {
		results, _ := GetSharedProcessNamespacePodsSkipFn(testCase.testEnv)()
		assert.Equal(t, testCase.expectedResult, results)
	}
}

func TestGetNoContainersUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Containers: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoContainersUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoPodsUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Pods: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoPodsUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoDeploymentsUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Deployments: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Deployments: []*provider.Deployment{
				{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoDeploymentsUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoStatefulSetsUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{StatefulSets: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{StatefulSets: []*provider.StatefulSet{
				{
					StatefulSet: &appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoStatefulSetsUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoCrdsUnderTestSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Crds: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Crds: []*apiextv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoCrdsUnderTestSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoNamespacesSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Namespaces: nil},
			expectedResult: true,
		},
		{
			testEnv:        &provider.TestEnvironment{Namespaces: []string{"test1"}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoNamespacesSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoRolesSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Roles: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoRolesSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoPersistentVolumesSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{PersistentVolumes: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{PersistentVolumes: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoPersistentVolumesSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNotEnoughWorkersSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Nodes: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{
				Nodes: map[string]provider.Node{
					"test1": {
						Data: &corev1.Node{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test1",
								Labels: map[string]string{
									"node-role.kubernetes.io/worker": "",
								},
							},
						},
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNotEnoughWorkersSkipFn(testCase.testEnv, 1)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetPodsWithoutAffinityRequiredLabelSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Pods: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
							Labels: map[string]string{
								"AffinityRequired": "true",
							},
						},
					},
				},
			}},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
							Labels: map[string]string{
								"UnrelatedLabel": "true",
							},
						},
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetPodsWithoutAffinityRequiredLabelSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoAffinityRequiredPodsSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{Pods: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
							Labels: map[string]string{
								"AffinityRequired": "true",
							},
						},
					},
				},
			}},
			expectedResult: false,
		},
		{
			testEnv: &provider.TestEnvironment{Pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
							Labels: map[string]string{
								"AffinityRequired": "false",
							},
						},
					},
				},
			}},
			expectedResult: true,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoAffinityRequiredPodsSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoStorageClassesSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv:        &provider.TestEnvironment{StorageClassList: nil},
			expectedResult: true,
		},
		{
			testEnv: &provider.TestEnvironment{StorageClassList: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			},
			},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoStorageClassesSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoPersistentVolumeClaimsSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{testEnv: &provider.TestEnvironment{PersistentVolumeClaims: nil}, expectedResult: true},
		{
			testEnv: &provider.TestEnvironment{PersistentVolumeClaims: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			}},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoPersistentVolumeClaimsSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoBareMetalNodesSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{
			testEnv: &provider.TestEnvironment{
				Nodes: map[string]provider.Node{
					"test1": {
						Data: &corev1.Node{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test1",
							},
							Spec: corev1.NodeSpec{
								ProviderID: "baremetalhost://test1",
							},
						},
					},
				},
			}, expectedResult: false,
		},
		{
			testEnv: &provider.TestEnvironment{
				Nodes: map[string]provider.Node{
					"test1": {
						Data: &corev1.Node{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test1",
							},
							Spec: corev1.NodeSpec{
								ProviderID: "test1",
							},
						},
					},
				},
			}, expectedResult: true,
		},
	}

	for _, testCase := range testCases {
		testFunc := GetNoBareMetalNodesSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoIstioSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{testEnv: &provider.TestEnvironment{IstioServiceMeshFound: false}, expectedResult: true},
		{testEnv: &provider.TestEnvironment{IstioServiceMeshFound: true}, expectedResult: false},
	}

	for _, testCase := range testCases {
		testFunc := GetNoIstioSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}

func TestGetNoOperatorsSkipFn(t *testing.T) {
	testCases := []struct {
		testEnv        *provider.TestEnvironment
		expectedResult bool
	}{
		{testEnv: &provider.TestEnvironment{Operators: nil}, expectedResult: true},
		{testEnv: &provider.TestEnvironment{Operators: []*provider.Operator{
			{
				Name: "test1",
			},
		}}, expectedResult: false},
	}

	for _, testCase := range testCases {
		testFunc := GetNoOperatorsSkipFn(testCase.testEnv)
		result, _ := testFunc()
		assert.Equal(t, testCase.expectedResult, result)
	}
}
