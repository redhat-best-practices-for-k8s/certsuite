// Copyright (C) 2026 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.

package autodiscover

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestRetryAPICallRetriesDiscoveryListCall(t *testing.T) {
	savedPolicy := defaultRetryPolicy
	defaultRetryPolicy = retryPolicy{maxRetries: 2}
	t.Cleanup(func() { defaultRetryPolicy = savedPolicy })

	pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv1"}}
	client := k8sfake.NewClientset(pv)
	calls := 0
	client.PrependReactor("list", "persistentvolumes", func(k8stesting.Action) (bool, runtime.Object, error) {
		calls++
		if calls == 1 {
			return true, nil, k8serrors.NewServiceUnavailable("apiserver restarting")
		}
		// Fall through to the default object tracker.
		return false, nil, nil
	})

	pvs, err := getPersistentVolumes(client.CoreV1())

	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	require.Len(t, pvs, 1)
	assert.Equal(t, "pv1", pvs[0].Name)
}

func TestRetryAPICallDoesNotRetryPermanentDiscoveryError(t *testing.T) {
	// Zero delays keep the test fast if a regression starts retrying permanent errors.
	savedPolicy := defaultRetryPolicy
	defaultRetryPolicy = retryPolicy{maxRetries: 2}
	t.Cleanup(func() { defaultRetryPolicy = savedPolicy })

	client := k8sfake.NewClientset()
	calls := 0
	client.PrependReactor("list", "persistentvolumes", func(k8stesting.Action) (bool, runtime.Object, error) {
		calls++
		return true, nil, k8serrors.NewForbidden(schema.GroupResource{Resource: "persistentvolumes"}, "", errors.New("denied"))
	})

	_, err := getPersistentVolumes(client.CoreV1())

	require.Error(t, err)
	assert.True(t, k8serrors.IsForbidden(err), "error should keep its API status through wrapping: %v", err)
	assert.Equal(t, 1, calls)
}

func TestRetryWithPolicyClampsInitialToMaxDelay(t *testing.T) {
	attempts := 0
	start := time.Now()
	_, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (struct{}, error) {
		attempts++
		if attempts == 1 {
			return struct{}{}, k8serrors.NewServiceUnavailable("temporarily unavailable")
		}
		return struct{}{}, nil
	}, retryPolicy{maxRetries: 1, initial: time.Hour, maxDelay: 10 * time.Millisecond})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestRetryWithPolicyRetriesTransientFailure(t *testing.T) {
	attempts := 0
	got, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", k8serrors.NewServiceUnavailable("temporarily unavailable")
		}
		return "success", nil
	}, retryPolicy{maxRetries: 3, initial: 0, maxDelay: 0})

	require.NoError(t, err)
	assert.Equal(t, "success", got)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithPolicyDoesNotRetryPermanentFailure(t *testing.T) {
	attempts := 0
	wantErr := k8serrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "", errors.New("access denied"))
	_, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (struct{}, error) {
		attempts++
		return struct{}{}, wantErr
	}, retryPolicy{maxRetries: 3, initial: 0, maxDelay: 0})

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithPolicyStopsAfterRetryLimit(t *testing.T) {
	attempts := 0
	wantErr := k8serrors.NewInternalError(errors.New("etcd leader changed"))
	_, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (struct{}, error) {
		attempts++
		return struct{}{}, wantErr
	}, retryPolicy{maxRetries: 2, initial: 0, maxDelay: 0})

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithPolicyHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	_, err := retryWithPolicy(ctx, "list pods", func(context.Context) (struct{}, error) {
		attempts++
		return struct{}{}, k8serrors.NewServiceUnavailable("temporarily unavailable")
	}, retryPolicy{maxRetries: 3, initial: time.Hour, maxDelay: time.Hour})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithPolicyHonorsContextCancellationDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempts := 0
	start := time.Now()
	_, err := retryWithPolicy(ctx, "list pods", func(context.Context) (struct{}, error) {
		attempts++
		time.AfterFunc(20*time.Millisecond, cancel)
		return struct{}{}, k8serrors.NewServiceUnavailable("temporarily unavailable")
	}, retryPolicy{maxRetries: 3, initial: time.Hour, maxDelay: time.Hour})

	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorContains(t, err, "temporarily unavailable", "last API error should be kept")
	assert.Equal(t, 1, attempts)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestRetryWithPolicyZeroRetriesAttemptsOnce(t *testing.T) {
	attempts := 0
	wantErr := k8serrors.NewServiceUnavailable("temporarily unavailable")
	_, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (struct{}, error) {
		attempts++
		return struct{}{}, wantErr
	}, retryPolicy{maxRetries: 0, initial: 0, maxDelay: 0})

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithPolicyPassesContextToCall(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "marker")
	got, err := retryWithPolicy(ctx, "list pods", func(callCtx context.Context) (string, error) {
		v, _ := callCtx.Value(ctxKey{}).(string)
		return v, nil
	}, retryPolicy{maxRetries: 0})

	require.NoError(t, err)
	assert.Equal(t, "marker", got)
}

func TestRetryWithPolicyBackoffDoublesUpToMaxDelay(t *testing.T) {
	// Waits are 10ms, 20ms, then capped at 20ms: at least 50ms in total.
	attempts := 0
	start := time.Now()
	_, err := retryWithPolicy(context.Background(), "list pods", func(context.Context) (struct{}, error) {
		attempts++
		return struct{}{}, k8serrors.NewServiceUnavailable("temporarily unavailable")
	}, retryPolicy{maxRetries: 3, initial: 10 * time.Millisecond, maxDelay: 20 * time.Millisecond})
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, 4, attempts)
	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
	assert.Less(t, elapsed, 2*time.Second)
}

func TestIsTransientAPIError(t *testing.T) {
	podsGR := schema.GroupResource{Resource: "pods"}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		// Already retried by client-go's rest client, so not retried again here.
		{name: "too many requests", err: k8serrors.NewTooManyRequests("rate limited", 0), want: false},
		{name: "EOF", err: io.EOF, want: false},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, want: false},
		{name: "wrapped connection reset", err: fmt.Errorf("read tcp: %w", syscall.ECONNRESET), want: false},

		{name: "server timeout", err: k8serrors.NewServerTimeout(podsGR, "list", 1), want: true},
		{name: "timeout (504)", err: k8serrors.NewTimeoutError("timed out", 1), want: true},
		{name: "service unavailable", err: k8serrors.NewServiceUnavailable("unavailable"), want: true},
		{name: "internal error", err: k8serrors.NewInternalError(errors.New("boom")), want: true},
		{name: "wrapped service unavailable", err: fmt.Errorf("listing pods: %w", k8serrors.NewServiceUnavailable("unavailable")), want: true},
		{name: "not found", err: k8serrors.NewNotFound(podsGR, "example"), want: false},
		{name: "forbidden", err: k8serrors.NewForbidden(podsGR, "", errors.New("access denied")), want: false},
		{name: "connection refused", err: &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, want: true},
		{name: "net timeout", err: os.ErrDeadlineExceeded, want: true},
		{name: "net non-timeout", err: &net.DNSError{Err: "no such host"}, want: false},
		{name: "other error", err: errors.New("bad request"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTransientAPIError(tt.err))
		})
	}
}
