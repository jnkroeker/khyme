package validate

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func GenerateID() string {
	return uuid.NewString()
}
