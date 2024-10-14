package types

// restake module event types
const (
	EventTypeClaimRewards    = "claim_rewards"
	EventTypeLockPower       = "lock_power"
	EventTypeAddRewards      = "add_rewards"
	EventTypeCreateVault     = "create_vault"
	EventTypeDeactivateVault = "deactivate_vault"

	AttributeKeyStaker       = "staker"
	AttributeKeyKey          = "key"
	AttributeKeyVaultAddress = "vault_address"
	AttributeKeyRewards      = "rewards"
	AttributeKeyPower        = "power"
)
