package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Total returns the total fees
func (tf TotalFees) Total() sdk.Coins {
	return tf.TotalBasePacketFee
}

// Validate validates the total fees
func (tf TotalFees) Validate() error {
	if !tf.TotalBasePacketFee.IsValid() {
		return fmt.Errorf("invalid total packet fee: %s", tf.TotalBasePacketFee)
	}
	return nil
}
