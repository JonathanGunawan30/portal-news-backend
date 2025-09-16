package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" || tag == "" {
			return fld.Name
		}
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	})
}

func ValidateStruct(s any) error {
	if s == nil {
		return errors.New("payload must not be empty")
	}

	if err := validate.Struct(s); err == nil {
		return nil
	} else if _, ok := err.(*validator.InvalidValidationError); ok {
		return fmt.Errorf("internal validation error: %w", err)
	} else {
		var msgs []string
		for _, fe := range err.(validator.ValidationErrors) {
			field := fe.Field()

			switch fe.Tag() {
			case "required":
				msgs = append(msgs, fmt.Sprintf("Field `%s` is required.", field))
			case "email":
				msgs = append(msgs, "Invalid email format.")
			case "min":
				if strings.EqualFold(field, "password") {
					msgs = append(msgs, fmt.Sprintf("Password must be at least %s characters.", fe.Param()))
				} else {
					msgs = append(msgs, fmt.Sprintf("Field `%s` must be at least %s.", field, fe.Param()))
				}
			case "max":
				msgs = append(msgs, fmt.Sprintf("Field `%s` must be at most %s.", field, fe.Param()))
			case "oneof":
				msgs = append(msgs, fmt.Sprintf("Field `%s` must be one of: %s.", field, fe.Param()))
			case "len":
				msgs = append(msgs, fmt.Sprintf("Field `%s` must be exactly %s characters.", field, fe.Param()))
			case "gte":
				msgs = append(msgs, fmt.Sprintf("Field `%s` must be greater than or equal to %s.", field, fe.Param()))
			case "lte":
				msgs = append(msgs, fmt.Sprintf("Field `%s` must be less than or equal to %s.", field, fe.Param()))
			default:
				msgs = append(msgs, fmt.Sprintf("Field `%s` is invalid.", field))
			}
		}

		return errors.New(strings.Join(msgs, " "))
	}
}
