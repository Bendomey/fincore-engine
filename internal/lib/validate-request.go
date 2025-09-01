package lib

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func ValidateRequest(validate *validator.Validate, body interface{}, w http.ResponseWriter) bool {
	if err := validate.Struct(body); err != nil {
		errs := err.(validator.ValidationErrors)
		errorMessages := make([]string, len(errs))

		for i, e := range errs {
			errorMessages[i] = fmt.Sprintf("field '%s' failed validation rule '%s'", e.Field(), e.Tag())
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": errorMessages,
		})

		return false
	}

	return true
}
