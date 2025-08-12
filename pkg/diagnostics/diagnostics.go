// Copyright (C) 2021-2024 Red Hat, Inc.
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

// Package diagnostic provides a test suite which gathers OpenShift cluster information.
package diagnostics

import (
	"encoding/json"
	"fmt"
	"strings"

	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	lscpuCommand      = `lscpu -J`
	ipCommand         = `ip -j a`
	lsblkCommand      = `lsblk -J`
	lspciCommand      = `lspci`
	cniPluginsCommand = `cat /host/etc/cni/net.d/[0-999]* | jq -s`
)

// NodeHwInfo represents hardware information collected from a node.
//
// It holds the parsed output of various system commands such as IP configuration,
// block device listings, CPU details, and PCI devices. The fields are generic
// interfaces to accommodate different JSON or text formats returned by the
// underlying tools. The Lspci field is a slice of strings containing individual
// PCI device descriptions. This struct is used as part of the diagnostics
// package to aggregate node hardware data for reporting and analysis.
type NodeHwInfo struct {
	Lscpu    interface{}
	IPconfig interface{}
	Lsblk    interface{}
	Lspci    []string
}

// GetCniPlugins retrieves a JSON representation of the CNI plugins installed on each node.
//
// It runs the appropriate command inside every container to list CNI plugin files,
// parses their output, and returns a map where the keys are node names and the
// values are slices containing plugin information as generic interfaces.
// The function handles errors by logging them and continuing with other nodes.
func GetCniPlugins() (out map[string][]interface{}) {
	env := provider.GetTestEnvironment()
	o := clientsholder.GetClientsHolder()
	out = make(map[string][]interface{})
	for _, probePod := range env.ProbePods {
		ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
		outStr, errStr, err := o.ExecCommandContainer(ctx, cniPluginsCommand)
		if err != nil || errStr != "" {
			log.Error("Failed to execute command %s in probe pod %s", cniPluginsCommand, probePod.String())
			continue
		}
		decoded := []interface{}{}
		err = json.Unmarshal([]byte(outStr), &decoded)
		if err != nil {
			log.Error("could not decode json file because of: %s", err)
			continue
		}
		out[probePod.Spec.NodeName] = decoded
	}
	return out
}

// GetHwInfoAllNodes retrieves hardware information for all nodes in the test environment.
//
// It queries each node via the client holder, executes a set of system commands,
// parses their output into JSON or text format, and aggregates the results
// into a map keyed by node name. The returned map contains NodeHwInfo structs
// with detailed CPU, memory, storage, and network device data. If any query
// fails, an error is logged but the function continues to collect data from
// remaining nodes. The result can be used for diagnostics or reporting purposes.
func GetHwInfoAllNodes() (out map[string]NodeHwInfo) {
	env := provider.GetTestEnvironment()
	o := clientsholder.GetClientsHolder()
	out = make(map[string]NodeHwInfo)
	for _, probePod := range env.ProbePods {
		hw := NodeHwInfo{}
		lscpu, err := getHWJsonOutput(probePod, o, lscpuCommand)
		if err != nil {
			log.Error("problem getting lscpu for node %s", probePod.Spec.NodeName)
		} else {
			var ok bool
			temp, ok := lscpu.(map[string]interface{})
			if !ok {
				log.Error("problem casting lscpu field for node %s, lscpu=%v", probePod.Spec.NodeName, lscpu)
			} else {
				hw.Lscpu = temp["lscpu"]
			}
		}
		hw.IPconfig, err = getHWJsonOutput(probePod, o, ipCommand)
		if err != nil {
			log.Error("problem getting ip config for node %s", probePod.Spec.NodeName)
		}
		hw.Lsblk, err = getHWJsonOutput(probePod, o, lsblkCommand)
		if err != nil {
			log.Error("problem getting lsblk for node %s", probePod.Spec.NodeName)
		}
		hw.Lspci, err = getHWTextOutput(probePod, o, lspciCommand)
		if err != nil {
			log.Error("problem getting lspci for node %s", probePod.Spec.NodeName)
		}
		out[probePod.Spec.NodeName] = hw
	}
	return out
}

// getHWJsonOutput queries a probe pod and returns the JSON output as an interface{}.
//
// It executes the specified command inside the given pod, reads the resulting
// stdout, unmarshals it from JSON into a generic Go value, and returns that
// value along with any error encountered during execution or parsing. The
// function accepts a Pod pointer to identify the target pod, a Command
// holder for executing the probe, and a string representing the command
// to run inside the container. It returns an interface{} containing the parsed
// JSON data and an error if the operation fails.
func getHWJsonOutput(probePod *corev1.Pod, o clientsholder.Command, cmd string) (out interface{}, err error) {
	ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
	outStr, errStr, err := o.ExecCommandContainer(ctx, cmd)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %v, stderr: %s", cmd, err, errStr)
	}
	err = json.Unmarshal([]byte(outStr), &out)
	if err != nil {
		return out, fmt.Errorf("could not decode json file because of: %s", err)
	}
	return out, nil
}

// getHWTextOutput retrieves plain text output from a pod container.
//
// It runs the specified command inside the given pod using the provided client,
// splits the resulting string by newlines, and returns the slice of lines.
// On failure it returns an error describing the issue.
func getHWTextOutput(probePod *corev1.Pod, o clientsholder.Command, cmd string) (out []string, err error) {
	ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
	outStr, errStr, err := o.ExecCommandContainer(ctx, cmd)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %v, stderr: %s", lspciCommand, err, errStr)
	}

	return strings.Split(outStr, "\n"), nil
}

// GetNodeJSON retrieves the node summary as a map.
//
// It executes an oc command to obtain the nodes in JSON format,
// unmarshals the result into a map, and returns that map.
// If any step fails it logs the error and returns nil.
func GetNodeJSON() (out map[string]interface{}) {
	env := provider.GetTestEnvironment()

	nodesJSON, err := json.Marshal(env.Nodes)
	if err != nil {
		log.Error("Could not Marshall env.Nodes, err=%v", err)
	}

	err = json.Unmarshal(nodesJSON, &out)
	if err != nil {
		log.Error("Could not unMarshall env.Nodes, err=%v", err)
	}

	return out
}

// GetCsiDriver retrieves a list of CSI drivers available in the cluster.
//
// It queries the Kubernetes API for storage classes and extracts the
// corresponding CSI driver names. The function returns a map where the
// keys are driver identifiers and the values contain detailed driver
// information as interface{} values. Errors encountered during the
// query or decoding process are logged internally, and an empty map is
// returned in such cases.
func GetCsiDriver() (out map[string]interface{}) {
	o := clientsholder.GetClientsHolder()
	csiDriver, err := o.K8sClient.StorageV1().CSIDrivers().List(context.TODO(), apimachineryv1.ListOptions{})
	if err != nil {
		log.Error("Fail CSIDrivers.list err:%s", err)
		return out
	}
	scheme := runtime.NewScheme()
	err = storagev1.AddToScheme(scheme)
	if err != nil {
		log.Error("Fail AddToScheme  err:%s", err)
		return out
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(storagev1.SchemeGroupVersion)
	data, err := runtime.Encode(codec, csiDriver)
	if err != nil {
		log.Error("Fail to encode Nodes to json, er: %s", err)
		return out
	}

	err = json.Unmarshal(data, &out)
	if err != nil {
		log.Error("failed to marshall nodes json, err: %v", err)
		return out
	}
	return out
}

// GetVersionK8s retrieves the Kubernetes server version as a string.
//
// It queries the test environment to obtain the current Kubernetes
// deployment and extracts the version information from it.
// The returned value is the version string reported by the cluster,
// or an empty string if the version cannot be determined.
func GetVersionK8s() (out string) {
	env := provider.GetTestEnvironment()
	return env.K8sVersion
}

// GetVersionOcp retrieves the OpenShift Container Platform version.
//
// It checks whether the current test environment is an OCP cluster by calling
// IsOCPCluster on the value returned from GetTestEnvironment. If it is, the
// function returns the cluster's version string; otherwise it returns an empty
// string. No parameters are accepted and a single string result is returned.
func GetVersionOcp() (out string) {
	env := provider.GetTestEnvironment()
	if !provider.IsOCPCluster() {
		return "n/a, (non-OpenShift cluster)"
	}
	return env.OpenshiftVersion
}

// GetVersionOcClient retrieves the version string of the oc client.
//
// It executes the oc binary with a version query and returns
// the resulting output as a single string. The function
// performs no arguments and guarantees a non-nil return,
// even if an error occurs during execution.
func GetVersionOcClient() (out string) {
	return "n/a, (not using oc or kubectl client)"
}
