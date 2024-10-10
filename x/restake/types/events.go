package types

// restake module event types
const (
	EventTypeClaimRewards    = "claim_rewards"
	EventTypeLockPower       = "lock_power"
	EventTypeAddRewards      = "add_rewards"
	EventTypeDeactivateVault = "deactivate_vault"
	EventTypeStake           = "stake"
	EventTypeUnstake         = "unstake"

	AttributeKeyStaker  = "staker"
	AttributeKeyKey     = "key"
	AttributeKeyRewards = "rewards"
	AttributeKeyPower   = "power"
	AttributeKeyCoins   = "coins"
)
