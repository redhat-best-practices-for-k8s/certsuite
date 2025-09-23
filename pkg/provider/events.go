// Copyright (C) 2022 Red Hat, Inc.
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

package provider

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// Event Represents a Kubernetes event with access to all core event data
//
// The type embeds the standard Kubernetes Event structure, giving it direct
// access to fields such as CreationTimestamp, InvolvedObject, Reason, and
// Message. It provides a convenient String method that formats these key
// properties into a single readable string for logging or debugging purposes.
// This struct is used throughout the provider package to encapsulate event
// information while keeping the original corev1.Event behavior intact.
type Event struct {
	*corev1.Event
}

// NewEvent Wraps a Kubernetes event object
//
// The function receives a pointer to a corev1.Event and returns an Event
// instance that encapsulates the original event. It assigns the passed event to
// the internal field of the returned struct, enabling further processing within
// the provider package. No additional transformation or validation is
// performed.
func NewEvent(aEvent *corev1.Event) (out Event) {
	out.Event = aEvent
	return out
}

// Event.String Formats event data into a readable string
//
// This method constructs a formatted text representation of an event, including
// its timestamp, involved object, reason, and message. It uses standard
// formatting utilities to combine these fields into a single line. The
// resulting string is returned for display or logging purposes.
func (e *Event) String() string {
	return fmt.Sprintf("timestamp=%s involved object=%s reason=%s message=%s", e.CreationTimestamp.Time, e.InvolvedObject, e.Reason, e.Message)
}
