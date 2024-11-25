package types

import "fmt"

// Validate validates the total fees
func (tf TotalFees) Validate() error {
	if !tf.TotalPacketFee.IsValid() {
		return fmt.Errorf("invalid total packet fee: %s", tf.TotalPacketFee)
	}
	return nil
}
