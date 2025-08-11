package schema

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

func RegisterAllValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("zone_name", validatorZoneName); err != nil {
		return err
	}
	if err := v.RegisterValidation("subdomain_part", validatorSubdomainPart); err != nil {
		return err
	}
	return nil
}

func validatorZoneName(fl validator.FieldLevel) bool {
	zn, ok := fl.Field().Interface().(string)

	if !ok {
		return false
	}

	if len(zn) == 0 {
		return false
	} else if len(zn) == 1 {
		return zn == "."
	} else {
		return !strings.HasPrefix(zn, ".") && strings.HasSuffix(zn, ".")
	}
}

func validatorSubdomainPart(fl validator.FieldLevel) bool {
	sp, ok := fl.Field().Interface().(string)

	if !ok {
		return false
	}

	if len(sp) == 0 {
		return false
	} else if len(sp) == 1 {
		return sp == "."
	} else {
		return strings.HasPrefix(sp, ".") && strings.HasSuffix(sp, ".")
	}
}
