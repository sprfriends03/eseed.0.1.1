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
			default:
				if values, ok := tags[e.Tag()]; ok {
					msgs[1] = fmt.Sprintf("must be one of (%v)", strings.Join(values, ","))
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
