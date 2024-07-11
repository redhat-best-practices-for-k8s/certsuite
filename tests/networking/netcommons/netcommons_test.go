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

package netcommons

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	corev1 "k8s.io/api/core/v1"
)

func Test_getIPVersion(t *testing.T) {
	type args struct {
		aIP string
	}
	tests := []struct {
		name    string
		args    args
		want    IPVersion
		wantErr bool
	}{
		{name: "GoodIPv4",
			args:    args{aIP: "2.2.2.2"},
			want:    IPv4,
			wantErr: false,
		},
		{name: "GoodIPv6",
			args:    args{aIP: "fd00:10:244:1::3"},
			want:    IPv6,
			wantErr: false,
		},
		{name: "BadIPv4",
			args:    args{aIP: "2.hfh.2.2"},
			want:    Undefined,
			wantErr: true,
		},
		{name: "BadIPv6",
			args:    args{aIP: "fd00:10:ono;ogmo:1::3"},
			want:    Undefined,
			wantErr: true,
		},
		{
			name:    "EmptyString",
			args:    args{aIP: ""},
			want:    Undefined,
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetIPVersion(tt.args.aIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIPVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIPVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIPListByIPVersion(t *testing.T) {
	type args struct {
		ipList     []string
		aIPVersion IPVersion
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "okIpv4",
			args: args{ipList: []string{"1.1.1.1", "2.2.2.2", "fd00:10:244:1::3"}, aIPVersion: IPv4},
			want: []string{"1.1.1.1", "2.2.2.2"},
		},
		{
			name: "okIpv6",
			args: args{ipList: []string{"1.1.1.1", "2.2.2.2", "fd00:10:244:1::3"}, aIPVersion: IPv6},
			want: []string{"fd00:10:244:1::3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterIPListByIPVersion(tt.args.ipList, tt.args.aIPVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterIPListByIPVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodIPsToStringList(t *testing.T) {
	testCases := []struct {
		testPodIPs   []corev1.PodIP
		expectedList []string
	}{
		{ // Test Case #1 - different IPs
			testPodIPs: []corev1.PodIP{
				{
					IP: "192.168.1.1",
				},
				{
					IP: "192.168.1.2",
				},
			},
			expectedList: []string{"192.168.1.1", "192.168.1.2"},
		},
		{ // Test Case #2 - same IPs
			testPodIPs: []corev1.PodIP{
				{
					IP: "192.168.1.1",
				},
				{
					IP: "192.168.1.1",
				},
			},
			expectedList: []string{"192.168.1.1", "192.168.1.1"},
		},
		{ // Test Case #3 - no IPs
			testPodIPs:   []corev1.PodIP{},
			expectedList: nil,
		},
	}

	for _, tc := range testCases {
		results := PodIPsToStringList(tc.testPodIPs)
		sort.Strings(results)
		assert.Equal(t, tc.expectedList, results)
	}
}

func TestContainerIPString(t *testing.T) {
	generateCIP := func(IP, ID string) *ContainerIP {
		return &ContainerIP{
			IP: IP,
			ContainerIdentifier: &provider.Container{
				UID: ID,
				Container: &corev1.Container{
					Name: "test-" + ID,
				},
			},
		}
	}

	testCases := []struct {
		testID         string
		testIP         string
		expectedString string
	}{
		{
			testID:         "ID1",
			testIP:         "192.168.1.1",
			expectedString: "192.168.1.1 ( node:  ns:  podName:  containerName: test-ID1 containerUID:  containerRuntime:  )",
		},
		{
			testID:         "ID2",
			testIP:         "192.168.1.2",
			expectedString: "192.168.1.2 ( node:  ns:  podName:  containerName: test-ID2 containerUID:  containerRuntime:  )",
		},
	}

	for _, tc := range testCases {
		cip := generateCIP(tc.testIP, tc.testID)
		assert.Equal(t, tc.expectedString, cip.String())
	}
}

func TestNetTestContextString(t *testing.T) {
	buildTestContext := func(IP, containerName string) NetTestContext {
		return NetTestContext{
			TesterContainerNodeName: "node1",
			TesterSource: ContainerIP{
				IP: IP,
				ContainerIdentifier: &provider.Container{
					Container: &corev1.Container{
						Name: containerName,
					},
				},
			},
			DestTargets: []ContainerIP{
				{
					IP: IP,
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: containerName,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testIP            string
		testContainerName string
		expectedString    string
	}{
		{
			testIP:            "192.168.1.1",
			testContainerName: "container1",
			expectedString:    "From initiating container: 192.168.1.1 ( node:  ns:  podName:  containerName: container1 containerUID:  containerRuntime:  )\n--> To target container: 192.168.1.1 ( node:  ns:  podName:  containerName: container1 containerUID:  containerRuntime:  )\n", //nolint:lll
		},
		{
			testIP:            "192.168.1.2",
			testContainerName: "container2",
			expectedString:    "From initiating container: 192.168.1.2 ( node:  ns:  podName:  containerName: container2 containerUID:  containerRuntime:  )\n--> To target container: 192.168.1.2 ( node:  ns:  podName:  containerName: container2 containerUID:  containerRuntime:  )\n", //nolint:lll
		},
	}

	for _, tc := range testCases {
		ntc := buildTestContext(tc.testIP, tc.testContainerName)
		assert.Equal(t, tc.expectedString, ntc.String())
	}
}

func TestNetTestContextMapString(t *testing.T) {
	buildTestContext := func(IP, containerName string) NetTestContext {
		return NetTestContext{
			TesterContainerNodeName: "node1",
			TesterSource: ContainerIP{
				IP: IP,
				ContainerIdentifier: &provider.Container{
					Container: &corev1.Container{
						Name: containerName,
					},
				},
			},
			DestTargets: []ContainerIP{
				{
					IP: IP,
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: containerName,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testIndex         string
		testIP            string
		testContainerName string
		expectedString    string
	}{
		{
			testIndex:         "test1",
			testIP:            "192.168.1.1",
			testContainerName: "container1",
			expectedString:    "***Test for Network attachment: test1\nFrom initiating container: 192.168.1.1 ( node:  ns:  podName:  containerName: container1 containerUID:  containerRuntime:  )\n--> To target container: 192.168.1.1 ( node:  ns:  podName:  containerName: container1 containerUID:  containerRuntime:  )\n\n", //nolint:lll
		},
		{
			testIndex:         "test2",
			testIP:            "192.168.1.2",
			testContainerName: "container2",
			expectedString:    "***Test for Network attachment: test2\nFrom initiating container: 192.168.1.2 ( node:  ns:  podName:  containerName: container2 containerUID:  containerRuntime:  )\n--> To target container: 192.168.1.2 ( node:  ns:  podName:  containerName: container2 containerUID:  containerRuntime:  )\n\n", //nolint:lll
		},
	}

	for _, tc := range testCases {
		m := make(map[string]NetTestContext)
		m[tc.testIndex] = buildTestContext(tc.testIP, tc.testContainerName)
		assert.Equal(t, tc.expectedString, PrintNetTestContextMap(m))
	}
}
