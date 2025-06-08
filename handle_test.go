package nice_test

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/antonyho/nice"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	Executed bool
}

func (m *mockHandler) Handle(artefact any) {
	m.Executed = true
	log.Printf("It panicked. Error: %+v", artefact)
}

func assertExecuted(t *testing.T, h *mockHandler) {
	assert.True(t, h.Executed, "The handler function have been executed.")
}
func assertNotExecuted(t *testing.T, h *mockHandler) {
	assert.False(t, h.Executed, "The handler function have not been executed yet.")
}

func TestHandler(t *testing.T) {
	t.Run("handle", func(t *testing.T) {
		mockHandler := &mockHandler{Executed: false}
		defer assertExecuted(t, mockHandler)

		// When generic error was registered, all types of error will be handled.
		// Any other error registered to the same handler is not checked.
		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler.Handle)

		customErr := errors.New("custom error")
		panicFunc := func() {
			panic(customErr)
		}
		panicFunc()

		// Output: It panicked. Error: custom error
	})

	t.Run("handle different artefact types with same handler", func(t *testing.T) {
		mockHandler := &mockHandler{Executed: false}
		defer assertExecuted(t, mockHandler)

		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler.Handle)
		defer nice.Tackle(reflect.TypeFor[string]()).With(mockHandler.Handle)

		panicFunc := func() {
			panic("error message string")
		}
		panicFunc()

		// Output: It panicked. Error: error message string
	})

	t.Run("handle different artefact types with different handlers", func(t *testing.T) {
		mockHandler4Str := &mockHandler{Executed: false}
		mockHandler4Err := &mockHandler{Executed: false}
		defer assertNotExecuted(t, mockHandler4Err)
		defer assertExecuted(t, mockHandler4Str)

		defer nice.Tackle(reflect.TypeFor[string]()).With(mockHandler4Str.Handle)
		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler4Err.Handle)

		panicFunc := func() {
			panic("error message string")
		}
		panicFunc()

		// Output: It panicked. Error: error message string
	})

	// The later registered handler should be executed,
	// because of the execution stack sequence of deferred function.
	t.Run("register same artefact type with multiple handlers", func(t *testing.T) {
		mockHandler1st := &mockHandler{Executed: false}
		mockHandler2nd := &mockHandler{Executed: false}
		defer assertNotExecuted(t, mockHandler1st)
		defer assertExecuted(t, mockHandler2nd)

		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler1st.Handle)
		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler2nd.Handle)

		panicFunc := func() {
			panic(errors.New("mock error"))
		}
		panicFunc()

		// Output: It panicked. Error: mock error
	})

	t.Run("handle exact error type", func(t *testing.T) {
		customErrorExpected := errors.New("expected error")
		customErrorUnexpected := errors.New("unexpected error")

		mockHandler4Expected := &mockHandler{Executed: false}
		mockHandler4Unexpected := &mockHandler{Executed: false}
		defer assertExecuted(t, mockHandler4Expected)
		defer assertNotExecuted(t, mockHandler4Unexpected)

		// Do not register string error by type, they cannot be distinguished by type.
		// They are all `*errors.stringError`.
		defer nice.Tackle(customErrorExpected).With(mockHandler4Expected.Handle)
		defer nice.Tackle(customErrorUnexpected).With(mockHandler4Unexpected.Handle)

		panicFunc := func() {
			panic(customErrorExpected)
		}
		panicFunc()

		// Output: It panicked. Error: expected error
	})

	t.Run("no matched artefact type", func(t *testing.T) {
		mockHandler := &mockHandler{Executed: false}
		defer assertNotExecuted(t, mockHandler)
		defer func() {
			if artefact := recover(); artefact == nil {
				t.Error("Unhandled panic did not fallthrough.")
			}
		}()

		defer nice.Tackle(reflect.TypeFor[error]()).With(mockHandler.Handle)
		defer nice.Tackle(reflect.TypeFor[string]()).With(mockHandler.Handle)

		panicFunc := func() {
			panic(7)
		}
		panicFunc()

		// Output:
	})

	t.Run("no further execution after panic was captured", func(t *testing.T) {
		mockErr1st := errors.New("mock error 1st")
		mockErr2nd := errors.New("mock error 2nd")
		mockHandler1st := &mockHandler{Executed: false}
		mockHandler2nd := &mockHandler{Executed: false}
		defer assertExecuted(t, mockHandler1st)
		defer assertNotExecuted(t, mockHandler2nd)

		defer nice.Tackle(mockErr1st).With(mockHandler1st.Handle)
		defer nice.Tackle(mockErr2nd).With(mockHandler2nd.Handle)

		panic(mockErr1st)
		panic(mockErr2nd) //nolint: unreachable

		// Output: It panicked. Error: mock error
	})
}

func ExampleTackle() {

	var customError = &struct {
		CustomMessage string
		Error         func() string
	}{
		CustomMessage: "error: custom message",
	}
	customError.Error = func() string {
		return customError.CustomMessage
	}
	customStrErr := errors.New("error: custom string")
	strErr := "error: string only"

	handleFunc := func(artefact any) {
		fmt.Printf("It panicked. Error: %+v", artefact)
	}

	defer nice.Tackle(
		customStrErr,
		reflect.TypeOf(customError),
		strErr,
	).With(handleFunc)

	panicFunc := func() {
		panic(customStrErr)
	}
	panicFunc()
	// Output: It panicked. Error: error: custom string
}
