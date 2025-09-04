// Copyright (C) 2020-2024 Red Hat, Inc.
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

package services

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netcommons"
	corev1 "k8s.io/api/core/v1"
)

// GetServiceIPVersion Determines the IP stack type of a Kubernetes Service
//
// The function examines a service's ClusterIP, IPFamilyPolicy, and any
// additional ClusterIPs to decide whether it is single‑stack IPv4,
// single‑stack IPv6, or dual‑stack. It returns an IPVersion value along
// with an error if the configuration cannot be resolved or violates
// expectations. Logging statements provide debug context for each decision
// path.
func GetServiceIPVersion(aService *corev1.Service) (result netcommons.IPVersion, err error) {
	ipver, err := netcommons.GetIPVersion(aService.Spec.ClusterIP)
	if err != nil {
		err = fmt.Errorf("%s cannot get aService clusterIP version", ToString(aService))
		return result, err
	}
	if aService.Spec.IPFamilyPolicy == nil {
		err = fmt.Errorf("%s does not have a IPFamilyPolicy configured", ToString(aService))
		return result, err
	}
	if *aService.Spec.IPFamilyPolicy == corev1.IPFamilyPolicySingleStack &&
		ipver == netcommons.IPv6 {
		log.Debug("%s is single stack ipv6", ToString(aService))
		return netcommons.IPv6, nil
	}
	if *aService.Spec.IPFamilyPolicy == corev1.IPFamilyPolicySingleStack &&
		ipver == netcommons.IPv4 {
		log.Debug("%s is single stack ipv4", ToString(aService))
		return netcommons.IPv4, nil
	}
	if (*aService.Spec.IPFamilyPolicy == corev1.IPFamilyPolicyPreferDualStack ||
		*aService.Spec.IPFamilyPolicy == corev1.IPFamilyPolicyRequireDualStack) &&
		len(aService.Spec.ClusterIPs) < 2 {
		err = fmt.Errorf("%s is dual stack but has only zero or one ClusterIPs", ToString(aService))
		return result, err
	}

	res, err := isClusterIPsDualStack(aService.Spec.ClusterIPs)
	if err != nil {
		err = fmt.Errorf("%s, err:%s", ToString(aService), err)
		return result, err
	}
	if res {
		log.Debug("%s is dual-stack", ToString(aService))
		return netcommons.IPv4v6, nil
	}

	err = fmt.Errorf("%s is not compliant, it is not single stack ipv6 or dual stack", ToString(aService))
	return result, err
}

// ToString Formats a service's namespace, name, cluster IPs, and IP family into a readable string
//
// This function takes a pointer to a Kubernetes Service object and constructs a
// single-line description that includes the service's namespace, name, primary
// ClusterIP, and all associated ClusterIPs. It uses string formatting to
// concatenate these fields in a human‑readable format, which is useful for
// logging and error messages. The result is returned as a plain string.
func ToString(aService *corev1.Service) (out string) {
	return fmt.Sprintf("Service ns: %s, name: %s ClusterIP:%s ClusterIPs: %v", aService.Namespace,
		aService.Name,
		aService.Spec.ClusterIP,
		aService.Spec.ClusterIPs)
}

// ToStringSlice Lists services with namespace, name, ClusterIP and IP addresses
//
// The function iterates over a slice of service objects, appending formatted
// information for each one to a single string. For every service it records the
// namespace, name, primary ClusterIP, and any additional cluster IPs. The
// resulting multi-line string is returned.
func ToStringSlice(manyServices []*corev1.Service) (out string) {
	for _, aService := range manyServices {
		out += fmt.Sprintf("Service ns: %s, name: %s ClusterIP:%s ClusterIPs: %v\n", aService.Namespace,
			aService.Name,
			aService.Spec.ClusterIP,
			aService.Spec.ClusterIPs)
	}
	return out
}

// isClusterIPsDualStack verifies that a service's ClusterIPs include both IPv4 and IPv6 addresses
//
// The function iterates over each IP string, determines its version using an
// external helper, and records whether any IPv4 or IPv6 address appears. If
// both types are present it returns true; otherwise false. Errors from the
// helper cause an early return with a descriptive message.
func isClusterIPsDualStack(ips []string) (result bool, err error) {
	var hasIPv4, hasIPv6 bool
	for _, ip := range ips {
		ipver, err := netcommons.GetIPVersion(ip)
		if err != nil {
			return result, fmt.Errorf("cannot get aService ClusterIPs (%s)  version - err: %v", ip, err)
		}
		switch ipver {
		case netcommons.IPv4:
			hasIPv4 = true
		case netcommons.IPv6:
			hasIPv6 = true
		case netcommons.IPv4v6, netcommons.Undefined:
		}
	}
	if hasIPv4 && hasIPv6 {
		return true, nil
	}
	return false, nil
}
