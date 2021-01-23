package validators

import (
	"github.com/go-playground/validator/v10"
)

// Validate validates user input, based on the given model.
func Validate(model interface{}) (bool, error) {
	validate := validator.New()
	err := validate.Struct(model)
	if err != nil {
		return false, err
	}
	return true, nil
}
