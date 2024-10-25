package types

import (
	"fmt"
	"net/url"

	"github.com/Masterminds/semver/v3"
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

// validateUint64 validates uint64.
func validateUint64(name string, positiveOnly bool, i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 && positiveOnly {
		return fmt.Errorf("%s must be positive: %d", name, v)
	}

	return nil
}

// validateURL validates URL format.
func validateURL(name string, u string) error {
	if _, err := url.ParseRequestURI(u); err != nil {
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

// validateVersion checks if the version string is valid according to Semantic Versioning
func validateVersion(name string, version string) error {
	if version == "[NOT_SET]" {
		return nil
	}
	_, err := semver.StrictNewVersion(version)
	if err != nil {
		return fmt.Errorf("%s is not in a valid version format", name)
	}
	return nil
}
