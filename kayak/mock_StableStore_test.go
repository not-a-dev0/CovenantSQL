/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by mockery v1.0.0. DO NOT EDIT.
package kayak

import mock "github.com/stretchr/testify/mock"

// MockStableStore is an autogenerated mock type for the StableStore type
type MockStableStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: key
func (_m *MockStableStore) Get(key []byte) ([]byte, error) {
	ret := _m.Called(key)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUint64 provides a mock function with given fields: key
func (_m *MockStableStore) GetUint64(key []byte) (uint64, error) {
	ret := _m.Called(key)

	var r0 uint64
	if rf, ok := ret.Get(0).(func([]byte) uint64); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Set provides a mock function with given fields: key, val
func (_m *MockStableStore) Set(key []byte, val []byte) error {
	ret := _m.Called(key, val)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, []byte) error); ok {
		r0 = rf(key, val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetUint64 provides a mock function with given fields: key, val
func (_m *MockStableStore) SetUint64(key []byte, val uint64) error {
	ret := _m.Called(key, val)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, uint64) error); ok {
		r0 = rf(key, val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
