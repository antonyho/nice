package nice

import (
	"reflect"
	"slices"
)

// Handler for the given artefact and error types
type Handler struct {
	artefactTypes []reflect.Type
	errorTypes    []error
}

// With takes a handle function from parameter
// and call the function while panic artfact type matches.
// The handle func does not catch panic from other level's goroutine.
func (h Handler) With(handle func(artefact any)) {
	if lastMsg := recover(); lastMsg != nil {
		switch asserted := lastMsg.(type) {
		case error:
			typeOfError := reflect.TypeFor[error]()
			// Handle general error registered
			if slices.Contains(h.artefactTypes, typeOfError) {
				handle(lastMsg)
				return
			}
			// Handle specific error registered
			if slices.Contains(h.errorTypes, asserted) {
				handle(lastMsg)
				return
			}
		default:
			typeOfLastMsg := reflect.TypeOf(lastMsg)
			if slices.Contains(h.artefactTypes, typeOfLastMsg) {
				handle(lastMsg)
				return
			}
		}

		// Fallthrough if not tackled
		panic(lastMsg) // This will ruin the call stack. Need a new solution.
	}
}

// Tackle panic with provided targets type
// returns a Handler, which shall be pairly used With().
// Pass exact error to the `targets`,
// if you want to handle particular type of error.
// Passsing `reflect.TypeFor[error]()` registers all types of error
// to be handled by the handle function.
func Tackle(targets ...any) Handler {
	artefactTypes := make([]reflect.Type, 0)
	errorTypes := make([]error, 0)

	for _, t := range targets {
		if errorType, matched := t.(error); matched {
			errorTypes = append(errorTypes, errorType)
			continue
		}
		if artefactType, matched := t.(reflect.Type); matched {
			artefactTypes = append(artefactTypes, artefactType)
		}
		// Unknown target is being ignored and is being discarded
	}

	return Handler{artefactTypes: artefactTypes, errorTypes: errorTypes}
}
