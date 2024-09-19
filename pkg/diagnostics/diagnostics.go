// Copyright (C) 2021-2023 Red Hat, Inc.
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

// CniPlugin holds info about a CNI plugin
// The JSON fields come from the jq output

// NodeHwInfo node HW info
type NodeHwInfo struct {
	Lscpu    interface{}
	IPconfig interface{}
	Lsblk    interface{}
	Lspci    []string
}

// GetCniPlugins gets a json representation of the CNI plugins installed in each nodes
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

// GetHwInfoAllNodes gets the Hardware information for each nodes
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

// getHWJsonOutput performs a query via probe pod and returns the JSON blob
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

// getHWTextOutput performs a query via debug and returns plaintext lines
func getHWTextOutput(probePod *corev1.Pod, o clientsholder.Command, cmd string) (out []string, err error) {
	ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
	outStr, errStr, err := o.ExecCommandContainer(ctx, cmd)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %v, stderr: %s", lspciCommand, err, errStr)
	}

	return strings.Split(outStr, "\n"), nil
}

// GetNodeJSON gets the nodes summary in JSON (similar to: oc get nodes -json)
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

// GetCsiDriver Gets the CSI driver list
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

func GetVersionK8s() (out string) {
	env := provider.GetTestEnvironment()
	return env.K8sVersion
}

func GetVersionOcp() (out string) {
	env := provider.GetTestEnvironment()
	if !provider.IsOCPCluster() {
		return "n/a, (non-OpenShift cluster)"
	}
	return env.OpenshiftVersion
}

func GetVersionOcClient() (out string) {
	return "n/a, (not using oc or kubectl client)"
}
