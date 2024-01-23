package types

import "fmt"

func (s *Symbol) Validate() error {
	if s.MinInterval <= 0 {
		return fmt.Errorf("minInterval must be positive")
	}

	if s.MaxInterval <= 0 {
		return fmt.Errorf("maxInterval must be positive")
	}

	if s.MinInterval >= s.MaxInterval {
		return fmt.Errorf("maxInterval must be more than minInterval")
	}

	return nil
}

func (us *UpdateSymbol) Validate() error {
	if us.MinInterval <= 0 {
		return fmt.Errorf("minInterval must be positive")
	}

	if us.MaxInterval <= 0 {
		return fmt.Errorf("maxInterval must be positive")
	}

	if us.MinInterval >= us.MaxInterval {
		return fmt.Errorf("maxInterval must be more than minInterval")
	}

	return nil
}
