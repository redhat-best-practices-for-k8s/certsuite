// // Copyright (C) 2022 Red Hat, Inc.
// //
// // This program is free software; you can redistribute it and/or modify
// // it under the terms of the GNU General Public License as published by
// // the Free Software Foundation; either version 2 of the License, or
// // (at your option) any later version.
// //
// // This program is distributed in the hope that it will be useful,
// // but WITHOUT ANY WARRANTY; without even the implied warranty of
// // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// // GNU General Public License for more details.
// //
// // You should have received a copy of the GNU General Public License along
// // with this program; if not, write to the Free Software Foundation, Inc.,
// // 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package rbac

func roleBindingOutOfNamespace(roleBindingNamespace, podNamespace, roleBindingName, serviceAccountName string) bool {
	// Skip if the rolebinding namespace is part of the pod namespace.
	if roleBindingNamespace == podNamespace {
		return false
	}

	// RoleBinding is in another namespace and the service account names match.
	if roleBindingName == serviceAccountName {
		return true
	}

	return false
}

func IsRoleBindingOutOfNamespace(podNamespace, serviceAccountName, roleBindingNamespace, roleBindingName string) bool {
	return roleBindingOutOfNamespace(roleBindingNamespace, podNamespace, roleBindingName, serviceAccountName)
}
