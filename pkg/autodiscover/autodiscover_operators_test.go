// Copyright (C) 2023-2024 Red Hat, Inc.

package autodiscover

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	fakeolmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	fakeolmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned/fake"
)

func TestGetAllNamespaces(t *testing.T) {
	generateNamespaces := func(ns []string) []*corev1.Namespace {
		var namespaces []*corev1.Namespace
		for _, n := range ns {
			namespaces = append(namespaces, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      n,
					Namespace: n,
				},
			})
		}
		return namespaces
	}

	testCases := []struct {
		testNamespaces     []string
		expectedNamespaces []string
	}{
		{
			testNamespaces:     []string{"ns1"},
			expectedNamespaces: []string{"ns1"},
		},
		{
			testNamespaces:     []string{"ns1", "ns2"},
			expectedNamespaces: []string{"ns1", "ns2"},
		},
	}

	for _, tc := range testCases {
		// Generate the namespaces for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range generateNamespaces(tc.testNamespaces) {
			testRuntimeObjects = append(testRuntimeObjects, n)
		}

		clientSet := fake.NewSimpleClientset(testRuntimeObjects...)
		namespaces, err := getAllNamespaces(clientSet.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedNamespaces, namespaces)
	}
}

func TestGetAllCatalogSources(t *testing.T) {
	generateCatalogSource := func(name string) *olmv1alpha1.CatalogSource {
		return &olmv1alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		testCatalogSources []string
	}{
		{
			testCatalogSources: []string{"cs1"},
		},
		{
			testCatalogSources: []string{"cs1", "cs2"},
		},
		{
			testCatalogSources: []string{},
		},
	}

	for _, tc := range testCases {
		// Generate the catalog sources for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testCatalogSources {
			testRuntimeObjects = append(testRuntimeObjects, generateCatalogSource(n))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		catalogSources := getAllCatalogSources(client.OperatorsV1alpha1())
		assert.Equal(t, len(tc.testCatalogSources), len(catalogSources))
		for i := range catalogSources {
			assert.Equal(t, tc.testCatalogSources[i], catalogSources[i].Name)
		}
	}
}

func TestGetAllInstallPlans(t *testing.T) {
	generateInstallPlan := func(name string) *olmv1alpha1.InstallPlan {
		return &olmv1alpha1.InstallPlan{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		testInstallPlans []string
	}{
		{
			testInstallPlans: []string{"ip1"},
		},
		{
			testInstallPlans: []string{"ip1", "ip2"},
		},
		{
			testInstallPlans: []string{},
		},
	}

	for _, tc := range testCases {
		// Generate the install plans for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testInstallPlans {
			testRuntimeObjects = append(testRuntimeObjects, generateInstallPlan(n))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		installPlans := getAllInstallPlans(client.OperatorsV1alpha1())
		assert.Equal(t, len(tc.testInstallPlans), len(installPlans))
		for i := range installPlans {
			assert.Equal(t, tc.testInstallPlans[i], installPlans[i].Name)
		}
	}
}

func TestGetAllPackageManifests(t *testing.T) {
	generatePackageManifest := func(name string) *olmpkgv1.PackageManifest {
		return &olmpkgv1.PackageManifest{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		testPackageManifests []string
	}{
		{testPackageManifests: []string{"pm1"}},
		{testPackageManifests: []string{"pm1", "pm2"}},
		{testPackageManifests: []string{}},
	}

	for _, tc := range testCases {
		// Generate the package manifests for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testPackageManifests {
			testRuntimeObjects = append(testRuntimeObjects, generatePackageManifest(n))
		}

		client := fakeolmpkgv1.NewSimpleClientset(testRuntimeObjects...)
		packageManifests := getAllPackageManifests(client.OperatorsV1().PackageManifests(""))
		assert.Equal(t, len(tc.testPackageManifests), len(packageManifests))
		for i := range packageManifests {
			assert.Equal(t, tc.testPackageManifests[i], packageManifests[i].Name)
		}
	}
}

func TestFindSubscriptions(t *testing.T) {
	generateSubscription := func(name, namespace string) *olmv1alpha1.Subscription {
		return &olmv1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testSubscriptions []string
	}{
		{testSubscriptions: []string{"sub1"}},
		{testSubscriptions: []string{"sub1", "sub2"}},
		{testSubscriptions: []string{}},
	}

	for _, tc := range testCases {
		// Generate the subscriptions for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testSubscriptions {
			testRuntimeObjects = append(testRuntimeObjects, generateSubscription(n, "default"))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		subscriptions := findSubscriptions(client.OperatorsV1alpha1(), []string{""})
		assert.Equal(t, len(tc.testSubscriptions), len(subscriptions))
		for i := range subscriptions {
			assert.Equal(t, tc.testSubscriptions[i], subscriptions[i].Name)
		}
	}
}

func TestGetAllOperators(t *testing.T) {
	generateClusterServiceVersion := func(name string) *olmv1alpha1.ClusterServiceVersion {
		return &olmv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		testClusterServiceVersions []string
	}{
		{testClusterServiceVersions: []string{"csv1"}},
		{testClusterServiceVersions: []string{"csv1", "csv2"}},
		{testClusterServiceVersions: []string{}},
	}

	for _, tc := range testCases {
		// Generate the cluster service versions for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testClusterServiceVersions {
			testRuntimeObjects = append(testRuntimeObjects, generateClusterServiceVersion(n))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		clusterServiceVersions, err := getAllOperators(client.OperatorsV1alpha1())
		assert.Nil(t, err)
		assert.Equal(t, len(tc.testClusterServiceVersions), len(clusterServiceVersions))
		for i := range clusterServiceVersions {
			assert.Equal(t, tc.testClusterServiceVersions[i], clusterServiceVersions[i].Name)
		}
	}
}

func TestIsIstioServiceMeshInstalled(t *testing.T) {
	generateDeployment := func(name, namespace string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testDeployment *appsv1.Deployment
		expectedResult bool
	}{
		{
			testDeployment: generateDeployment("istiod", "istio-system"),
			expectedResult: true,
		},
		{
			testDeployment: generateDeployment("istiod", "default"),
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		// Generate the deployment for the test
		clientSet := fake.NewSimpleClientset(tc.testDeployment)
		result := isIstioServiceMeshInstalled(clientSet.AppsV1(), []string{"istio-system"})
		assert.Equal(t, tc.expectedResult, result)
	}

	result := isIstioServiceMeshInstalled(nil, []string{"not-istio-system"})
	assert.False(t, result)
}

func TestFindOperatorsMatchingAtLeastOneLabel(t *testing.T) {
	generateClusterServiceVersion := func(name, namespace string) *olmv1alpha1.ClusterServiceVersion {
		return &olmv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					"key": "value",
				},
			},
		}
	}

	testCases := []struct {
		testClusterServiceVersions []string
	}{
		{testClusterServiceVersions: []string{"csv1"}},
		{testClusterServiceVersions: []string{"csv1", "csv2"}},
		{testClusterServiceVersions: []string{}},
	}

	for _, tc := range testCases {
		// Generate the cluster service versions for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testClusterServiceVersions {
			testRuntimeObjects = append(testRuntimeObjects, generateClusterServiceVersion(n, "default"))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		labels := []labelObject{{LabelKey: "key", LabelValue: "value"}}
		clusterServiceVersions := findOperatorsMatchingAtLeastOneLabel(client.OperatorsV1alpha1(), labels, configuration.Namespace{Name: "default"})
		assert.Equal(t, len(tc.testClusterServiceVersions), len(clusterServiceVersions.Items))
		for i := range clusterServiceVersions.Items {
			assert.Equal(t, tc.testClusterServiceVersions[i], clusterServiceVersions.Items[i].Name)
		}
	}
}

func TestFindOperatorsByLabels(t *testing.T) {
	generateClusterServiceVersion := func(name, namespace string) *olmv1alpha1.ClusterServiceVersion {
		return &olmv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Annotations: map[string]string{
					"olm.operatorNamespace": "default",
				},
				Labels: map[string]string{
					"key": "value",
				},
			},
		}
	}

	testCases := []struct {
		testClusterServiceVersions []string
	}{
		{testClusterServiceVersions: []string{"csv1"}},
		{testClusterServiceVersions: []string{"csv1", "csv2"}},
		{testClusterServiceVersions: []string{}},
	}

	for _, tc := range testCases {
		// Generate the cluster service versions for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range tc.testClusterServiceVersions {
			testRuntimeObjects = append(testRuntimeObjects, generateClusterServiceVersion(n, "default"))
		}

		client := fakeolmv1alpha1.NewSimpleClientset(testRuntimeObjects...)
		labels := []labelObject{{LabelKey: "key", LabelValue: "value"}}
		namespaces := []configuration.Namespace{{Name: "default"}}
		clusterServiceVersions := findOperatorsByLabels(client.OperatorsV1alpha1(), labels, namespaces)
		assert.Equal(t, len(tc.testClusterServiceVersions), len(clusterServiceVersions))
		for i := range clusterServiceVersions {
			assert.Equal(t, tc.testClusterServiceVersions[i], clusterServiceVersions[i].Name)
		}
	}
}
