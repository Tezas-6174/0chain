// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// IteratorI is an autogenerated mock type for the IteratorI type
type IteratorI struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *IteratorI) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Scan provides a mock function with given fields: _a0
func (_m *IteratorI) Scan(_a0 ...interface{}) bool {
	var _ca []interface{}
	_ca = append(_ca, _a0...)
	ret := _m.Called(_ca...)

	var r0 bool
	if rf, ok := ret.Get(0).(func(...interface{}) bool); ok {
		r0 = rf(_a0...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
