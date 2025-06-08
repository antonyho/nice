package nice

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTackle(t *testing.T) {
	t.Run("Single artefact yype", func(t *testing.T) {
		h := Tackle(reflect.TypeFor[string]())
		expected := Handler{
			artefactTypes: []reflect.Type{reflect.TypeFor[string]()},
			errorTypes:    []error{},
		}
		assert.Equal(t, expected, h)
	})

	t.Run("Single generic error type", func(t *testing.T) {
		h := Tackle(reflect.TypeFor[error]())
		expected := Handler{
			artefactTypes: []reflect.Type{reflect.TypeFor[error]()},
			errorTypes:    []error{},
		}
		assert.Equal(t, expected, h)
	})

	t.Run("Single custom string error type", func(t *testing.T) {
		customStringError := errors.New("error: custom")

		h := Tackle(customStringError)
		expected := Handler{
			artefactTypes: []reflect.Type{},
			errorTypes:    []error{customStringError},
		}
		assert.Equal(t, expected, h)
	})

	t.Run("Multiple Artefact Types", func(t *testing.T) {
		customStringError := errors.New("error: custom")

		h := Tackle(
			reflect.TypeFor[string](),
			reflect.TypeFor[error](),
			customStringError,
		)

		expectedArtefactTypes := []reflect.Type{
			reflect.TypeFor[string](),
			reflect.TypeFor[error](),
		}
		expectedErrorTypes := []error{customStringError}
		expected := Handler{
			artefactTypes: expectedArtefactTypes,
			errorTypes:    expectedErrorTypes,
		}

		assert.Equal(t, expected, h)
	})
}
