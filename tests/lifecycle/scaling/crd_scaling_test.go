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

package scaling

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets"
	"github.com/stretchr/testify/assert"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/scale"
	k8stesting "k8s.io/client-go/testing"
)

// fakeScaleImpl implements scale.ScaleInterface for testing
type fakeScaleImpl struct {
	getResult *v1autoscaling.Scale
	getErr    error
	updateErr error
}

func (f *fakeScaleImpl) Get(_ context.Context, _ schema.GroupResource, _ string, _ metav1.GetOptions) (*v1autoscaling.Scale, error) {
	return f.getResult, f.getErr
}

//nolint:gocritic
func (f *fakeScaleImpl) Update(_ context.Context, _ schema.GroupResource, s *v1autoscaling.Scale, _ metav1.UpdateOptions) (*v1autoscaling.Scale, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return s, nil
}

//nolint:gocritic
func (f *fakeScaleImpl) Patch(_ context.Context, _ schema.GroupVersionResource, _ string, _ types.PatchType, _ []byte, _ metav1.PatchOptions) (*v1autoscaling.Scale, error) {
	return nil, nil
}

// fakeScalesGetter implements scale.ScalesGetter for testing
type fakeScalesGetter struct {
	si *fakeScaleImpl
}

func (f *fakeScalesGetter) Scales(_ string) scale.ScaleInterface {
	return f.si
}

func TestScaleHpaCRDHelper(t *testing.T) {
	testCases := []struct {
		getResult      error
		updateResult   error
		expectedOutput bool
	}{
		{ // Test Case 1 - No errors issuing the get or update
			getResult:      nil,
			updateResult:   nil,
			expectedOutput: true,
		},
		{ // Test Case 2 - Error updating the HPA
			getResult:      nil,
			updateResult:   errors.New("this is an error"),
			expectedOutput: false,
		},
		{ // Test Case 3 - Error getting the HPA
			getResult:      errors.New("this is an error"),
			updateResult:   nil,
			expectedOutput: false,
		},
	}

	defer clientsholder.ClearTestClientsHolder()

	// Always return that the scaling is complete
	origFunc := podsets.WaitForScalingToComplete
	defer func() {
		podsets.WaitForScalingToComplete = origFunc
	}()
	podsets.WaitForScalingToComplete = func(ns, name string, timeout time.Duration, groupResourceSchema schema.GroupResource, logger *log.Logger) bool {
		return true
	}

	int32Ptr := func(i int32) *int32 { return &i }
	gr := schema.GroupResource{Group: "apps", Resource: "myresource"}

	for _, tc := range testCases {
		hpatest := &v1autoscaling.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hpaName",
			},
			Spec: v1autoscaling.HorizontalPodAutoscalerSpec{
				MinReplicas: int32Ptr(1),
				MaxReplicas: 3,
			},
		}

		var runtimeObjs []runtime.Object
		runtimeObjs = append(runtimeObjs, hpatest)
		clientsholder.GetTestClientsHolder(runtimeObjs)

		client := k8sfake.Clientset{}
		client.AddReactor("get", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.getResult
		})

		client.AddReactor("update", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.updateResult
		})

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		result := scaleHpaCRDHelper(client.AutoscalingV1().HorizontalPodAutoscalers("ns1"), "hpaName", "cr1", "ns1", 1, 3, 10*time.Second, gr, log.GetLogger())
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestScaleCrHelper(t *testing.T) {
	testCases := []struct {
		description    string
		getErr         error
		updateErr      error
		expectedOutput bool
	}{
		{
			description:    "No errors",
			getErr:         nil,
			updateErr:      nil,
			expectedOutput: true,
		},
		{
			description:    "Error updating the scale object",
			getErr:         nil,
			updateErr:      errors.New("update error"),
			expectedOutput: false,
		},
		{
			description:    "Error getting the scale object",
			getErr:         errors.New("get error"),
			updateErr:      nil,
			expectedOutput: false,
		},
	}

	// Always return that the scaling is complete
	origFunc := podsets.WaitForScalingToComplete
	defer func() {
		podsets.WaitForScalingToComplete = origFunc
	}()
	podsets.WaitForScalingToComplete = func(ns, name string, timeout time.Duration, groupResourceSchema schema.GroupResource, logger *log.Logger) bool {
		return true
	}

	gr := schema.GroupResource{Group: "apps", Resource: "myresource"}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			crScale := &provider.CrScale{
				Scale: &v1autoscaling.Scale{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cr1",
						Namespace: "ns1",
					},
					Spec: v1autoscaling.ScaleSpec{
						Replicas: 1,
					},
				},
			}

			scaleObj := &v1autoscaling.Scale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cr1",
					Namespace: "ns1",
				},
				Spec: v1autoscaling.ScaleSpec{
					Replicas: 1,
				},
			}

			fakeGetter := &fakeScalesGetter{
				si: &fakeScaleImpl{
					getResult: scaleObj,
					getErr:    tc.getErr,
					updateErr: tc.updateErr,
				},
			}

			var logArchive strings.Builder
			log.SetupLogger(&logArchive, "INFO")
			result := scaleCrHelper(fakeGetter, gr, crScale, 2, true, 10*time.Second, log.GetLogger())
			assert.Equal(t, tc.expectedOutput, result)
		})
	}
}
