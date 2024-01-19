package types

import "fmt"

func (s *Symbol) Validate() error {
	if s.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	return nil
}

func (us *UpdateSymbol) Validate() error {
	if us.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	return nil
}
