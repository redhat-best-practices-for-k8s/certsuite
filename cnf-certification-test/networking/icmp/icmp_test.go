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

package icmp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_parsePingResult(t *testing.T) {
	type args struct {
		stdout string
		stderr string
	}
	tests := []struct {
		name        string
		args        args
		wantResults PingResults
		wantErr     bool
	}{
		{
			name: "pingOk",
			args: args{stdout: `PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
		 64 bytes from 8.8.8.8: icmp_seq=1 ttl=57 time=36.1 ms
		 64 bytes from 8.8.8.8: icmp_seq=2 ttl=57 time=32.6 ms
		 64 bytes from 8.8.8.8: icmp_seq=3 ttl=57 time=35.9 ms
		 64 bytes from 8.8.8.8: icmp_seq=4 ttl=57 time=38.2 ms
		 64 bytes from 8.8.8.8: icmp_seq=5 ttl=57 time=36.0 ms
		 
		 --- 8.8.8.8 ping statistics ---
		 5 packets transmitted, 5 received, 0% packet loss, time 4005ms
		 rtt min/avg/max/mdev = 32.593/35.761/38.212/1.802 ms`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.SUCCESS,
				transmitted: 5,
				received:    5,
				errors:      0,
			},
			wantErr: false,
		},
		{
			name: "pingErrorPacket",
			args: args{stdout: `PING 192.168.1.1 (192.168.1.1) 56(84) bytes of data.
			64 bytes from 192.168.1.1: icmp_seq=1 ttl=61 time=1.79 ms
			64 bytes from 192.168.1.1: icmp_seq=2 ttl=61 time=3.37 ms
			64 bytes from 192.168.1.1: icmp_seq=3 ttl=61 time=2.14 ms
			64 bytes from 192.168.1.1: icmp_seq=4 ttl=61 time=3.62 ms
			From 10.0.2.2 icmp_seq=5 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=6 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=7 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=8 Destination Net Unreachable
			64 bytes from 192.168.1.1: icmp_seq=9 ttl=61 time=297 ms
			64 bytes from 192.168.1.1: icmp_seq=10 ttl=61 time=258 ms
			64 bytes from 192.168.1.1: icmp_seq=11 ttl=61 time=276 ms
			64 bytes from 192.168.1.1: icmp_seq=12 ttl=61 time=1.58 ms
			64 bytes from 192.168.1.1: icmp_seq=13 ttl=61 time=445 ms
			64 bytes from 192.168.1.1: icmp_seq=14 ttl=61 time=3.57 ms
			64 bytes from 192.168.1.1: icmp_seq=15 ttl=61 time=60.5 ms
			64 bytes from 192.168.1.1: icmp_seq=16 ttl=61 time=585 ms
			64 bytes from 192.168.1.1: icmp_seq=17 ttl=61 time=155 ms
			64 bytes from 192.168.1.1: icmp_seq=18 ttl=61 time=20.4 ms
			64 bytes from 192.168.1.1: icmp_seq=19 ttl=61 time=26.7 ms
			64 bytes from 192.168.1.1: icmp_seq=20 ttl=61 time=2.14 ms
			
			--- 192.168.1.1 ping statistics ---
			20 packets transmitted, 16 received, +4 errors, 20% packet loss, time 19118ms
			rtt min/avg/max/mdev = 1.582/134.079/585.861/179.394 ms`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.ERROR,
				transmitted: 20,
				received:    16,
				errors:      4,
			},
			wantErr: false,
		},
		{
			name: "pingIncorrectIp",
			args: args{stdout: `connect: Invalid argument
			command terminated with exit code 2`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.ERROR,
				transmitted: 0,
				received:    0,
				errors:      0,
			},
			wantErr: true,
		},
		{
			name: "pingPassingPacketLoss",
			args: args{stdout: `PING 192.168.1.5 (192.168.1.5) 56(84) bytes of data.
			64 bytes from 192.168.1.5: icmp_seq=1 ttl=61 time=14.8 ms
			64 bytes from 192.168.1.5: icmp_seq=2 ttl=61 time=11.2 ms
			64 bytes from 192.168.1.5: icmp_seq=3 ttl=61 time=10.9 ms
			64 bytes from 192.168.1.5: icmp_seq=5 ttl=61 time=9.68 ms
			64 bytes from 192.168.1.5: icmp_seq=6 ttl=61 time=4.55 ms
			64 bytes from 192.168.1.5: icmp_seq=7 ttl=61 time=3.38 ms
			64 bytes from 192.168.1.5: icmp_seq=8 ttl=61 time=3.67 ms
			64 bytes from 192.168.1.5: icmp_seq=9 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=10 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=11 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=12 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=13 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=14 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=15 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=16 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=17 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=18 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=19 ttl=61 time=3.77 ms
			
			--- 192.168.1.5 ping statistics ---
			20 packets transmitted, 19 received, 5% packet loss, time 19297ms
			rtt min/avg/max/mdev = 3.381/7.772/14.867/4.167 ms`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.SUCCESS,
				transmitted: 20,
				received:    19,
				errors:      0,
			},
			wantErr: false,
		},
		{
			name: "pingFailingPacketLoss",
			args: args{stdout: `
			PING 192.168.1.2 (192.168.1.2) 56(84) bytes of data.
			
			--- 192.168.1.2 ping statistics ---
			1 packets transmitted, 0 received, 100% packet loss, time 0ms
			
			command terminated with exit code 1`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.FAILURE,
				transmitted: 1,
				received:    0,
				errors:      0,
			},
			wantErr: false,
		},
		{
			name: "pingHostnameNoPacketLoss",
			args: args{stdout: `PING www.google.com (172.217.12.132) 56(84) bytes of data.
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=1 ttl=61 time=25.4 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=2 ttl=61 time=27.1 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=3 ttl=61 time=26.7 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=4 ttl=61 time=24.2 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=5 ttl=61 time=28.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=6 ttl=61 time=37.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=7 ttl=61 time=21.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=8 ttl=61 time=30.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=9 ttl=61 time=27.3 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=10 ttl=61 time=27.9 ms
			
			--- www.google.com ping statistics ---
			10 packets transmitted, 10 received, 0% packet loss, time 9014ms
			rtt min/avg/max/mdev = 21.650/27.619/37.003/3.885 ms`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.SUCCESS,
				transmitted: 10,
				received:    10,
				errors:      0,
			},
			wantErr: false,
		},
		{
			name: "decodingError",
			args: args{stdout: `PING www.google.com (172.217.12.132) 56(84) bytes of data.
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=1 ttl=61 time=25.4 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=2 ttl=61 time=27.1 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=3 ttl=61 time=26.7 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=4 ttl=61 time=24.2 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=5 ttl=61 time=28.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=6 ttl=61 time=37.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=7 ttl=61 time=21.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=8 ttl=61 time=30.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=9 ttl=61 time=27.3 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=10 ttl=61 time=27.9 ms
			
			--- www.google.com ping statistics ---
			10 pacets transmitted, 10 received, 0% packet loss, time 9014ms
			rtt min/avg/max/mdev = 21.650/27.619/37.003/3.885 ms`, stderr: ""},
			wantResults: PingResults{
				outcome:     testhelper.FAILURE,
				transmitted: 0,
				received:    0,
				errors:      0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResults, err := parsePingResult(tt.args.stdout, tt.args.stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePingResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("parsePingResult() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}

func TestProcessContainerIpsPerNet(t *testing.T) {
	type args struct {
		containerID       *provider.Container
		netKey            string
		ipAddresses       []string
		netsUnderTest     map[string]netcommons.NetTestContext
		aIPVersion        netcommons.IPVersion
		wantNetsUnderTest map[string]netcommons.NetTestContext
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				containerID: &provider.Container{
					Container: &corev1.Container{},
					Status:    corev1.ContainerStatus{},
					Namespace: "namespace1",
					Podname:   "pod1",
					NodeName:  "node1",
					Runtime:   "containerd",
					UID:       "16165165165165",
				},
				netKey:        "net1",
				ipAddresses:   []string{"1.1.1.1", "2.2.2.2", "3.3.3.3", "fd00:10:244:1::3"},
				netsUnderTest: map[string]netcommons.NetTestContext{},
				aIPVersion:    netcommons.IPv4,
				wantNetsUnderTest: map[string]netcommons.NetTestContext{
					"net1": {
						TesterContainerNodeName: "",
						TesterSource: netcommons.ContainerIP{
							IP: "1.1.1.1",
							ContainerIdentifier: &provider.Container{
								Container: &corev1.Container{},
								Status:    corev1.ContainerStatus{},
								Namespace: "namespace1",
								Podname:   "pod1",
								NodeName:  "node1",
								Runtime:   "containerd",
								UID:       "16165165165165",
							},
						},
						DestTargets: []netcommons.ContainerIP{{
							IP: "2.2.2.2",
							ContainerIdentifier: &provider.Container{
								Container: &corev1.Container{},
								Status:    corev1.ContainerStatus{},
								Namespace: "namespace1",
								Podname:   "pod1",
								NodeName:  "node1",
								Runtime:   "containerd",
								UID:       "16165165165165",
							},
						}, {
							IP: "3.3.3.3",
							ContainerIdentifier: &provider.Container{
								Container: &corev1.Container{},
								Status:    corev1.ContainerStatus{},
								Namespace: "namespace1",
								Podname:   "pod1",
								NodeName:  "node1",
								Runtime:   "containerd",
								UID:       "16165165165165",
							},
						},
						},
					},
				},
			},
		},
	}
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processContainerIpsPerNet(
				tt.args.containerID,
				tt.args.netKey,
				tt.args.ipAddresses,
				tt.args.netsUnderTest,
				tt.args.aIPVersion,
				log.GetLogger(),
			)
			if !reflect.DeepEqual(tt.args.netsUnderTest, tt.args.wantNetsUnderTest) {
				t.Errorf(
					"ProcessContainerIpsPerNet() = %v, want %v",
					tt.args.netsUnderTest,
					tt.args.wantNetsUnderTest,
				)
			}
		})
	}
}

func TestBuildNetTestContext(t *testing.T) {
	type args struct {
		pods       []*provider.Pod
		aIPVersion netcommons.IPVersion
		aType      netcommons.IFType
	}
	tests := []struct {
		name              string
		args              args
		wantNetsUnderTest map[string]netcommons.NetTestContext
	}{
		{
			name: "ipv4ok",
			args: args{pods: []*provider.Pod{&pod1, &pod2},
				aIPVersion: netcommons.IPv4,
				aType:      netcommons.DEFAULT,
			},
			wantNetsUnderTest: map[string]netcommons.NetTestContext{
				"default": {
					TesterContainerNodeName: "",
					TesterSource: netcommons.ContainerIP{
						IP: "10.244.195.231",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: []netcommons.ContainerIP{{
						IP: "10.244.195.232",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test2",
							},
							Namespace: "tnf",
							Podname:   "test-1",
							NodeName:  "kind-worker4",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
						},
					},
					},
				},
			},
		},
		{
			name: "skip net test",
			args: args{pods: []*provider.Pod{&pod1, &pod3},
				aIPVersion: netcommons.IPv4,
				aType:      netcommons.DEFAULT,
			},
			wantNetsUnderTest: map[string]netcommons.NetTestContext{
				"default": {
					TesterContainerNodeName: "",
					TesterSource: netcommons.ContainerIP{
						IP: "10.244.195.231",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: nil,
				},
			},
		},
		{
			name: "ipv4ok multus",
			args: args{pods: []*provider.Pod{&pod1, &pod2},
				aIPVersion: netcommons.IPv4,
				aType:      netcommons.MULTUS,
			},

			wantNetsUnderTest: map[string]netcommons.NetTestContext{
				"tnf/mynet-ipv4-0": {
					TesterSource: netcommons.ContainerIP{
						IP: "192.168.0.3",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: []netcommons.ContainerIP{
						{
							IP: "192.168.0.4",
							ContainerIdentifier: &provider.Container{
								Container: &corev1.Container{
									Name: "test2",
								},
								Namespace: "tnf",
								Podname:   "test-1",
								NodeName:  "kind-worker4",
								Runtime:   "containerd",
								UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
							},
						},
					},
				},
				"tnf/mynet-ipv4-1": {
					TesterContainerNodeName: "",
					TesterSource: netcommons.ContainerIP{
						IP: "192.168.1.3",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: []netcommons.ContainerIP{
						{
							IP: "192.168.1.4",
							ContainerIdentifier: &provider.Container{
								Container: &corev1.Container{
									Name: "test2",
								},
								Namespace: "tnf",
								Podname:   "test-1",
								NodeName:  "kind-worker4",
								Runtime:   "containerd",
								UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
							},
						},
					},
				},
			},
		},
		{
			name: "skip multus net test",
			args: args{pods: []*provider.Pod{&pod1, &pod4},
				aIPVersion: netcommons.IPv4,
				aType:      netcommons.MULTUS,
			},
			wantNetsUnderTest: map[string]netcommons.NetTestContext{
				"tnf/mynet-ipv4-0": {
					TesterSource: netcommons.ContainerIP{
						IP: "192.168.0.3",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: nil,
				},
				"tnf/mynet-ipv4-1": {
					TesterContainerNodeName: "",
					TesterSource: netcommons.ContainerIP{
						IP: "192.168.1.3",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test1",
							},
							Namespace: "tnf",
							Podname:   "test-0",
							NodeName:  "kind-worker3",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
						},
					},
					DestTargets: nil,
				},
			},
		},
	}
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for idx := range tt.args.pods {
				tt.args.pods[idx].MultusIPs = make(map[string][]string)
				var err error
				tt.args.pods[idx].MultusIPs, err = provider.GetPodIPsPerNet(
					tt.args.pods[idx].GetAnnotations()[provider.CniNetworksStatusKey],
				)
				if err != nil {
					fmt.Printf("Could not decode networks-status annotation")
				}
			}

			gotNetsUnderTest := BuildNetTestContext(
				tt.args.pods,
				tt.args.aIPVersion,
				tt.args.aType,
				log.GetLogger(),
			)

			out, _ := json.MarshalIndent(gotNetsUnderTest, "", "")
			fmt.Printf("%s", out)
			if !reflect.DeepEqual(gotNetsUnderTest, tt.wantNetsUnderTest) {
				t.Errorf(
					"BuildNetTestContext() gotNetsUnderTest = %v, want %v",
					gotNetsUnderTest,
					tt.wantNetsUnderTest,
				)
			}
		})
	}
}

var (
	pod1 = provider.Pod{ //nolint:dupl
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "ns1",
				Annotations: map[string]string{
					"k8s.v1.cni.cncf.io/network-status": "[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.244.195.231\",\n        \"fd00:10:244:88:58fd:b191:5c13:9ce6\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-0\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"192.168.0.3\"\n    ],\n    \"mac\": \"96:e8:f5:33:9c:66\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-1\",\n    \"interface\": \"net2\",\n    \"ips\": [\n        \"192.168.1.3\"\n    ],\n    \"mac\": \"4e:c5:60:c2:1c:55\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-0\",\n    \"interface\": \"net3\",\n    \"ips\": [\n        \"3ffe:ffff::3\"\n    ],\n    \"mac\": \"ca:f5:77:b4:2f:49\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-1\",\n    \"interface\": \"net4\",\n    \"ips\": [\n        \"3ffe:ffff:0:1::3\"\n    ],\n    \"mac\": \"26:7b:69:1b:b0:5c\",\n    \"dns\": {}\n}]", //nolint:lll
				},
			},
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{
					{
						IP: "10.244.195.231",
					},
					{
						IP: "fd00:10:244:88:58fd:b191:5c13:9ce6",
					},
				},
			},
		},
		MultusIPs: map[string][]string{
			"": {},
		},
		SkipNetTests:       false,
		SkipMultusNetTests: false,
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{
					Name: "test1",
				},
				Namespace: "tnf",
				NodeName:  "kind-worker3",
				Podname:   "test-0",
				Runtime:   "containerd",
				UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
			},
		},
	}
	pod2 = provider.Pod{ //nolint:dupl
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns1",
				Annotations: map[string]string{
					"k8s.v1.cni.cncf.io/network-status": "[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.244.195.232\",\n        \"fd00:10:244:88:58fd:b191:5c13:9ce7\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-0\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"192.168.0.4\"\n    ],\n    \"mac\": \"96:e8:f5:33:9c:67\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-1\",\n    \"interface\": \"net2\",\n    \"ips\": [\n        \"192.168.1.4\"\n    ],\n    \"mac\": \"4e:c5:60:c2:1c:56\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-0\",\n    \"interface\": \"net3\",\n    \"ips\": [\n        \"3ffe:ffff::4\"\n    ],\n    \"mac\": \"ca:f5:77:b4:2f:50\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-1\",\n    \"interface\": \"net4\",\n    \"ips\": [\n        \"3ffe:ffff:0:1::4\"\n    ],\n    \"mac\": \"26:7b:69:1b:b0:5d\",\n    \"dns\": {}\n}]", //nolint:lll
				},
			},
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{
					{
						IP: "10.244.195.232",
					},
					{
						IP: "fd00:10:244:88:58fd:b191:5c13:9ce7",
					},
				},
			},
		},
		MultusIPs: map[string][]string{
			"": {},
		},
		SkipNetTests:       false,
		SkipMultusNetTests: false,
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{
					Name: "test2",
				},
				Namespace: "tnf",
				NodeName:  "kind-worker4",
				Podname:   "test-1",
				Runtime:   "containerd",
				UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
			},
		},
	}
	pod3 = provider.Pod{ //nolint:dupl
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns1",
				Annotations: map[string]string{
					"k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.244.195.232\",\n        \"fd00:10:244:88:58fd:b191:5c13:9ce7\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-0\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"192.168.0.4\"\n    ],\n    \"mac\": \"96:e8:f5:33:9c:67\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-1\",\n    \"interface\": \"net2\",\n    \"ips\": [\n        \"192.168.1.4\"\n    ],\n    \"mac\": \"4e:c5:60:c2:1c:56\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-0\",\n    \"interface\": \"net3\",\n    \"ips\": [\n        \"3ffe:ffff::4\"\n    ],\n    \"mac\": \"ca:f5:77:b4:2f:50\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-1\",\n    \"interface\": \"net4\",\n    \"ips\": [\n        \"3ffe:ffff:0:1::4\"\n    ],\n    \"mac\": \"26:7b:69:1b:b0:5d\",\n    \"dns\": {}\n}]", //nolint:lll
				},
			},
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{
					{
						IP: "10.244.195.232",
					},
					{
						IP: "fd00:10:244:88:58fd:b191:5c13:9ce7",
					},
				},
			},
		},
		MultusIPs: map[string][]string{
			"": {},
		},
		SkipNetTests:       true,
		SkipMultusNetTests: false,
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{
					Name: "test2",
				},
				Namespace: "tnf",
				NodeName:  "kind-worker4",
				Podname:   "test-1",
				Runtime:   "containerd",
				UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
			},
		},
	}
	pod4 = provider.Pod{ //nolint:dupl
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns1",
				Annotations: map[string]string{
					"k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.244.195.232\",\n        \"fd00:10:244:88:58fd:b191:5c13:9ce7\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-0\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"192.168.0.4\"\n    ],\n    \"mac\": \"96:e8:f5:33:9c:67\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv4-1\",\n    \"interface\": \"net2\",\n    \"ips\": [\n        \"192.168.1.4\"\n    ],\n    \"mac\": \"4e:c5:60:c2:1c:56\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-0\",\n    \"interface\": \"net3\",\n    \"ips\": [\n        \"3ffe:ffff::4\"\n    ],\n    \"mac\": \"ca:f5:77:b4:2f:50\",\n    \"dns\": {}\n},{\n    \"name\": \"tnf/mynet-ipv6-1\",\n    \"interface\": \"net4\",\n    \"ips\": [\n        \"3ffe:ffff:0:1::4\"\n    ],\n    \"mac\": \"26:7b:69:1b:b0:5d\",\n    \"dns\": {}\n}]", //nolint:lll
				},
			},
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{
					{
						IP: "10.244.195.232",
					},
					{
						IP: "fd00:10:244:88:58fd:b191:5c13:9ce7",
					},
				},
			},
		},
		MultusIPs: map[string][]string{
			"": {},
		},
		SkipNetTests:       false,
		SkipMultusNetTests: true,
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{
					Name: "test2",
				},
				Namespace: "tnf",
				NodeName:  "kind-worker4",
				Podname:   "test-1",
				Runtime:   "containerd",
				UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
			},
		},
	}
)

func TestRunNetworkingTests(t *testing.T) {
	type args struct {
		netsUnderTest map[string]netcommons.NetTestContext
		count         int
		aIPVersion    netcommons.IPVersion
	}
	tests := []struct {
		name            string
		args            args
		wantReport      testhelper.FailureReasonOut
		testPingSuccess bool
	}{
		{name: "ok",
			args: args{netsUnderTest: map[string]netcommons.NetTestContext{"default": {
				TesterContainerNodeName: "",
				TesterSource: netcommons.ContainerIP{
					IP: "10.244.195.231",
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: "test1",
						},
						Namespace: "tnf",
						Podname:   "test-0",
						NodeName:  "kind-worker3",
						Runtime:   "containerd",
						UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
					},
				},
				DestTargets: []netcommons.ContainerIP{{
					IP: "10.244.195.232",
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: "test2",
						},
						Namespace: "tnf",
						Podname:   "test-1",
						NodeName:  "kind-worker4",
						Runtime:   "containerd",
						UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
					},
				},
				},
			},
			}, count: 10, aIPVersion: netcommons.IPv4,
			},
			wantReport: testhelper.FailureReasonOut{
				CompliantObjectsOut: []*testhelper.ReportObject{
					{
						ObjectType: "ICMP result",
						ObjectFieldsKeys: []string{
							testhelper.ReasonForCompliance,
							"Namespace",
							testhelper.PodName,
							testhelper.ContainerName,
							testhelper.NetworkName,
							testhelper.SourceIP,
							testhelper.DestinationNamespace,
							testhelper.DestinationPodName,
							testhelper.DestinationContainerName,
							testhelper.DestinationIP,
						},
						ObjectFieldsValues: []string{
							"Pinging destination container/IP from source container (identified by Namespace/Pod Name/Container Name) Succeeded",
							"tnf",
							"test-0",
							"test1",
							"default",
							"10.244.195.231",
							"tnf",
							"test-1",
							"test2",
							"10.244.195.232",
						},
					},
					{
						ObjectType:       "Network",
						ObjectFieldsKeys: []string{testhelper.ReasonForCompliance, testhelper.NetworkName},
						ObjectFieldsValues: []string{
							"ICMP tests were successful for all 1 IP source/destination in this network",
							"default",
						},
					},
				},
				NonCompliantObjectsOut: []*testhelper.ReportObject{},
			},
			testPingSuccess: true,
		},
		{name: "noNetToTest",
			args: args{netsUnderTest: map[string]netcommons.NetTestContext{},
				count: 10, aIPVersion: netcommons.IPv4,
			},
			wantReport:      testhelper.FailureReasonOut{},
			testPingSuccess: true,
		},
		{name: "only one container",
			args: args{netsUnderTest: map[string]netcommons.NetTestContext{"default": {
				TesterContainerNodeName: "",
				TesterSource: netcommons.ContainerIP{
					IP: "10.244.195.231",
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: "test1",
						},
						Namespace: "tnf",
						Podname:   "test-0",
						NodeName:  "kind-worker3",
						Runtime:   "containerd",
						UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
					},
				},
				DestTargets: []netcommons.ContainerIP{},
			},
			}, count: 10, aIPVersion: netcommons.IPv4,
			},
			wantReport:      testhelper.FailureReasonOut{},
			testPingSuccess: true,
		},
		{name: "ping fails",
			args: args{netsUnderTest: map[string]netcommons.NetTestContext{"default": {
				TesterContainerNodeName: "",
				TesterSource: netcommons.ContainerIP{
					IP: "10.244.195.231",
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: "test1",
						},
						Namespace: "tnf",
						Podname:   "test-0",
						NodeName:  "kind-worker3",
						Runtime:   "containerd",
						UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd261",
					},
				},
				DestTargets: []netcommons.ContainerIP{{
					IP: "10.244.195.232",
					ContainerIdentifier: &provider.Container{
						Container: &corev1.Container{
							Name: "test2",
						},
						Namespace: "tnf",
						Podname:   "test-1",
						NodeName:  "kind-worker4",
						Runtime:   "containerd",
						UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd262",
					},
				},
					{
						IP: "10.244.195.233",
						ContainerIdentifier: &provider.Container{
							Container: &corev1.Container{
								Name: "test3",
							},
							Namespace: "tnf",
							Podname:   "test-1",
							NodeName:  "kind-worker4",
							Runtime:   "containerd",
							UID:       "a94eea4619dbf6046e843955744e823ea4e9d83daa435acc08973f4c35ddd264",
						},
					},
				},
			},
			}, count: 10, aIPVersion: netcommons.IPv4,
			},
			wantReport: testhelper.FailureReasonOut{
				CompliantObjectsOut: []*testhelper.ReportObject{},
				NonCompliantObjectsOut: []*testhelper.ReportObject{
					{
						ObjectType: "ICMP result",
						ObjectFieldsKeys: []string{
							testhelper.ReasonForNonCompliance,
							"Namespace",
							testhelper.PodName,
							testhelper.ContainerName,
							testhelper.NetworkName,
							testhelper.SourceIP,
							testhelper.DestinationNamespace,
							testhelper.DestinationPodName,
							testhelper.DestinationContainerName,
							testhelper.DestinationIP,
						},
						ObjectFieldsValues: []string{
							"Pinging destination container/IP from source container (identified by Namespace/Pod Name/Container Name) Failed",
							"tnf",
							"test-0",
							"test1",
							"default",
							"10.244.195.231",
							"tnf",
							"test-1",
							"test2",
							"10.244.195.232",
						},
					},
					{
						ObjectType: "ICMP result",
						ObjectFieldsKeys: []string{
							testhelper.ReasonForNonCompliance,
							"Namespace",
							testhelper.PodName,
							testhelper.ContainerName,
							testhelper.NetworkName,
							testhelper.SourceIP,
							testhelper.DestinationNamespace,
							testhelper.DestinationPodName,
							testhelper.DestinationContainerName,
							testhelper.DestinationIP,
						},
						ObjectFieldsValues: []string{
							"Pinging destination container/IP from source container (identified by Namespace/Pod Name/Container Name) Failed",
							"tnf",
							"test-0",
							"test1",
							"default",
							"10.244.195.231",
							"tnf",
							"test-1",
							"test3",
							"10.244.195.233",
						},
					},
					{
						ObjectType:       "Network",
						ObjectFieldsKeys: []string{testhelper.ReasonForNonCompliance, testhelper.NetworkName},
						ObjectFieldsValues: []string{
							"ICMP tests failed for 2 IP source/destination in this network",
							"default",
						},
					},
				},
			},
			testPingSuccess: false,
		},
	}

	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testPingSuccess {
				TestPing = TestPingSuccess
			} else {
				TestPing = TestPingFailure
			}
			gotReport, _ := RunNetworkingTests(
				tt.args.netsUnderTest,
				tt.args.count,
				tt.args.aIPVersion,
				log.GetLogger(),
			)
			if !gotReport.Equal(tt.wantReport) {
				t.Errorf(
					"RunNetworkingTests() gotReport = %q, want %q",
					testhelper.FailureReasonOutTestString(gotReport),
					testhelper.FailureReasonOutTestString(tt.wantReport),
				)
			}
		})
	}
}

var TestPingSuccess = func(sourceContainerID *provider.Container, targetContainerIP netcommons.ContainerIP, count int) (results PingResults, err error) {
	return PingResults{outcome: testhelper.SUCCESS, transmitted: 10, received: 10, errors: 0}, nil
}

var TestPingFailure = func(sourceContainerID *provider.Container, targetContainerIP netcommons.ContainerIP, count int) (results PingResults, err error) {
	return PingResults{
			outcome:     testhelper.FAILURE,
			transmitted: 10,
			received:    5,
			errors:      5,
		}, fmt.Errorf(
			"ping failed",
		)
}
