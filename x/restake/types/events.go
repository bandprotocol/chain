package types

// restake module event types
const (
	EventTypeClaimRewards    = "claim_rewards"
	EventTypeUnlockPower     = "unlock_power"
	EventTypeLockPower       = "lock_power"
	EventTypeAddRewards      = "add_rewards"
	EventTypeDeactivateVault = "deactivate_vault"

	AttributeKeyStaker  = "staker"
	AttributeKeyKey     = "key"
	AttributeKeyRewards = "rewards"
	AttributeKeyPower   = "power"
)
