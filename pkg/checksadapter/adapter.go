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

package checksadapter

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
)

type Adapter struct {
	checkFunc checks.CheckFunc
}

func NewAdapter(checkFunc checks.CheckFunc) *Adapter {
	return &Adapter{
		checkFunc: checkFunc,
	}
}

func (a *Adapter) Execute(check *checksdb.Check, env *provider.TestEnvironment) error {
	resources := ConvertToDiscoveredResources(env)
	result := a.checkFunc(resources)
	ConvertAndSetResult(check, result)
	return nil
}

func (a *Adapter) ExecuteIntrusive(check *checksdb.Check, env *provider.TestEnvironment) error {
	defer env.SetNeedsRefresh()
	defer ResetCache()
	return a.Execute(check, env)
}

func (a *Adapter) MakeIntrusiveCheckFn(env *provider.TestEnvironment) func(*checksdb.Check) error {
	return func(check *checksdb.Check) error {
		return a.ExecuteIntrusive(check, env)
	}
}

func (a *Adapter) MakeCheckFn(env *provider.TestEnvironment) func(*checksdb.Check) error {
	return func(check *checksdb.Check) error {
		return a.Execute(check, env)
	}
}

// AddCheck creates a new check from the library and adds it to the group.
func AddCheck(group *checksdb.ChecksGroup, id string, checkFn checks.CheckFunc, env *provider.TestEnvironment, skips ...checksdb.SkipCheckFn) {
	group.Add(checksdb.NewCheck(GetCheckIDAndLabels(id)).
		WithSkipCheckFn(skips...).
		WithCheckFn(NewAdapter(checkFn).MakeCheckFn(env)))
}

// AddIntrusiveCheck is like AddCheck but uses ExecuteIntrusive.
func AddIntrusiveCheck(group *checksdb.ChecksGroup, id string, checkFn checks.CheckFunc, env *provider.TestEnvironment, skips ...checksdb.SkipCheckFn) {
	group.Add(checksdb.NewCheck(GetCheckIDAndLabels(id)).
		WithSkipCheckFn(skips...).
		WithCheckFn(NewAdapter(checkFn).MakeIntrusiveCheckFn(env)))
}
