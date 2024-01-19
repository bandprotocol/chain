package types

import "fmt"

func (p *Price) Validate() error {
	if p.Timestamp <= 0 {
		return fmt.Errorf("timestamp must be positive")
	}

	return nil
}
