// Copyright (C) 2022-2024 Red Hat, Inc.
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

package autodiscover

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestFindAbnormalEvents(t *testing.T) {
	testCases := []struct {
		expectedEvents []*corev1.Event
	}{
		{
			expectedEvents: []*corev1.Event{
				{
					Reason: "FailedMount",
					Type:   "Warning",
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "test-event",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		for _, event := range testCase.expectedEvents {
			runtimeObjects = append(runtimeObjects, event)
		}

		// Create fake client
		client := k8sfake.NewSimpleClientset(runtimeObjects...)
		abnormalEvents := findAbnormalEvents(client.CoreV1(), []string{"test-namespace"})
		assert.Len(t, abnormalEvents, len(testCase.expectedEvents))

		for _, event := range abnormalEvents {
			for _, event2 := range testCase.expectedEvents {
				if event.Name == event2.Name {
					assert.Equal(t, event.Reason, event2.Reason)
					assert.Equal(t, event.Type, event2.Type)
				}
			}
		}
	}
}
