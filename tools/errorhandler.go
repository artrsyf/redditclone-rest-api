package tools

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
)

func JSONError(w http.ResponseWriter, status int, msg string, method string) {
	defer func() {
		if r := recover(); r != nil {
			Logger.WithFields(logrus.Fields{
				"method": method,
				"status": status,
				"panic":  r,
			}).Error("panic occurred")
		}
	}()

	Logger.WithFields(logrus.Fields{
		"method": method,
		"status": status,
	}).Error(msg)

	w.WriteHeader(status)

	resp, err := json.Marshal(map[string]interface{}{
		"status": status,
		"error":  msg,
	})
	if err != nil {
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		return
	}
}

func ValidationError(w http.ResponseWriter, validationError error) {
	if allErrs, ok := validationError.(govalidator.Errors); ok {
		for _, wrongField := range allErrs.Errors() {
			message := fmt.Sprintf("field: %#v\n\n", wrongField)
			JSONError(w, http.StatusUnprocessableEntity, message, "errorresponses.ValidationError")
		}
	}
}
