// Copyright (C) 2020-2022 Red Hat, Inc.
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

package testhelper

import (
	"fmt"
	"reflect"
)

const (
	SUCCESS = iota
	FAILURE
	ERROR
)

func ResultToString(result int) (str string) {
	switch result {
	case SUCCESS:
		return "SUCCESS" //nolint:goconst
	case FAILURE:
		return "FAILURE" //nolint:goconst
	case ERROR:
		return "ERROR" //nolint:goconst
	}
	return ""
}

func SkipIfEmptyAny(skip func(string, ...int), object ...interface{}) {
	for _, o := range object {
		s := reflect.ValueOf(o)
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}

		if s.Len() == 0 {
			skip(fmt.Sprintf("Test skipped because there are no %s to test, please check under test labels", reflect.TypeOf(o)))
		}
	}
}

func SkipIfEmptyAll(skip func(string, ...int), object ...interface{}) {
	countLenZero := 0
	allTypes := ""
	for _, o := range object {
		s := reflect.ValueOf(o)
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}

		if s.Len() == 0 {
			countLenZero++
			allTypes = allTypes + reflect.TypeOf(o).String() + ", "
		}
	}
	// all objects have len() of 0
	if countLenZero == len(object) {
		skip(fmt.Sprintf("Test skipped because there are no %s to test, please check under test labels", allTypes))
	}
}

func AddTestResultLog(prefix string, object interface{}, log func(string, ...interface{}), fail func(string, ...int)) {
	s := reflect.ValueOf(object)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
		panic("AddTestResultLog object param is a non slice/map type")
	}
	if s.Len() > 0 {
		log(fmt.Sprintf("%s %s: %v", prefix, reflect.TypeOf(object), object))
		fail(fmt.Sprintf("Number of %s %s = %d", prefix, reflect.TypeOf(object), s.Len()))
	}
}
