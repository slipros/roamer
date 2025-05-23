// Code generated by mockery. DO NOT EDIT.

package mockroamer

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// Decoder is an autogenerated mock type for the Decoder type
type Decoder struct {
	mock.Mock
}

type Decoder_Expecter struct {
	mock *mock.Mock
}

func (_m *Decoder) EXPECT() *Decoder_Expecter {
	return &Decoder_Expecter{mock: &_m.Mock}
}

// ContentType provides a mock function with no fields
func (_m *Decoder) ContentType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ContentType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Decoder_ContentType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContentType'
type Decoder_ContentType_Call struct {
	*mock.Call
}

// ContentType is a helper method to define mock.On call
func (_e *Decoder_Expecter) ContentType() *Decoder_ContentType_Call {
	return &Decoder_ContentType_Call{Call: _e.mock.On("ContentType")}
}

func (_c *Decoder_ContentType_Call) Run(run func()) *Decoder_ContentType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Decoder_ContentType_Call) Return(_a0 string) *Decoder_ContentType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Decoder_ContentType_Call) RunAndReturn(run func() string) *Decoder_ContentType_Call {
	_c.Call.Return(run)
	return _c
}

// Decode provides a mock function with given fields: r, ptr
func (_m *Decoder) Decode(r *http.Request, ptr interface{}) error {
	ret := _m.Called(r, ptr)

	if len(ret) == 0 {
		panic("no return value specified for Decode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, interface{}) error); ok {
		r0 = rf(r, ptr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Decoder_Decode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Decode'
type Decoder_Decode_Call struct {
	*mock.Call
}

// Decode is a helper method to define mock.On call
//   - r *http.Request
//   - ptr interface{}
func (_e *Decoder_Expecter) Decode(r interface{}, ptr interface{}) *Decoder_Decode_Call {
	return &Decoder_Decode_Call{Call: _e.mock.On("Decode", r, ptr)}
}

func (_c *Decoder_Decode_Call) Run(run func(r *http.Request, ptr interface{})) *Decoder_Decode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request), args[1].(interface{}))
	})
	return _c
}

func (_c *Decoder_Decode_Call) Return(_a0 error) *Decoder_Decode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Decoder_Decode_Call) RunAndReturn(run func(*http.Request, interface{}) error) *Decoder_Decode_Call {
	_c.Call.Return(run)
	return _c
}

// NewDecoder creates a new instance of Decoder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDecoder(t interface {
	mock.TestingT
	Cleanup(func())
}) *Decoder {
	mock := &Decoder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
