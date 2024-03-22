package types

import "fmt"

func (f *Feed) Validate() error {
	if f.Interval <= 0 {
		return fmt.Errorf("minInterval must be positive")
	}

	return nil
}
