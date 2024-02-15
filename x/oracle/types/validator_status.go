package types

import "time"

func NewValidatorStatus(
	isActive bool,
	since time.Time,
) ValidatorStatus {
	return ValidatorStatus{
		IsActive: isActive,
		Since:    since,
	}
}
