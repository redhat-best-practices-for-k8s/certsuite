// Copyright (C) 2026 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.

package autodiscover

import (
	"context"
	"fmt"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilnet "k8s.io/apimachinery/pkg/util/net"
)

const (
	apiRetryMaxRetries    = 3
	apiRetryInitial       = time.Second
	apiRetryMaxDelay      = 4 * time.Second
	apiRetryBackoffFactor = 2
)

type retryPolicy struct {
	maxRetries int
	initial    time.Duration
	maxDelay   time.Duration
}

var defaultRetryPolicy = retryPolicy{
	maxRetries: apiRetryMaxRetries,
	initial:    apiRetryInitial,
	maxDelay:   apiRetryMaxDelay,
}

// retryAPICall retries read-only Kubernetes API calls when the failure is
// likely transient. Calls are attempted once plus defaultRetryPolicy.maxRetries times.
func retryAPICall[T any](ctx context.Context, operation string, call func(context.Context) (T, error)) (T, error) {
	return retryWithPolicy(ctx, operation, call, defaultRetryPolicy)
}

func retryWithPolicy[T any](ctx context.Context, operation string, call func(context.Context) (T, error), policy retryPolicy) (T, error) {
	delay := min(policy.initial, policy.maxDelay)

	for attempt := 0; ; attempt++ {
		result, err := call(ctx)
		if err == nil || attempt >= policy.maxRetries || !isTransientAPIError(err) {
			return result, err
		}

		log.Warn("Transient error during %s (attempt %d/%d); retrying in %s: %v",
			operation, attempt+1, policy.maxRetries+1, delay, err)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return result, fmt.Errorf("%s: %w (last error: %w)", operation, ctx.Err(), err)
		case <-timer.C:
		}

		delay = min(delay*apiRetryBackoffFactor, policy.maxDelay)
	}
}

// isTransientAPIError skips failures that client-go's rest client already
// retries for GETs (429 with Retry-After, connection resets and EOFs) to avoid
// multiplying retries. A 5xx carrying Retry-After may still be retried by both.
func isTransientAPIError(err error) bool {
	return k8serrors.IsServiceUnavailable(err) ||
		k8serrors.IsServerTimeout(err) ||
		k8serrors.IsTimeout(err) ||
		k8serrors.IsInternalError(err) ||
		utilnet.IsConnectionRefused(err) ||
		utilnet.IsTimeout(err)
}
