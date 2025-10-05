package utility

import (
	"net/http"

	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/apperrors"
)

type PatientError uint

const (
	INVALID_BODY PatientError = iota
	INVALID_BODY_VALIDATION
	INVALID_LOCATION
	SCENARIO_NOT_FOUND
	DEVICE_NOT_FOUND
	MODULE_NOT_FOUND
	QUESTION_NOT_FOUND
	SCENARIO_VERSION_NOT_FOUND
	OTOKIOSK_NOT_FOUND
	OTOKIOSK_NOT_VALID
	PATIENT_ALREADY_EXIST
	INTEGRATION_ERROR
)

func (PatientError) Subject() string {
	return "VAL"
}

func (err PatientError) Code() uint {
	return uint(err)
}

func (err PatientError) StatusCode() int {
	return http.StatusBadRequest
}

func (err PatientError) Error() string {
	switch err {
	case INVALID_BODY:
		return "PersonalData or at least one result must be valued"
	case INVALID_BODY_VALIDATION:
		return "body must be valid"
	case INVALID_LOCATION:
		return "location code must be valid"
	case SCENARIO_NOT_FOUND:
		return "Scenario not found"
	case DEVICE_NOT_FOUND:
		return "Device not found"
	case MODULE_NOT_FOUND:
		return "Module not found"
	case QUESTION_NOT_FOUND:
		return "Question not found"
	case SCENARIO_VERSION_NOT_FOUND:
		return "Version not found"
	case OTOKIOSK_NOT_FOUND:
		return "Otokiosk not found"
	case OTOKIOSK_NOT_VALID:
		return "Otokiosk code User error"
	case PATIENT_ALREADY_EXIST:
		return "Patient ID already exist"
	case INTEGRATION_ERROR:
		return "Integration error"
	}

	return apperrors.GENERIC_ERROR.Error()
}
