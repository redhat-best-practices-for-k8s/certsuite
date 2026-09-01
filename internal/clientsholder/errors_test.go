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

package clientsholder

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	k8sexec "k8s.io/client-go/util/exec"
)

func TestExecErrorMessage(t *testing.T) {
	t.Run("with exit code", func(t *testing.T) {
		e := &ExecError{
			Command:   "ls /tmp",
			Namespace: "test-ns",
			PodName:   "test-pod",
			ExitCode:  125,
			Err:       errors.New("container not running"),
		}
		assert.Contains(t, e.Error(), "exit code 125")
		assert.Contains(t, e.Error(), "ls /tmp")
		assert.Contains(t, e.Error(), "test-ns/test-pod")
	})

	t.Run("with exit code zero", func(t *testing.T) {
		e := &ExecError{
			Command:   "true",
			Namespace: "ns",
			PodName:   "pod",
			ExitCode:  0,
			Err:       errors.New("stream error"),
		}
		assert.Contains(t, e.Error(), "exit code 0")
	})

	t.Run("without exit code", func(t *testing.T) {
		e := &ExecError{
			Command:   "ls /tmp",
			Namespace: "test-ns",
			PodName:   "test-pod",
			ExitCode:  -1,
			Err:       errors.New("connection refused"),
		}
		assert.NotContains(t, e.Error(), "exit code")
		assert.Contains(t, e.Error(), "connection refused")
	})

	t.Run("with nil error", func(t *testing.T) {
		e := &ExecError{
			Command:   "test",
			Namespace: "ns",
			PodName:   "pod",
			ExitCode:  -1,
			Err:       nil,
		}
		assert.Contains(t, e.Error(), "test")
		assert.Nil(t, e.Unwrap())
	})
}

func TestExecErrorHasExitCode(t *testing.T) {
	e := &ExecError{ExitCode: 125}
	assert.True(t, e.HasExitCode(125))
	assert.False(t, e.HasExitCode(1))
	assert.False(t, e.HasExitCode(-1))
}

func TestExecErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	e := &ExecError{
		Command:   "test",
		Namespace: "ns",
		PodName:   "pod",
		ExitCode:  -1,
		Err:       inner,
	}
	assert.ErrorIs(t, e, inner)

	var target *ExecError
	wrapped := fmt.Errorf("outer: %w", e)
	assert.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "test", target.Command)
}

func TestNewExecErrorWithCodeExitError(t *testing.T) {
	codeErr := k8sexec.CodeExitError{
		Err:  errors.New("command terminated with exit code 125"),
		Code: 125,
	}
	e := newExecError("podman diff", "test-ns", "test-pod", codeErr)
	assert.Equal(t, 125, e.ExitCode)
	assert.True(t, e.HasExitCode(125))
	assert.Contains(t, e.Error(), "exit code 125")
}

func TestNewExecErrorWithPlainError(t *testing.T) {
	e := newExecError("ls", "test-ns", "test-pod", errors.New("connection refused"))
	assert.Equal(t, -1, e.ExitCode)
	assert.False(t, e.HasExitCode(125))
	assert.NotContains(t, e.Error(), "exit code")
}

func TestNewExecErrorWithWrappedCodeExitError(t *testing.T) {
	codeErr := k8sexec.CodeExitError{
		Err:  errors.New("command terminated with exit code 1"),
		Code: 1,
	}
	wrapped := fmt.Errorf("stream failed: %w", codeErr)
	e := newExecError("cat /etc/hosts", "ns", "pod", wrapped)
	assert.Equal(t, 1, e.ExitCode)
	assert.True(t, e.HasExitCode(1))
}
