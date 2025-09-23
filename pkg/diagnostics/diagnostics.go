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

// NodeHwInfo Container for node hardware details
//
// This structure stores parsed output from various system utilities, including
// CPU information, IP configuration, block device layout, and PCI devices. Each
// field holds the raw or processed data returned by the diagnostics functions.
// The struct is populated per-node and used to aggregate hardware profiles
// across a cluster.
type NodeHwInfo struct {
	Lscpu    interface{}
	IPconfig interface{}
	Lsblk    interface{}
	Lspci    []string
}

// GetCniPlugins Retrieves CNI plugin information from probe pods
//
// This function gathers the JSON output of a command run inside each probe pod
// to collect installed CNI plugins for every node. It executes the command,
// parses the returned JSON into generic interface slices, and maps them by node
// name. Errors during execution or decoding are logged and that node is
// skipped.
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

// GetHwInfoAllNodes Collects hardware details from all probe pods
//
// This function iterates over each probe pod defined in the test environment,
// executing a series of commands to gather CPU, memory, network, block device,
// and PCI information. The results are parsed into a structured map keyed by
// node name, with errors logged but not stopping the collection for other
// nodes. It returns a map where each entry contains a NodeHwInfo struct holding
// the gathered data.
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

// getHWJsonOutput Executes a command in a pod and decodes its JSON output
//
// This function runs the supplied shell command inside a specified container of
// a pod, captures the standard output, and unmarshals it into an interface. If
// the command fails or returns non‑empty stderr, an error is returned.
// Successful execution yields the parsed JSON data.
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

// getHWTextOutput Runs a command in a pod container and returns its output lines
//
// The function constructs a context for the specified pod and container, then
// executes the given command using the client holder. If the command fails or
// produces error output, it returns an error describing the failure. On
// success, it splits the standard output by newline characters and returns the
// resulting slice of strings.
func getHWTextOutput(probePod *corev1.Pod, o clientsholder.Command, cmd string) (out []string, err error) {
	ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
	outStr, errStr, err := o.ExecCommandContainer(ctx, cmd)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %v, stderr: %s", lspciCommand, err, errStr)
	}

	return strings.Split(outStr, "\n"), nil
}

// GetNodeJSON Retrieves a JSON representation of node information
//
// The function obtains the test environment, marshals its Nodes field into
// JSON, then unmarshals that data back into a generic map structure for use
// elsewhere. It logs errors if either marshaling or unmarshaling fails and
// returns the resulting map.
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

// GetCsiDriver Retrieves a list of CSI drivers from the Kubernetes cluster
//
// This function accesses the Kubernetes client holder to query the StorageV1
// API for all CSI drivers, encodes the result into JSON, and then unmarshals it
// into a map. Errors during listing, scheme setup, encoding, or decoding are
// logged and cause an empty map to be returned. The resulting map contains
// driver details suitable for inclusion in diagnostic reports.
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

// GetVersionK8s Returns the Kubernetes version used in the test environment
//
// This function obtains the current test environment configuration and extracts
// the Kubernetes version string. It accesses the global environment state via
// provider.and returns the K8sVersion field. The result is a plain string
// representing the cluster's Kubernetes release.
func GetVersionK8s() (out string) {
	env := provider.GetTestEnvironment()
	return env.K8sVersion
}

// GetVersionOcp Retrieves the OpenShift version of the current environment
//
// This function first obtains test environment data, then checks whether the
// cluster is an OpenShift instance. If it is not, a placeholder string
// indicating a non‑OpenShift cluster is returned; otherwise the stored
// OpenshiftVersion value is provided as output.
func GetVersionOcp() (out string) {
	env := provider.GetTestEnvironment()
	if !provider.IsOCPCluster() {
		return "n/a, (non-OpenShift cluster)"
	}
	return env.OpenshiftVersion
}

// GetVersionOcClient Returns a placeholder indicating oc client is not used
//
// The function simply provides the text "n/a, " to signal that no OpenShift
// client version information is available in this context.
func GetVersionOcClient() (out string) {
	return "n/a, (not using oc or kubectl client)"
}
