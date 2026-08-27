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

package crclient

import (
	"errors"
	"fmt"
	"testing"
)

func TestProbeToolExecError_Error(t *testing.T) {
	wrapped := errors.New("unable to upgrade connection: container not found")
	err := &ProbeToolExecError{Reason: "probe pod unavailable", Wrapped: wrapped}
	got := err.Error()
	if got != "probe tool execution error: probe pod unavailable: unable to upgrade connection: container not found" {
		t.Errorf("unexpected Error() string: %q", got)
	}
}

func TestProbeToolExecError_Unwrap(t *testing.T) {
	wrapped := errors.New("context deadline exceeded")
	err := &ProbeToolExecError{Reason: "timeout", Wrapped: wrapped}
	if !errors.Is(err, wrapped) {
		t.Error("Unwrap should expose the wrapped error via errors.Is")
	}
}

func TestIsProbeToolExecError_direct(t *testing.T) {
	err := &ProbeToolExecError{Reason: "test", Wrapped: errors.New("inner")}
	if !IsProbeToolExecError(err) {
		t.Error("expected IsProbeToolExecError to return true for a direct ProbeToolExecError")
	}
}

func TestIsProbeToolExecError_wrapped(t *testing.T) {
	inner := &ProbeToolExecError{Reason: "test", Wrapped: errors.New("inner")}
	outer := fmt.Errorf("netutil wrapper: %w", inner)
	if !IsProbeToolExecError(outer) {
		t.Error("expected IsProbeToolExecError to return true when ProbeToolExecError is wrapped by fmt.Errorf with %%w")
	}
}

func TestIsProbeToolExecError_plain(t *testing.T) {
	err := errors.New("some other error")
	if IsProbeToolExecError(err) {
		t.Error("expected IsProbeToolExecError to return false for a plain error")
	}
}

func TestIsProbeToolExecError_nil(t *testing.T) {
	if IsProbeToolExecError(nil) {
		t.Error("expected IsProbeToolExecError to return false for nil")
	}
}
