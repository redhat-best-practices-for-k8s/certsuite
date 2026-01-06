// Copyright (C) 2020-2024 Red Hat, Inc.
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

import "sync"

// Ensure MockCommand implements Command interface
var _ Command = (*MockCommand)(nil)

// MockCommand is a mock implementation of the Command interface for testing.
// It allows tests to configure the behavior of ExecCommandContainer.
//
// Example usage:
//
//	mock := &MockCommand{
//	    ExecFunc: func(ctx Context, command string) (string, string, error) {
//	        return "output", "", nil
//	    },
//	}
type MockCommand struct {
	// ExecFunc is the function to call when ExecCommandContainer is invoked.
	// Tests can set this to return specific values for their test cases.
	ExecFunc func(ctx Context, command string) (stdout, stderr string, err error)

	// calls tracks all calls made to ExecCommandContainer for assertion purposes.
	calls []execCall
	mu    sync.RWMutex
}

type execCall struct {
	Context Context
	Command string
}

// ExecCommandContainer implements the Command interface.
func (m *MockCommand) ExecCommandContainer(ctx Context, command string) (stdout, stderr string, err error) {
	m.mu.Lock()
	m.calls = append(m.calls, execCall{Context: ctx, Command: command})
	m.mu.Unlock()

	if m.ExecFunc == nil {
		return "", "", nil
	}
	return m.ExecFunc(ctx, command)
}

// CallCount returns the number of times ExecCommandContainer was called.
func (m *MockCommand) CallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.calls)
}

// Calls returns all recorded calls to ExecCommandContainer.
func (m *MockCommand) Calls() []execCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]execCall{}, m.calls...)
}
