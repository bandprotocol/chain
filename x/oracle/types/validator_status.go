package types

import "time"

func NewValidatorStatus(
	IsActive bool,
	Since time.Time,
) ValidatorStatus {
	return ValidatorStatus{
		IsActive: IsActive,
		Since:    Since,
	}
}
