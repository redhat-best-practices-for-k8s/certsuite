package cnis

import (
	"testing"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	"gotest.tools/v3/assert"
)

func TestGetNetworksMap(t *testing.T) {
	testCases := []struct {
		name            string
		nets            []claim.CNINetwork
		expectedNetsMap map[string]*claim.CNINetwork
	}{
		{
			name:            "empty slice",
			nets:            []claim.CNINetwork{},
			expectedNetsMap: map[string]*claim.CNINetwork{},
		},
		{
			name: "single element in slice",
			nets: []claim.CNINetwork{{Name: "name1"}},
			expectedNetsMap: map[string]*claim.CNINetwork{
				"name1": {Name: "name1"},
			},
		},
		{
			name: "two elements in slice, in order",
			nets: []claim.CNINetwork{{Name: "name1"}, {Name: "name2"}},
			expectedNetsMap: map[string]*claim.CNINetwork{
				"name1": {Name: "name1"},
				"name2": {Name: "name2"},
			},
		},
		{
			name: "two elements in slice, reverse order",
			nets: []claim.CNINetwork{{Name: "name2"}, {Name: "name1"}},
			expectedNetsMap: map[string]*claim.CNINetwork{
				"name1": {Name: "name1"},
				"name2": {Name: "name2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			netsMap := getNetworksMap(tc.nets)
			assert.DeepEqual(t, tc.expectedNetsMap, netsMap)
		})
	}
}

func TestGetMergedListOfNetworksNames(t *testing.T) {
	testCases := []struct {
		name          string
		nets1         []claim.CNINetwork
		nets2         []claim.CNINetwork
		expectedNames []string
	}{
		{
			name:          "empty slices",
			nets1:         []claim.CNINetwork{},
			nets2:         []claim.CNINetwork{},
			expectedNames: []string{},
		},
		{
			name: "only first slice has elements",
			nets1: []claim.CNINetwork{
				{Name: "net1"},
				{Name: "net2"},
			},
			nets2:         []claim.CNINetwork{},
			expectedNames: []string{"net1", "net2"},
		},
		{
			name:  "only second slice has elements",
			nets1: []claim.CNINetwork{},
			nets2: []claim.CNINetwork{
				{Name: "net1"},
				{Name: "net2"},
			},
			expectedNames: []string{"net1", "net2"},
		},
		{
			name: "both have different elements",
			nets1: []claim.CNINetwork{
				{Name: "net3"},
				{Name: "net4"},
			},
			nets2: []claim.CNINetwork{
				{Name: "net1"},
				{Name: "net2"},
			},
			expectedNames: []string{"net1", "net2", "net3", "net4"},
		},
		{
			name: "both have elements but they share some of them",
			nets1: []claim.CNINetwork{
				{Name: "net1"},
				{Name: "net2"},
				{Name: "net3"},
			},
			nets2: []claim.CNINetwork{
				{Name: "net1"},
				{Name: "net2"},
				{Name: "net4"},
			},
			expectedNames: []string{"net1", "net2", "net3", "net4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			names := getMergedListOfNetworksNames(tc.nets1, tc.nets2)
			assert.DeepEqual(t, tc.expectedNames, names)
		})
	}
}

func TestGetPluginsMap(t *testing.T) {
	testCases := []struct {
		name               string
		plugins            []claim.CNIPlugin
		expectedPluginsMap map[string]claim.CNIPlugin
	}{
		{
			name:               "empty slice",
			plugins:            []claim.CNIPlugin{},
			expectedPluginsMap: map[string]claim.CNIPlugin{},
		},
		{
			name:    "single element in slice",
			plugins: []claim.CNIPlugin{{"type": "multus"}},
			expectedPluginsMap: map[string]claim.CNIPlugin{
				"multus": {"type": "multus"},
			},
		},
		{
			name:    "two elements in slice, in order",
			plugins: []claim.CNIPlugin{{"type": "loopback"}, {"type": "multus"}},
			expectedPluginsMap: map[string]claim.CNIPlugin{
				"loopback": {"type": "loopback"},
				"multus":   {"type": "multus"},
			},
		},
		{
			name:    "two elements in slice, reverse order",
			plugins: []claim.CNIPlugin{{"type": "multus"}, {"type": "loopback"}},
			expectedPluginsMap: map[string]claim.CNIPlugin{
				"loopback": {"type": "loopback"},
				"multus":   {"type": "multus"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pluginsMap := getPluginsMap(tc.plugins)
			assert.DeepEqual(t, tc.expectedPluginsMap, pluginsMap)
		})
	}
}

func TestGetMergedListOfPluginsNames(t *testing.T) {
	testCases := []struct {
		name          string
		plugins1      []claim.CNIPlugin
		plugins2      []claim.CNIPlugin
		expectedNames []string
	}{
		{
			name:          "empty slices",
			plugins1:      []claim.CNIPlugin{},
			plugins2:      []claim.CNIPlugin{},
			expectedNames: []string{},
		},
		{
			name: "only first slice has elements",
			plugins1: []claim.CNIPlugin{
				{"type": "multus"},
				{"type": "sriov"},
			},
			plugins2:      []claim.CNIPlugin{},
			expectedNames: []string{"multus", "sriov"},
		},
		{
			name:     "only second slice has elements",
			plugins1: []claim.CNIPlugin{},
			plugins2: []claim.CNIPlugin{
				{"type": "multus"},
				{"type": "sriov"},
			},
			expectedNames: []string{"multus", "sriov"},
		},
		{
			name: "both have different elements",
			plugins1: []claim.CNIPlugin{
				{"type": "multus"},
				{"type": "fakeName1"},
			},
			plugins2: []claim.CNIPlugin{

				{"type": "sriov"},
				{"type": "fakeName2"},
			},
			expectedNames: []string{"fakeName1", "fakeName2", "multus", "sriov"},
		},
		{
			name: "both have elements but they share some of them",
			plugins1: []claim.CNIPlugin{
				{"type": "multus"},
				{"type": "sriov"},
				{"type": "fakeName1"},
			},
			plugins2: []claim.CNIPlugin{
				{"type": "multus"},
				{"type": "sriov"},
				{"type": "fakeName2"},
			},
			expectedNames: []string{"fakeName1", "fakeName2", "multus", "sriov"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			names := getMergedListOfPluginsNames(tc.plugins1, tc.plugins2)
			assert.DeepEqual(t, tc.expectedNames, names)
		})
	}
}

func TestGetMergedListOfPluginFields(t *testing.T) {
	// Helper function to create a CNIPlugin (map of interfaces) from a map of strings.
	createpPluginFieldsMap := func(fields map[string]string) claim.CNIPlugin {
		fieldsMap := map[string]interface{}{}
		for k, v := range fields {
			fieldsMap[k] = v
		}
		return fieldsMap
	}

	testCases := []struct {
		name           string
		claim1Plugin   map[string]string
		claim2Plugin   map[string]string
		expectedFields []string
	}{
		{
			name:           "empty maps",
			claim1Plugin:   map[string]string{},
			claim2Plugin:   map[string]string{},
			expectedFields: []string{},
		},
		{
			name: "fields in first map only",
			claim1Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			claim2Plugin:   map[string]string{},
			expectedFields: []string{"field1", "field2"},
		},
		{
			name:         "fields in second map only",
			claim1Plugin: map[string]string{},
			claim2Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			expectedFields: []string{"field1", "field2"},
		},
		{
			name: "same fields in both maps",
			claim1Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			claim2Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			expectedFields: []string{"field1", "field2"},
		},
		{
			name: "same and different fields",
			claim1Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
				"field3": "value3",
			},
			claim2Plugin: map[string]string{
				"field1": "value1",
				"field2": "value2",
				"field4": "value2",
			},
			expectedFields: []string{"field1", "field2", "field3", "field4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pluginsMap1 := createpPluginFieldsMap(tc.claim1Plugin)
			plugins2Map := createpPluginFieldsMap(tc.claim2Plugin)

			fields := getMergedListOfPluginFields(pluginsMap1, plugins2Map)
			assert.DeepEqual(t, tc.expectedFields, fields)
		})
	}
}

func TestNetworkDiffIsNotFoundIn(t *testing.T) {
	testCases := []struct {
		name               string
		differences        string
		expectedIsNotFound bool
	}{
		{
			name:               "empty differences string",
			differences:        "",
			expectedIsNotFound: false,
		},
		{
			name:               "single random diff",
			differences:        "diff1",
			expectedIsNotFound: false,
		},
		{
			name:               "multiple random diffs",
			differences:        "diff1,diff2",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with wrong claim file number",
			differences:        "not found in claim0",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with wrong format",
			differences:        "not found in claim1,ipam",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with claim file 1",
			differences:        "not found in claim1",
			expectedIsNotFound: true,
		},
		{
			name:               "not found with claim file 2",
			differences:        "not found in claim2",
			expectedIsNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsNotFound, NetworkDiffIsNotFoundIn(tc.differences))
		})
	}
}

func TestCNIPluginsDiffReportsString(t *testing.T) {
	const header = "PLUGIN                        DIFFERENCES\n"
	testCases := []struct {
		name           string
		pluginsDiff    CNIPluginsDiffReports
		expectedString string
	}{
		{
			name:           "empty list",
			pluginsDiff:    []CNIPluginDiffReport{},
			expectedString: header,
		},
		{
			name: "list with one plugin one diff",
			pluginsDiff: []CNIPluginDiffReport{
				{
					PluginName:  "name1",
					Differences: []string{"diff1"},
				},
			},
			expectedString: header + "name1                         diff1\n",
		},
		{
			name: "list with one plugin three diffs",
			pluginsDiff: []CNIPluginDiffReport{
				{
					PluginName:  "name1",
					Differences: []string{"diff1", "diff2", "diff3"},
				},
			},
			expectedString: header +
				"name1                         diff1,diff2,diff3\n",
		},
		{
			name: "list with three plugins, with one, two and three diffs respectively",
			pluginsDiff: []CNIPluginDiffReport{
				{
					PluginName:  "name1",
					Differences: []string{"diff1"},
				},
				{
					PluginName:  "name2",
					Differences: []string{"diff2", "diff3"},
				},
				{
					PluginName:  "name3",
					Differences: []string{"diff4", "diff5", "diff6"},
				},
			},
			expectedString: header +
				"name1                         diff1\n" +
				"name2                         diff2,diff3\n" +
				"name3                         diff4,diff5,diff6\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.pluginsDiff.String()
			assert.Equal(t, tc.expectedString, str)
		})
	}
}

func TestCNINetworksDiffReportsString(t *testing.T) {
	const header = "CNI-NETWORK                   DIFFERENCES\n"

	testCases := []struct {
		name           string
		networksDiff   CNINetworksDiffReports
		expectedString string
	}{
		{
			name:           "empty list",
			networksDiff:   []CNINetworkDiffReport{},
			expectedString: header,
		},
		{
			name: "single network with one diff",
			networksDiff: []CNINetworkDiffReport{
				{
					NetworkName: "net1",
					Differences: []string{"diff1"},
				},
			},
			expectedString: header +
				"net1                          diff1\n",
		},
		{
			name: "three networks with three, two and one diffs respectively",
			networksDiff: []CNINetworkDiffReport{
				{
					NetworkName: "net1",
					Differences: []string{"diff1", "diff2", "diff3"},
				},
				{
					NetworkName: "net2",
					Differences: []string{"diff4", "diff5"},
				},
				{
					NetworkName: "net3",
					Differences: []string{"diff6"},
				},
			},
			expectedString: header +
				"net1                          diff1,diff2,diff3\n" +
				"net2                          diff4,diff5\n" +
				"net3                          diff6\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.networksDiff.String()
			assert.Equal(t, tc.expectedString, str)
		})
	}
}
