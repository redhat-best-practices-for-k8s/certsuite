package cnis

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
)

const (
	elemNotFoundIn = "not found in "
)

// CNI Network config sections
const (
	differentCNIVersion   = "cniVersion"
	differentPlugins      = "plugins"
	differentDisableCheck = "disable_check"
)

// CNINetworkDiffReport holds the differences report for a cni network.
// This is the structure of a CNI network config:
// - name
// - cniVersion
// - disableCheck
// - plugins
// See https://github.com/containernetworking/cni/blob/main/SPEC.md#section-1-network-configuration-format
// The Differences slice shows which of those fields have a different value
// in the claim files for that network of the same node.
type CNINetworkDiffReport struct {
	NetworkName        string                `json:"networkName"`
	Differences        []string              `json:"differences"`
	PluginsDiffReports CNIPluginsDiffReports `json:"pluginsDiffReport,omitempty"`
}

// CNIPluginDiffReport holds the differences report for a cni network plugin.
// Plugins' config have some known fields (described in the spec) but can also
// have custom fields whose attributes' names and values' types are not known.
// Only the plugin itself knows the meaning of those fields.
// The Differences slice will hold the name of each attribute whoe value is
// different.
type CNIPluginDiffReport struct {
	PluginName  string   `json:"pluginName"`
	Differences []string `json:"differences"`
}

type CNINetworksDiffReports []CNINetworkDiffReport
type CNIPluginsDiffReports []CNIPluginDiffReport

// Stringer method to show in a table the fields where cni networks are different.
// The output looks like this:
//
// CNI-NETWORK                   DIFFERENCES
// crio                          cniVersion,plugins
// loopback						 not found in claim2
//
// That example shows that the crio network is not the same in both claim files
// for a given node. The difference column shows a list of "fields" whose value
// is different. Plugins means that there's at least one crio plugin whose config
// is different.
// In case the network appears only in one claim file, the only difference will
// be "not found in claim[1|2]"
func (networksDiff CNINetworksDiffReports) String() string {
	const diffRowFmt = "%-30s%-s\n"

	str := fmt.Sprintf(diffRowFmt, "CNI-NETWORK", "DIFFERENCES")

	for _, netDiff := range networksDiff {
		differences := ""
		for i := range netDiff.Differences {
			if i != 0 {
				differences += ","
			}
			differences += netDiff.Differences[i]
		}
		str += fmt.Sprintf(diffRowFmt, netDiff.NetworkName, differences)
	}

	return str
}

// Stringer method for CNIPluginsDiffReports. The output table will look
// like this:
//
// PLUGIN                        DIFFERENCES
// bridge                        hairpinMode,ipam
// test							 not found in claim1
//
// For each plugin, the DIFFERENCES column will show the name of the
// attribute whose value won't match in both claim files.
// In case the plugin appears only in one claim file, the only difference will
// be "not found in claim[1|2]"
func (pluginsDiff CNIPluginsDiffReports) String() string {
	const diffRowFmt = "%-30s%-s\n"

	str := fmt.Sprintf(diffRowFmt, "PLUGIN", "DIFFERENCES")
	for _, pluginDiff := range pluginsDiff {
		differences := ""
		for i := range pluginDiff.Differences {
			if i != 0 {
				differences += ","
			}
			differences += pluginDiff.Differences[i]
		}
		str += fmt.Sprintf(diffRowFmt, pluginDiff.PluginName, differences)
	}

	return str
}

// Helper function to parse a string and returns true in case it's
// a "not found in claim[1|2]".
func NetworkDiffIsNotFoundIn(diff string) bool {
	r := regexp.MustCompile("^" + elemNotFoundIn + "claim[1|2]$")
	return r.MatchString(diff)
}

// Helper function to get a map of network names mapped to their corresponding claim.CNINetwork.
func getNetworksMap(networks []claim.CNINetwork) map[string]*claim.CNINetwork {
	networksMap := map[string]*claim.CNINetwork{}

	for i := range networks {
		network := &networks[i]
		networksMap[network.Name] = network
	}

	return networksMap
}

// Helper function to get a merged and sorted list of unique network names from two
// maps of networks by name.
func getMergedListOfNetworksNames(networksClaim1, networksClaim2 map[string]*claim.CNINetwork) []string {
	networkNames := []string{}
	networkNamesMap := map[string]struct{}{}

	for netName := range networksClaim1 {
		networkNamesMap[netName] = struct{}{}
	}

	for netName := range networksClaim2 {
		networkNamesMap[netName] = struct{}{}
	}

	for netName := range networkNamesMap {
		networkNames = append(networkNames, netName)
	}

	sort.Strings(networkNames)
	return networkNames
}

// Helper function to get a map of network names mapped to their corresponding claim.CNINetwork.
// The plugin name is usually the "type" field, as that's the only required field as per
// the CNI network plugins spec:
// https://github.com/containernetworking/cni/blob/main/SPEC.md#plugin-configuration-objects
func getPluginsMap(plugins []claim.CNIPlugin) map[string]claim.CNIPlugin {
	pluginsMap := map[string]claim.CNIPlugin{}

	for _, plugin := range plugins {
		name := plugin["type"].(string)
		pluginsMap[name] = plugin
	}

	return pluginsMap
}

// Helper function to get a merged and sorted list of unique plugins names from two
// maps of claim.CNIPlugin objects.
func getMergedListOfPluginsNames(claim1Plugins, claim2Plugins map[string]claim.CNIPlugin) []string {
	pluginNames := []string{}
	pluginNamesMap := map[string]struct{}{}

	for name := range claim1Plugins {
		pluginNamesMap[name] = struct{}{}
	}

	for name := range claim2Plugins {
		pluginNamesMap[name] = struct{}{}
	}

	for name := range pluginNamesMap {
		pluginNames = append(pluginNames, name)
	}

	sort.Strings(pluginNames)
	return pluginNames
}

// Helper function to get a merged and sorted list of unique plugins fields from two pugins.
func getMergedListOfPluginFields(claim1Plugin, claim2Plugin claim.CNIPlugin) []string {
	fields := []string{}
	fieldsMap := map[string]struct{}{}

	for fieldName := range claim1Plugin {
		fieldsMap[fieldName] = struct{}{}
	}

	for fieldName := range claim2Plugin {
		fieldsMap[fieldName] = struct{}{}
	}

	for name := range fieldsMap {
		fields = append(fields, name)
	}

	sort.Strings(fields)
	return fields
}

// Parses two lists of plugins and returns a diff report for each plugin. In case one plugin
// does not exist in the other slice, it will be marked as "not found in claim[1|2]."
// For each plugin, each config field will be checked to have the same value with reflect.DeepEqual().
// Each field name whose value is different will be added to the slice of Differences for that plugin.
func getCNIPluginsDiffReport(claim1Plugins, claim2Plugins []claim.CNIPlugin) []CNIPluginDiffReport {
	diffReports := []CNIPluginDiffReport{}

	// Helper maps to get plugins by name (type field).
	claim1PluginsMap := getPluginsMap(claim1Plugins)
	claim2PluginsMap := getPluginsMap(claim2Plugins)

	pluginNames := getMergedListOfPluginsNames(claim1PluginsMap, claim2PluginsMap)

	for _, pluginName := range pluginNames {
		report := CNIPluginDiffReport{PluginName: pluginName}

		claim1Plugin, found := claim1PluginsMap[pluginName]
		if !found {
			report.Differences = append(report.Differences, elemNotFoundIn+"claim1")
			diffReports = append(diffReports, report)
			continue
		}

		claim2Plugin, found := claim2PluginsMap[pluginName]
		if !found {
			report.Differences = append(report.Differences, elemNotFoundIn+"claim2")
			diffReports = append(diffReports, report)
			continue
		}

		// The plugin exists in both claim files in the same CNI network for the same node.
		// Now, get a the list of all its fields' names.
		pluginsFields := getMergedListOfPluginFields(claim1Plugin, claim2Plugin)

		// Compare plugin fields and return the name of the field whose value is different.
		for _, fieldName := range pluginsFields {
			claim1Value, foundInClaim1 := claim1Plugin[fieldName]
			claim2Value, foundInClaim2 := claim2Plugin[fieldName]

			if !foundInClaim1 || !foundInClaim2 || !reflect.DeepEqual(claim1Value, claim2Value) {
				report.Differences = append(report.Differences, fieldName)
			}
		}

		if len(report.Differences) > 0 {
			diffReports = append(diffReports, report)
		}
	}

	return diffReports
}

// Generates a CNINetworkDiffReport from two slices of claim.CNINetwork objects.
func GetDiffReports(networksClaim1, networksClaim2 []claim.CNINetwork) []CNINetworkDiffReport {
	diffReports := []CNINetworkDiffReport{}

	// Helper maps to get CNI networks by network name.
	netsClaim1Map := getNetworksMap(networksClaim1)
	netsClaim2Map := getNetworksMap(networksClaim2)

	netNames := getMergedListOfNetworksNames(netsClaim1Map, netsClaim2Map)

	for _, netName := range netNames {
		report := CNINetworkDiffReport{NetworkName: netName}

		network1, found := netsClaim1Map[netName]
		if !found {
			report.Differences = append(report.Differences, elemNotFoundIn+"claim1")
			diffReports = append(diffReports, report)
			continue
		}

		network2, found := netsClaim2Map[netName]
		if !found {
			report.Differences = append(report.Differences, elemNotFoundIn+"claim2")
			diffReports = append(diffReports, report)
			continue
		}

		if network1.CNIVersion != network2.CNIVersion {
			report.Differences = append(report.Differences, differentCNIVersion)
		}

		if network1.DisableCheck != network2.DisableCheck {
			report.Differences = append(report.Differences, differentDisableCheck)
		}

		pluginsDiffReports := getCNIPluginsDiffReport(network1.Plugins, network2.Plugins)
		if len(pluginsDiffReports) > 0 {
			report.Differences = append(report.Differences, differentPlugins)
			report.PluginsDiffReports = pluginsDiffReports
		}

		if len(report.Differences) > 0 {
			diffReports = append(diffReports, report)
		}
	}

	return diffReports
}
