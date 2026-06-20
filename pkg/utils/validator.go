package utils

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// ValidationError is the shape returned inside the 400 error array.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

func getValidator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New()

		// Register tag name function
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// Mengambil tag "json"
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

			// Jika tag json bernilai "-", kita kembalikan string kosong (default)
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return validate
}

// ValidateStruct validates a struct and returns a slice of ValidationError.
// Returns nil if the struct is valid.
func ValidateStruct(s interface{}) []ValidationError {
	err := getValidator().Struct(s)
	if err == nil {
		return nil
	}

	var errs []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   e.Field(),
			Message: buildMessage(e),
		})
	}
	return errs
}

func buildMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", e.Field(), e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", e.Field(), e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", e.Field(), e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s is invalid (%s)", e.Field(), e.Tag())
	}
}
