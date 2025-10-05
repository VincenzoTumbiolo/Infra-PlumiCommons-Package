package axnet

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/apperrors"
	"github.com/go-playground/assert/v2"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Num   int    `validate:"required"`
}

func TestValidator(t *testing.T) {
	invalidStruct := TestStruct{
		Email: "asdf",
		Name:  "",
	}

	err := ValidateRequest(context.Background(), invalidStruct)
	if err == nil {
		t.Error("Validation should fail for invalid struct")
		return
	}
	fmt.Println(err.Error())
	assert.Equal(t, errors.Is(err, apperrors.INVALID_BODY), true)
}
