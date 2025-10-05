package axnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/apperrors"
	"github.com/go-playground/validator/v10"
)

type (
	StructValidator struct {
		validator *validator.Validate
	}
	ValidationError struct {
		FailedField string
		Tag         string
		Value       any
	}
)

var structValidator = &StructValidator{
	validator: validator.New(),
}

// ValidateRequest validates a struct using the validator library and returns validation error
func ValidateRequest(ctx context.Context, req any) error {
	if err := structValidator.validate(ctx, req); err != nil {
		return errors.Join(apperrors.INVALID_BODY, err)
	}

	return nil
}

func (e ValidationError) Error() error {
	return errors.New(
		fmt.Sprintf(
			" %v: %#v | Needs to implement '%s'",
			e.FailedField,
			e.Value,
			e.Tag,
		),
	)
}

func (v StructValidator) validate(ctx context.Context, data any) error {
	validationErrors := []error{}

	errs := v.validator.StructCtx(ctx, data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			// In this case data object is actually holding the User struct
			var elem ValidationError

			elem.FailedField = err.Field() // Export struct field name
			elem.Tag = err.Tag()           // Export struct tag
			elem.Value = err.Value()       // Export field value

			validationErrors = append(validationErrors, elem.Error())
		}
		return errors.Join(validationErrors...)
	}

	return nil
}
