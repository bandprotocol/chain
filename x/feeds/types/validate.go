package types

import (
	"fmt"
	"net/url"
)

// validateInt64 validates int64 and check its positivity.
func validateInt64(name string, positiveOnly bool, i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 && positiveOnly {
		return fmt.Errorf("%s must be positive: %d", name, v)
	}
	return nil
}

// validateURL validates URL format.
func validateURL(name string, u string) error {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("%s has invalid URL format", name)
	}
	return nil
}

// validateString validates string.
func validateString(name string, allowEmpty bool, i interface{}) error {
	s, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if s == "" && !allowEmpty {
		return fmt.Errorf("%s cannot be empty", name)
	}
	return nil
}
