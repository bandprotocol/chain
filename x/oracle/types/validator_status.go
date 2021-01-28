package types

import time "time"

func NewValidatorStatus(
	IsActive bool,
	Since time.Time,
) ValidatorStatus {
	return ValidatorStatus{
		IsActive: IsActive,
		Since:    Since,
	}
}
