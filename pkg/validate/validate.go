// Package validate wraps go-playground/validator so handlers can turn struct
// tags (`validate:"required,email"`) into a clean field->message map that maps
// straight onto the httpx error envelope and, ultimately, the form fields in
// the React frontend.
package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator is a thread-safe, reusable validator instance.
type Validator struct {
	v *validator.Validate
}

// New builds a Validator that reports errors keyed by the struct's JSON field
// names (not the Go field names), so the keys match exactly what the client
// sent.
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{v: v}
}

// Struct validates s. It returns nil when valid, or a map of
// fieldName -> humanMessage when not.
func (val *Validator) Struct(s any) map[string]string {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}

	var invalid validator.ValidationErrors
	if !asValidationErrors(err, &invalid) {
		return map[string]string{"_": err.Error()}
	}

	fields := make(map[string]string, len(invalid))
	for _, fe := range invalid {
		fields[fe.Field()] = message(fe)
	}
	return fields
}

func asValidationErrors(err error, dst *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*dst = ve
		return true
	}
	return false
}

// message turns a single validation failure into a friendly sentence.
func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "eqfield":
		return "must match " + fe.Param()
	default:
		return "is not valid"
	}
}
