package types

import (
	"fmt"
	"time"
)

const (
	DefaultActiveDuration          time.Duration = time.Hour * 24      // 1 days
	DefaultInactivePenaltyDuration time.Duration = time.Minute * 10    // 10 minutes
	DefaultJailPenaltyDuration     time.Duration = time.Hour * 24 * 30 // 30 days
	// compute the BandTSS reward following the allocation to Oracle. If the Oracle reward amounts to 40%,
	// the BandTSS reward will be determined from the remaining 60%.
	DefaultRewardPercentage = uint64(50)
)

// NewParams creates a new Params instance
func NewParams(
	activeDuration time.Duration,
	rewardPercentage uint64,
	inactivePenaltyDuration time.Duration,
	jailPenaltyDuration time.Duration,
) Params {
	return Params{
		ActiveDuration:          activeDuration,
		RewardPercentage:        rewardPercentage,
		InactivePenaltyDuration: inactivePenaltyDuration,
		JailPenaltyDuration:     jailPenaltyDuration,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		ActiveDuration:          DefaultActiveDuration,
		RewardPercentage:        DefaultRewardPercentage,
		InactivePenaltyDuration: DefaultInactivePenaltyDuration,
		JailPenaltyDuration:     DefaultJailPenaltyDuration,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateTimeDuration("active duration")(p.ActiveDuration); err != nil {
		return err
	}

	if err := validateTimeDuration("inactive penalty duration")(p.InactivePenaltyDuration); err != nil {
		return err
	}

	if err := validateTimeDuration("jail penalty duration")(p.JailPenaltyDuration); err != nil {
		return err
	}

	if err := validateUint64("reward percentage", false)(p.RewardPercentage); err != nil {
		return err
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
