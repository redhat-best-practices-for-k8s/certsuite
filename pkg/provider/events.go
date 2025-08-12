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

// Event represents a Kubernetes event from the core API.
//
// It embeds *corev1.Event to provide access to all standard event fields
// such as Type, Reason, Message, and involved object information.
// The String method formats the event into a human‑readable string,
// typically using Sprintf to combine relevant fields. This struct is
// used by the provider package to expose events in a convenient
// wrapper that retains full functionality of the underlying corev1.Event.
type Event struct {
	*corev1.Event
}

// NewEvent creates an internal Event representation from a Kubernetes event.
//
// It accepts a pointer to a corev1.Event object and extracts the relevant
// information such as type, reason, message, source, and involved object.
// The function returns an Event value that is used by the provider package for
// further processing or reporting.
func NewEvent(aEvent *corev1.Event) (out Event) {
	out.Event = aEvent
	return out
}

// String returns a human readable representation of the event.
//
// It formats the event's type, message and timestamp into a single string,
// using fmt.Sprintf to create a concise description suitable for logging or
// display. The returned string contains no newlines and is safe to use in
// plain text outputs.
func (e *Event) String() string {
	return fmt.Sprintf("timestamp=%s involved object=%s reason=%s message=%s", e.CreationTimestamp.Time, e.InvolvedObject, e.Reason, e.Message)
}
