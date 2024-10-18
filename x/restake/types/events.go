package types

// restake module event types
const (
	EventTypeLockPower       = "lock_power"
	EventTypeCreateVault     = "create_vault"
	EventTypeDeactivateVault = "deactivate_vault"
	EventTypeStake           = "stake"
	EventTypeUnstake         = "unstake"

	AttributeKeyStaker       = "staker"
	AttributeKeyKey          = "key"
	AttributeKeyVaultAddress = "vault_address"
	AttributeKeyPower        = "power"
	AttributeKeyCoins        = "coins"
)
