package cnis

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

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

type CNINetworkDiffReport struct {
	NetworkName        string                `json:"networkName"`
	Differences        []string              `json:"differences"`
	PluginsDiffReports CNIPluginsDiffReports `json:"pluginsDiffReport,omitempty"`
}

type CNIPluginDiffReport struct {
	PluginName  string   `json:"pluginName"`
	Differences []string `json:"differences"`
}

type CNINetworksDiffReports []CNINetworkDiffReport
type CNIPluginsDiffReports []CNIPluginDiffReport

func (networksDiff CNINetworksDiffReports) String() string {
	const diffRowFmt = "%-30s%-60s\n"

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

func (pluginsDiff CNIPluginsDiffReports) String() string {
	const diffRowFmt = "%-30s%-60s\n"

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

func NetworkDiffIsNotFoundIn(diff string) bool {
	return strings.Contains(diff, elemNotFoundIn)
}

func getMergedListOfNetworksNames(networksClaim1, networksClaim2 []claim.CNINetwork) []string {
	networkNames := []string{}
	networkNamesMap := map[string]struct{}{}

	for _, network := range networksClaim1 {
		networkNamesMap[network.Name] = struct{}{}
	}

	for _, network := range networksClaim2 {
		networkNamesMap[network.Name] = struct{}{}
	}

	for name := range networkNamesMap {
		networkNames = append(networkNames, name)
	}

	sort.Strings(networkNames)
	return networkNames
}

func getNetworksMap(networks []claim.CNINetwork) map[string]claim.CNINetwork {
	networksMap := map[string]claim.CNINetwork{}

	for _, network := range networks {
		networksMap[network.Name] = network
	}

	return networksMap
}

func getMergedListOfPluginsNames(claim1Plugins, claim2Plugins []claim.CNIPlugin) []string {
	pluginNames := []string{}
	pluginNamesMap := map[string]struct{}{}

	for _, plugin := range claim1Plugins {
		name := plugin["type"].(string)
		pluginNamesMap[name] = struct{}{}
	}

	for _, plugin := range claim2Plugins {
		name := plugin["type"].(string)
		pluginNamesMap[name] = struct{}{}
	}

	for name := range pluginNamesMap {
		pluginNames = append(pluginNames, name)
	}

	sort.Strings(pluginNames)
	return pluginNames
}

func getPluginsMap(plugins []claim.CNIPlugin) map[string]claim.CNIPlugin {
	pluginsMap := map[string]claim.CNIPlugin{}

	for _, plugin := range plugins {
		name := plugin["type"].(string)
		pluginsMap[name] = plugin
	}

	return pluginsMap
}

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

func getCNIPluginsDiffReport(claim1Plugins, claim2Plugins []claim.CNIPlugin) []CNIPluginDiffReport {
	diffReports := []CNIPluginDiffReport{}

	claim1PluginsMap := getPluginsMap(claim1Plugins)
	claim2PluginsMap := getPluginsMap(claim2Plugins)

	pluginNames := getMergedListOfPluginsNames(claim1Plugins, claim2Plugins)

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

		pluginsFields := getMergedListOfPluginFields(claim1Plugin, claim2Plugin)
		// Compare plugin fields and return the name of the field whose value is different.
		for _, fieldName := range pluginsFields {
			claim1Value, foundInClaim1 := claim1Plugin[fieldName]
			claim2Value, foundInClaim2 := claim2Plugin[fieldName]

			if !foundInClaim1 || !foundInClaim2 {
				report.Differences = append(report.Differences, fieldName)
			} else if !reflect.DeepEqual(claim1Value, claim2Value) {
				report.Differences = append(report.Differences, fieldName)
			}
		}

		if len(report.Differences) > 0 {
			diffReports = append(diffReports, report)
		}
	}

	return diffReports
}

func GetDiffReports(networksClaim1, networksClaim2 []claim.CNINetwork) []CNINetworkDiffReport {
	diffReports := []CNINetworkDiffReport{}

	netsClaim1Map := getNetworksMap(networksClaim1)
	netsClaim2Map := getNetworksMap(networksClaim2)

	netNames := getMergedListOfNetworksNames(networksClaim1, networksClaim2)

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
