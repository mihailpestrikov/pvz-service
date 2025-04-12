package dto

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationResponse struct {
	Errors []ValidationError `json:"errors"`
}

func ValidateRequest(data interface{}) []ValidationError {
	var errors []ValidationError

	err := Validator.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.Field()

			switch err.Tag() {
			case "required":
				element.Message = fmt.Sprintf("Field %s is required", err.Field())
			case "email":
				element.Message = fmt.Sprintf("Field %s must be a valid email address", err.Field())
			case "min":
				element.Message = fmt.Sprintf("Field %s must contain at least %s characters", err.Field(), err.Param())
			case "max":
				element.Message = fmt.Sprintf("Field %s must contain at most %s characters", err.Field(), err.Param())
			case "oneof":
				element.Message = fmt.Sprintf("Field %s must be one of: %s", err.Field(), err.Param())
			case "uuid4":
				element.Message = fmt.Sprintf("Field %s must be a valid UUID v4", err.Field())
			case "gtfield":
				element.Message = fmt.Sprintf("Field %s must be greater than %s", err.Field(), err.Param())
			default:
				element.Message = fmt.Sprintf("Field %s failed validation: %s", err.Field(), err.Tag())
			}

			errors = append(errors, element)
		}
	}

	return errors
}

func RespondWithValidationErrors(w http.ResponseWriter, errors []ValidationError) {
	response := ValidationResponse{
		Errors: errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Validation error", http.StatusBadRequest)
	}
}

func ValidateAndRespond(w http.ResponseWriter, data interface{}) bool {
	errors := ValidateRequest(data)
	if len(errors) > 0 {
		RespondWithValidationErrors(w, errors)
		return false
	}
	return true
}
