package validator

import (
	"fmt"
	"strings"
)

type Validator struct {
	Errors map[string]string
}

// Creates a new validator with an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// returns true if the validator has errors, false otherwise
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// adds an error to the validator
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// if ok if false adds the error.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// if ok is false adds the error with a default message
func (v *Validator) CheckRequired(ok bool, key string) {
	v.Check(ok, key, "required field")
}

type ValidationError struct {
	Errors map[string]string
}

func (v *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation failed:")
	for field, msg := range v.Errors {
		sb.WriteString(fmt.Sprintf(" %s: %s;", field, msg))
	}
	return sb.String()
}
