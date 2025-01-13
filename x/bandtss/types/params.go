package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultInactivePenaltyDuration time.Duration = time.Minute * 10   // 10 minutes
	DefaultMinTransitionDuration   time.Duration = time.Hour * 24     // 1 days
	DefaultMaxTransitionDuration   time.Duration = time.Hour * 24 * 7 // 7 days
	// compute the bandtss reward following the allocation to Oracle. If the Oracle reward amounts to 70%,
	// the bandtss reward will be determined from the remaining 10%, which is 10% * 30% = 3%.
	DefaultRewardPercentage = uint64(10)
)

// DefaultFeePerSigner is the default value for the signing request fee per signer.
var DefaultFeePerSigner = sdk.NewCoins(sdk.NewInt64Coin("uband", 50))

// NewParams creates a new Params instance
func NewParams(
	rewardPercentage uint64,
	inactivePenaltyDuration time.Duration,
	minTransitionDuration time.Duration,
	maxTransitionDuration time.Duration,
	feePerSigner sdk.Coins,
) Params {
	return Params{
		RewardPercentage:        rewardPercentage,
		InactivePenaltyDuration: inactivePenaltyDuration,
		MinTransitionDuration:   minTransitionDuration,
		MaxTransitionDuration:   maxTransitionDuration,
		FeePerSigner:            feePerSigner,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(
		DefaultRewardPercentage,
		DefaultInactivePenaltyDuration,
		DefaultMinTransitionDuration,
		DefaultMaxTransitionDuration,
		DefaultFeePerSigner,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateTimeDuration("inactive penalty duration")(p.InactivePenaltyDuration); err != nil {
		return err
	}

	if err := validateTimeDuration("max transition duration")(p.MaxTransitionDuration); err != nil {
		return err
	}

	if err := validateTimeDuration("min transition duration")(p.MinTransitionDuration); err != nil {
		return err
	}

	if err := validateUint64("reward percentage", false)(p.RewardPercentage); err != nil {
		return err
	}

	// Validate fee
	if !p.FeePerSigner.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(p.FeePerSigner.String())
	}

	return nil
}

func validateUint64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}

func validateTimeDuration(name string) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(time.Duration)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}

		if v.Seconds() <= 0 {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
