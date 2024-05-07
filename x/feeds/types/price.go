package types

import "fmt"

// Validate validates a Price.
func (p *Price) Validate() error {
	if p.Timestamp <= 0 {
		return fmt.Errorf("timestamp must be positive")
	}

	return nil
}
