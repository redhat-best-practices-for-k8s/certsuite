// Copyright (C) 2021 Red Hat, Inc.
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

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	lscpuCommand = `lscpu -J`
	ipCommand    = `ip -j a`
	lsblkCommand = `lsblk -J`
	lspciCommand = `lspci`
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
	for _, debugPod := range env.DebugPods {
		ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
		outStr, errStr, err := o.ExecCommandContainer(ctx, `cat /host/etc/cni/net.d/[0-999]* | jq -s '[ .[] | {name:.name, type:.type, version:.cniVersion, plugins: .plugins}]'`)
		if err != nil || errStr != "" {
			continue
		}
		decoded := []interface{}{}
		err = json.Unmarshal([]byte(outStr), &decoded)
		if err != nil {
			logrus.Errorf("could not decode json file because of: %s", err)
			continue
		}
		out[debugPod.Spec.NodeName] = decoded
	}
	return out
}

// GetHwInfoAllNodes gets the Hardware information for each nodes
func GetHwInfoAllNodes() (out map[string]NodeHwInfo) {
	env := provider.GetTestEnvironment()
	o := clientsholder.GetClientsHolder()
	out = make(map[string]NodeHwInfo)
	var err error
	for _, debugPod := range env.DebugPods {
		hw := NodeHwInfo{}
		hw.Lscpu, err = getLscpu(debugPod, o)
		if err != nil {
			logrus.Errorf("problem getting lscpu for node %s", debugPod.Spec.NodeName)
		}
		hw.IPconfig, err = getIPConfig(debugPod, o)
		if err != nil {
			logrus.Errorf("problem getting ip config for node %s", debugPod.Spec.NodeName)
		}
		hw.Lsblk, err = getLsblk(debugPod, o)
		if err != nil {
			logrus.Errorf("problem getting lsblk for node %s", debugPod.Spec.NodeName)
		}
		hw.Lspci, err = getLspci(debugPod, o)
		if err != nil {
			logrus.Errorf("problem getting lspci for node %s", debugPod.Spec.NodeName)
		}
		out[debugPod.Spec.NodeName] = hw
	}
	return out
}

// getLscpu gets lscpu in Json format for a given node
func getLscpu(debugPod *v1.Pod, o *clientsholder.ClientsHolder) (out interface{}, err error) { //nolint:dupl
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
	outStr, errStr, err := o.ExecCommandContainer(ctx, lscpuCommand)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %s , stderr: %s", lscpuCommand, err, errStr)
	}
	err = json.Unmarshal([]byte(outStr), &out)
	if err != nil {
		return out, fmt.Errorf("could not decode json file because of: %s", err)
	}
	return out, nil
}

// getIPConfig gets ipconfig -a in Json format for a given node
func getIPConfig(debugPod *v1.Pod, o *clientsholder.ClientsHolder) (out interface{}, err error) { //nolint:dupl
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
	outStr, errStr, err := o.ExecCommandContainer(ctx, ipCommand)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %s , stderr: %s", ipCommand, err, errStr)
	}
	err = json.Unmarshal([]byte(outStr), &out)
	if err != nil {
		return out, fmt.Errorf("could not decode json file because of: %s", err)
	}
	return out, nil
}

// getLsblk gets lsblk in Json format for a given node
func getLsblk(debugPod *v1.Pod, o *clientsholder.ClientsHolder) (out interface{}, err error) { //nolint:dupl
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
	outStr, errStr, err := o.ExecCommandContainer(ctx, lsblkCommand)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %s , stderr: %s", lsblkCommand, err, errStr)
	}
	err = json.Unmarshal([]byte(outStr), &out)
	if err != nil {
		return out, fmt.Errorf("could not decode json file because of: %s", err)
	}
	return out, nil
}

// getLspci gets lspci in Json format for a given node
func getLspci(debugPod *v1.Pod, o *clientsholder.ClientsHolder) (out []string, err error) {
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
	outStr, errStr, err := o.ExecCommandContainer(ctx, lspciCommand)
	if err != nil || errStr != "" {
		return out, fmt.Errorf("command %s failed with error err: %s , stderr: %s", lspciCommand, err, errStr)
	}

	return strings.Split(outStr, "\n"), nil
}

// GetNodeJSON gets the nodes summary in JSON (similar to: oc get nodes -json)
func GetNodeJSON() (out map[string]interface{}) {
	env := provider.GetTestEnvironment()

	scheme := runtime.NewScheme()
	err := v1.AddToScheme(scheme)
	if err != nil {
		logrus.Errorf("Fail GetNodeJSON err:%s", err)
		return out
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(v1.SchemeGroupVersion)
	data, err := runtime.Encode(codec, env.Nodes)
	if err != nil {
		logrus.Errorf("Fail to encode Nodes to json, er: %s", err)
		return out
	}

	err = json.Unmarshal(data, &out)
	if err != nil {
		logrus.Errorf("failed to marshall nodes json, err: %s", err)
		return out
	}
	return out
}

// GetCsiDriver Gets the CSI driver list
func GetCsiDriver() (out map[string]interface{}) {
	o := clientsholder.GetClientsHolder()
	csiDriver, err := o.StorageClient.CSIDrivers().List(context.TODO(), apimachineryv1.ListOptions{})
	if err != nil {
		logrus.Errorf("Fail CSIDrivers.list err:%s", err)
		return out
	}
	scheme := runtime.NewScheme()
	err = storagev1.AddToScheme(scheme)
	if err != nil {
		logrus.Errorf("Fail AddToScheme  err:%s", err)
		return out
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(storagev1.SchemeGroupVersion)
	data, err := runtime.Encode(codec, csiDriver)
	if err != nil {
		logrus.Errorf("Fail to encode Nodes to json, er: %s", err)
		return out
	}

	err = json.Unmarshal(data, &out)
	if err != nil {
		logrus.Errorf("failed to marshall nodes json, err: %s", err)
		return out
	}
	return out
}
