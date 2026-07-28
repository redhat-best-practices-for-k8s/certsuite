// Copyright (C) 2020-2026 Red Hat, Inc.
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
	"reflect"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

func TestCreateLabels(t *testing.T) {
	type args struct {
		labelStrings []string
	}
	tests := []struct {
		name             string
		args             args
		wantLabelObjects []labelObject
	}{
		{
			name:             "ok",
			args:             args{labelStrings: []string{"redhat-best-practices-for-k8s.com/generic: target"}},
			wantLabelObjects: []labelObject{{LabelKey: "redhat-best-practices-for-k8s.com/generic", LabelValue: "target"}},
		},
		{
			name:             "ok1",
			args:             args{labelStrings: []string{"redhat-best-practices-for-k8s.com/generic   : 1"}},
			wantLabelObjects: []labelObject{{LabelKey: "redhat-best-practices-for-k8s.com/generic", LabelValue: "1"}},
		},
		{
			name: "nok",
			args: args{labelStrings: []string{"redhat-best-practices-for-k8s.com/generic= target"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLabelObjects := CreateLabels(tt.args.labelStrings); !reflect.DeepEqual(gotLabelObjects, tt.wantLabelObjects) {
				t.Errorf("CreateLabels() = %v, want %v", gotLabelObjects, tt.wantLabelObjects)
			}
		})
	}
}

func TestMatchesAnyLabel(t *testing.T) {
	tests := []struct {
		name           string
		resourceLabels map[string]string
		selectors      []labelObject
		want           bool
	}{
		{
			name:           "match first selector",
			resourceLabels: map[string]string{"app": "target"},
			selectors:      []labelObject{{LabelKey: "app", LabelValue: "target"}},
			want:           true,
		},
		{
			name:           "match second selector",
			resourceLabels: map[string]string{"role": "worker"},
			selectors:      []labelObject{{LabelKey: "app", LabelValue: "target"}, {LabelKey: "role", LabelValue: "worker"}},
			want:           true,
		},
		{
			name:           "no match",
			resourceLabels: map[string]string{"app": "other"},
			selectors:      []labelObject{{LabelKey: "app", LabelValue: "target"}},
			want:           false,
		},
		{
			name:           "nil labels map",
			resourceLabels: nil,
			selectors:      []labelObject{{LabelKey: "app", LabelValue: "target"}},
			want:           false,
		},
		{
			name:           "empty selectors",
			resourceLabels: map[string]string{"app": "target"},
			selectors:      []labelObject{},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchesAnyLabel(tt.resourceLabels, tt.selectors))
		})
	}
}

func TestNamespacesListToStringList(t *testing.T) {
	testCases := []struct {
		testList       []configuration.Namespace
		expectedOutput []string
	}{
		{
			testList: []configuration.Namespace{
				{
					Name: "ns1",
				},
				{
					Name: "ns2",
				},
			},
			expectedOutput: []string{"ns1", "ns2"},
		},
		{
			testList:       []configuration.Namespace{},
			expectedOutput: nil,
		},
		{
			testList: []configuration.Namespace{
				{Name: "name1"},
				{Name: "name1"},
			},
			expectedOutput: []string{"name1", "name1"},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, namespacesListToStringList(tc.testList))
	}
}
