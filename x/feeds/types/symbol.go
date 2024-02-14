package types

import "fmt"

func (s *Symbol) Validate() error {
	if s.Interval <= 0 {
		return fmt.Errorf("minInterval must be positive")
	}

	return nil
}
