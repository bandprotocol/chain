package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultInactivePenaltyDuration time.Duration = time.Minute * 10 // 10 minutes
	DefaultMaxTransitionDuration   time.Duration = time.Hour * 120  // 5 days
	// compute the bandtss reward following the allocation to Oracle. If the Oracle reward amounts to 70%,
	// the bandtss reward will be determined from the remaining 10%, which is 10% * 30% = 3%.
	DefaultRewardPercentage = uint64(10)
)

// DefaultFee is the default value for the signing request fee
// The value is ["10uband"]
var DefaultFee = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))

// NewParams creates a new Params instance
func NewParams(
	rewardPercentage uint64,
	inactivePenaltyDuration time.Duration,
	maxTransitionDuration time.Duration,
	fee sdk.Coins,
) Params {
	return Params{
		RewardPercentage:        rewardPercentage,
		InactivePenaltyDuration: inactivePenaltyDuration,
		MaxTransitionDuration:   maxTransitionDuration,
		Fee:                     fee,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(
		DefaultRewardPercentage,
		DefaultInactivePenaltyDuration,
		DefaultMaxTransitionDuration,
		DefaultFee,
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

	if err := validateUint64("reward percentage", false)(p.RewardPercentage); err != nil {
		return err
	}

	// Validate fee
	if !p.Fee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(p.Fee.String())
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
