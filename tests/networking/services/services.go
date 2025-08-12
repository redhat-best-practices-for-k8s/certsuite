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

func ToString(aService *corev1.Service) (out string) {
	return fmt.Sprintf("Service ns: %s, name: %s ClusterIP:%s ClusterIPs: %v", aService.Namespace,
		aService.Name,
		aService.Spec.ClusterIP,
		aService.Spec.ClusterIPs)
}

func ToStringSlice(manyServices []*corev1.Service) (out string) {
	for _, aService := range manyServices {
		out += fmt.Sprintf("Service ns: %s, name: %s ClusterIP:%s ClusterIPs: %v\n", aService.Namespace,
			aService.Name,
			aService.Spec.ClusterIP,
			aService.Spec.ClusterIPs)
	}
	return out
}

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
