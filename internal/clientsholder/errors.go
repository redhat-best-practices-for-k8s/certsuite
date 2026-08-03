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

	k8sexec "k8s.io/client-go/util/exec"
)

type ExecError struct {
	Command   string
	Namespace string
	PodName   string
	ExitCode  int
	Err       error
}

func (e *ExecError) Error() string {
	if e.ExitCode >= 0 {
		return fmt.Sprintf("failed to execute command %q in %s/%s (exit code %d): %s",
			e.Command, e.Namespace, e.PodName, e.ExitCode, e.Err)
	}
	return fmt.Sprintf("failed to execute command %q in %s/%s: %s",
		e.Command, e.Namespace, e.PodName, e.Err)
}

func (e *ExecError) Unwrap() error {
	return e.Err
}

func (e *ExecError) HasExitCode(code int) bool {
	return e.ExitCode == code
}

func newExecError(command, namespace, podName string, err error) *ExecError {
	exitCode := -1
	var codeErr k8sexec.CodeExitError
	if errors.As(err, &codeErr) {
		exitCode = codeErr.Code
	}
	return &ExecError{
		Command:   command,
		Namespace: namespace,
		PodName:   podName,
		ExitCode:  exitCode,
		Err:       err,
	}
}
