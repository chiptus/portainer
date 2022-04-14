package validate

import (
	"github.com/go-playground/validator/v10"
	portaineree "github.com/portainer/portainer-ee/api"
)

func registerValidationMethods(v *validator.Validate) {
	v.RegisterValidation("validate_urls", ValidateLDAPURLs)
	v.RegisterValidation("validate_bool", ValidateBool)
}

/**
 * Validation methods below are being used for custom validation
 */

// ValidateLDAPURLs validates hostname_port validation for slices
func ValidateLDAPURLs(fl validator.FieldLevel) bool {
	urls := fl.Field().Interface().([]string)
	ldp := fl.Parent().Interface().(portaineree.LDAPSettings)

	if ldp.URL != "" && len(urls) == 0 {
		return false
	}

	validate = validator.New()

	for _, url := range urls {
		err := validate.Var(url, "hostname_port")
		return err == nil
	}
	return true
}

// ValidateBool validate boolean fields
func ValidateBool(fl validator.FieldLevel) bool {
	_, ok := fl.Field().Interface().(bool)
	return ok
}
