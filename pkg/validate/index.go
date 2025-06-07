package validate

import (
	"app/pkg/enum"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

var tags = enum.Tags()

type StructValidator struct {
	validate *validator.Validate
}

func New() *StructValidator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(e reflect.StructField) string {
		return strings.Split(e.Tag.Get("json"), ",")[0]
	})
	for tag, values := range tags {
		validate.RegisterValidation(tag, func(e validator.FieldLevel) bool {
			return slices.Contains(values, e.Field().String())
		})
	}
	validate.RegisterValidation("regexp", func(e validator.FieldLevel) bool {
		return regexp.MustCompile(e.Param()).MatchString(e.Field().String())
	})

	// Register custom alphanum validator since built-in one isn't working reliably
	validate.RegisterValidation("alphanum", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		for _, char := range value {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				return false
			}
		}
		return len(value) > 0
	})

	// Register custom e164 validator since built-in one isn't working reliably
	validate.RegisterValidation("e164", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		// E164 format: +[1-9][0-9]{1,14}
		if len(value) < 2 || value[0] != '+' {
			return false
		}
		digits := value[1:]
		if len(digits) < 1 || len(digits) > 15 {
			return false
		}
		// First digit must be 1-9
		if digits[0] < '1' || digits[0] > '9' {
			return false
		}
		// Rest must be digits
		for i := 1; i < len(digits); i++ {
			if digits[i] < '0' || digits[i] > '9' {
				return false
			}
		}
		return true
	})

	return &StructValidator{validate}
}

func (s StructValidator) Engine() any {
	return s.validate
}

func (s StructValidator) Validate(out any) error {
	return s.ValidateStruct(out)
}

func (s StructValidator) ValidateStruct(out any) error {
	if err := s.validate.Struct(out); err != nil {
		results := make([]string, 0)
		for _, e := range err.(validator.ValidationErrors) {
			msgs := []string{fmt.Sprintf("Field '%v'", e.Field()), e.Tag()}
			switch e.Tag() {
			case "required":
				msgs[1] = "is required"
			case "min":
				msgs[1] = "minimum of"
			case "max":
				msgs[1] = "maximum of"
			case "len":
				msgs[1] = "must be length"
			case "unique":
				msgs[1] = "must be unique"
			case "lowercase":
				msgs[1] = "must be lowercase"
			case "uppercase":
				msgs[1] = "must be uppercase"
			case "regexp":
				msgs[1] = "must match pattern"
			case "alphanum":
				msgs[1] = "must contain only alphanumeric characters"
			case "e164":
				msgs[1] = "must be a valid E164 phone number format"
			case "email":
				msgs[1] = "must be a valid email address"
			default:
				if values, ok := tags[e.Tag()]; ok {
					msgs[1] = fmt.Sprintf("must be one of (%v)", strings.Join(values, ","))
				} else {
					// For unknown tags, show the tag name as-is
					msgs[1] = fmt.Sprintf("failed '%s' validation", e.Tag())
				}
			}
			if len(e.Param()) > 0 {
				msgs = append(msgs, e.Param())
			}
			results = append(results, strings.Join(msgs, " "))
		}
		return errors.New(strings.Join(results, ". "))
	}
	return nil
}
