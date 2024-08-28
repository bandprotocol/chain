# `x/restake`

## Abstract

This document specifies the Restake module. The `restake` module provides a mechanism for locking and distributing rewards within a blockchain.  

![Untitled](https://github.com/user-attachments/assets/305e449a-713a-46f4-9996-ddc5e00ede92)

This module is used in the BandChain.

## Contents

- [`x/restake`](#xrestake)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
  - [State](#state)
    - [Vault](#vault)
    - [Lock](#lock)
      - [LocksByPowerIndex](#locksbypowerindex)
    - [Stake](#stake)
    - [Params](#params)
  - [Messages](#messages)
    - [MsgClaimRewards](#msgclaimrewards)
    - [MsgStake](#msgstake)
    - [MsgUnstake](#msgunstake)
    - [MsgUpdateParams](#msgupdateparams)
  - [Staking hooks](#staking-hooks)
  - [Expected keepers](#expected-keepers)
    - [SetLockedPower](#setlockedpower)
    - [GetLockedPower](#getlockedpower)
    - [AddRewards](#addrewards)
    - [DeactivateVault](#deactivatevault)

## Concepts

- The power of a user comes from delegation power and staked power
  - delegation power is total coins that the address has delegated to validators.
  - staked power is the total coins that the address has staked to the module
- Users can stake their coins (such as liquid staking tokens) into the module to get staked power.
- Users cannot undelegate/unstake coins exceeding the locked power under any vault.
- Users must send a transaction to claim rewards by themselves.
  - They need to claim rewards by key of vault.
  - There is a CLI command that helps to claim all vaults (sending multiple msgs).
- Modules can lock the power of users by using key of vault.
- Modules can deposit coins into the vault.
  - The coins will be distributed to users based on the locked power of that key.
  - Use MasterChefV2 logic from SushiSwap
- Modules must call a provided function to deactivate a vault once it is no longer used.
  - Once deactivated, a vault cannot be reactivated.

## State

### Vault

The `Vault` is a space for holding the vault information.
- `0x00 | Key -> ProtocolBuffer(Vault)`

```protobuf
// Vault is used for tracking the status and rewards of the vaults.
message Vault {
  option (gogoproto.equal) = true;

  // key is the key of the vault.
  string key = 1;

  // vault_address is the address that holds rewards for this vault.
  string vault_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // is_active is the status of the vault
  bool is_active = 3;

  // rewards_per_power is rewards per power (DecCoins)
  // new_rewards_per_power = current_rewards_per_power + (rewards / total_power)
  repeated cosmos.base.v1beta1.DecCoin rewards_per_power = 4 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
    (gogoproto.nullable)     = false,
    (amino.dont_omitempty)   = true
  ];

  // total_power is the total locked power of the vault.
  string total_power = 5 [
    (cosmos_proto.scalar)  = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable)   = false
  ];

  // remainders is a list of the remainder amounts in the vault.
  // this field is used to track remainder amount from claimings in the vault.
  repeated cosmos.base.v1beta1.DecCoin remainders = 6 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
    (gogoproto.nullable)     = false,
    (amino.dont_omitempty)   = true
  ];
}
```

### Lock

The `Lock` is a space for holding the locking information of each account of each vault.
- `0x01 | AddrLength | Addr | Key -> ProtocolBuffer(Lock)`

```protobuf
// Lock is used to store lock information of each user on each vault along with their reward information.
message Lock {
  option (gogoproto.equal) = true;

  // staker_address is the owner's address of the staker.
  string staker_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // key is the key of the vault that this lock is locked to.
  string key = 2;

  // power is the number of locked power.
  string power = 3 [
    (cosmos_proto.scalar)  = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable)   = false
  ];

  // pos_reward_debts is a list of reward debt for each reward (only the positive side).
  // Note: Coin and DecCoin can't have negative amounts. so, we split it into two numbers.
  repeated cosmos.base.v1beta1.DecCoin pos_reward_debts = 4 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
    (gogoproto.nullable)     = false,
    (amino.dont_omitempty)   = true
  ];

  // neg_reward_debts is a list of reward debt for each reward (only negative side).
  repeated cosmos.base.v1beta1.DecCoin neg_reward_debts = 5 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
    (gogoproto.nullable)     = false,
    (amino.dont_omitempty)   = true
  ];
}
```

#### LocksByPowerIndex

`LocksByPowerIndex` allows to retrieve Stake ordering by power:
`0x80| AddrLength | Addr | BigEndian(Power) | Key -> Key` 

### Stake

The `Stake` is a space for holding the staking information of each address.
- `0x02 | AddrLength | Addr -> ProtocolBuffer(Stake)`

```protobuf
// Stake is used to store staked coins of an address.
message Stake {
  // staker_address is the address that this stake belongs to.
  string staker_address = 1;

  // coins are the coins that the address has staked.
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

### Params

The `restake` module stores its params in the state with the prefix of `0x90`, it can be updated with a governance proposal or the address with authority.
- `0x90 -> ProtocolBuffer(Params)`

```protobuf
// Params is the data structure that keeps the parameters.
message Params {
  // allowed_denoms is a list of denoms that the module allows to stake to get power.
  repeated string allowed_denoms = 1;
}
```

## Messages

In this section, we describe the processing of the `restake` messages and the corresponding updates to the state.

```protobuf
// Msg defines the restake Msg service.
service Msg {
  // ClaimRewards defines a method for claiming rewards for the user.
  rpc ClaimRewards(MsgClaimRewards) returns (MsgClaimRewardsResponse);

  // Stake defines a method for staking coins into the module.
  rpc Stake(MsgStake) returns (MsgStakeResponse);

  // Unstake defines a method for unstaking coins into the module.
  rpc Unstake(MsgUnstake) returns (MsgUnstakeResponse);

  // UpdateParams defines a method for updating parameters.
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
}
```

### MsgClaimRewards

A user can claim rewards by using the `MsgClaimRewards` message.

```protobuf

// MsgClaimRewards is the request message type for claiming rewards.
message MsgClaimRewards {
  option (cosmos.msg.v1.signer) = "staker_address";
  option (amino.name)           = "restake/MsgClaimRewards";

  // staker_address is the address that will claim the rewards.
  string staker_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // key is the key that want to claim rewards from.
  string key = 2;
}
```

**Logic**

- Calculate the rewards of the vault for the address and truncate the decimal.
  - Transfer the truncated amount from the module account to the user.
  - Re-calculate reward debts of the address.
- If the vault is inactive.
  - Delete `lock` state of the address
  - Add the left rewards of the address for this vault to the remainder amount of the vault.

### MsgStake

A user can stake allowed coins by using the `MsgStake` message.

```protobuf
// MsgStake is the request message type for staking coins.
message MsgStake {
  option (cosmos.msg.v1.signer) = "staker_address";
  option (amino.name)           = "restake/MsgStake";

  // staker_address is the address that will stake the coins.
  string staker_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // coins are the coins that will be staked.
  repeated cosmos.base.v1beta1.Coin coins = 2;
}

```

**Logic**

- Check if the denom is in `allowed_denom` parameter.
  - If not, return an error.
- Transfer coins to the global module account.

### MsgUnstake

A user can unstake staked coins by using the `MsgUnstake` message.

```protobuf
// MsgUnstake is the request message type for unstaking coins.
message MsgUnstake {
  option (cosmos.msg.v1.signer) = "staker_address";
  option (amino.name)           = "restake/MsgUnstake";

  // staker_address is the address that will unstake the coins.
  string staker_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // coins are the coins that will be unstaked.
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

**Logic**

- Check if the staked coins are more than the specific coins.
  - If not, return an error.
- Check if the locked power is still valid after unstaking coins.
  - If not, return an error.
- Transfer coins from the global module account to the address.

### MsgUpdateParams

The parameters of the module can be updated by using the `MsUpdateParams` message.

```protobuf
// MsgUpdateParams is the transaction message to update parameters.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "restake/MsgUpdateParams";

  // authority is the address of the governance account.
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params is parameters to update.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

**Logic**

- Check if authority is governance address
  - If not, return an error.
- Override parameters in the state.


## Staking hooks

The purpose is to prevent a user to un-delegate more than what is locked for any vault.

**Here is the logic that the module will check for each hook below.**
- Calculate the new total power
- Loop `LocksByPowerIndex` from the maximum locked power to the minimum locked power
    - Find the first active vault of the lock
        - if total power < locked power, return error

### BeforeDelegationRemoved

- The `Staking` module will call this function if a user un-delegates all tokens from the validator.

### AfterDelegationModified

- The `Staking` module will call this function if a user un-delegates partial tokens from the validator.


## Expected keepers

Here is the public function of `restake` keeper that other modules can use for locking power, adding rewards, and vault deactivation.

```go
type RestakeKeeper interface {
  SetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string, power math.Int) error
  GetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string) (math.Int, error)	
  AddRewards(ctx sdk.Context, sender sdk.AccAddress, key string, rewards sdk.Coins) error 
  DeactivateVault(ctx sdk.Context, key string) error
}
```

### SetLockedPower

`SetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string, power math.Int) error` 

This function is used to lock the power of an account to a specific vault.

**Logic**

- Return an error if the total power <= `power`
- Return an error if the vault is inactive.

### GetLockedPower

`GetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string) (math.Int, error)`

This function is used to get the locked power of the account on the vault.

**Logic**

- Return an error if the vault doesn’t exist.
- Return an error if the vault is inactive.
- Return an error if there is no lock for this account on this vault.

### AddRewards

`AddRewards(ctx sdk.Context, sender sdk.AccAddress, key string, rewards sdk.Coins) error`

This function is used to add reward coins to the vault.

**Logic**

- Return an error if the vault doesn’t exist.
- Return an error if the vault is inactive.
- Return an error if the total locked power to the vault is zero.

### DeactivateVault

`DeactivateVault(ctx sdk.Context, key string) error`

This function is used to set the status of the vault to inactive

**Note:** Once the vault is deactivated, it won’t be able to re-use again.

**Logic**

- Return an error if the vault doesn’t exist.
- Return an error if the vault is inactive.
