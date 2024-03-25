package keeper

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistrKeeper
	stakingKeeper types.StakingKeeper
	tssKeeper     types.TSSKeeper

	authority        string
	feeCollectorName string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistrKeeper,
	stakingKeeper types.StakingKeeper,
	tssKeeper types.TSSKeeper,
	authority string,
	feeCollectorName string,
) *Keeper {
	// ensure Bandtss module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		paramSpace:       paramSpace,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		distrKeeper:      distrKeeper,
		stakingKeeper:    stakingKeeper,
		tssKeeper:        tssKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
	}
}

// GetBandtssAccount returns the Bandtss ModuleAccount
func (k Keeper) GetBandtssAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetMemberStatus sets a status of a member of the group in the store.
func (k Keeper) SetStatus(ctx sdk.Context, status types.Status) {
	address := sdk.MustAccAddressFromBech32(status.Address)
	ctx.KVStore(k.storeKey).Set(types.StatusStoreKey(address), k.cdc.MustMarshal(&status))
}

// GetStatusesIterator gets an iterator all statuses of address.
func (k Keeper) GetStatusesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StatusStoreKeyPrefix)
}

// GetStatus retrieves a status of the address.
func (k Keeper) GetStatus(ctx sdk.Context, address sdk.AccAddress) types.Status {
	bz := ctx.KVStore(k.storeKey).Get(types.StatusStoreKey(address))
	if bz == nil {
		return types.Status{
			Address: address.String(),
			Status:  types.MEMBER_STATUS_UNSPECIFIED,
			Since:   time.Time{},
		}
	}

	status := types.Status{}
	k.cdc.MustUnmarshal(bz, &status)
	return status
}

// GetStatuses retrieves all statuses of the store.
func (k Keeper) GetStatuses(ctx sdk.Context) []types.Status {
	var statuses []types.Status
	iterator := k.GetStatusesIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var status types.Status
		k.cdc.MustUnmarshal(iterator.Value(), &status)
		statuses = append(statuses, status)
	}
	return statuses
}

// DeleteStatus removes the status of the address of the group
func (k Keeper) DeleteStatus(ctx sdk.Context, address sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(types.StatusStoreKey(address))
}

// SetCurrentGroupID sets a current groupID of the Bandtss module.
func (k Keeper) SetCurrentGroupID(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Set(types.CurrentGroupIDStoreKey, sdk.Uint64ToBigEndian(uint64(groupID)))
}

// GetCurrentGroupID retrieves a current groupID of the Bandtss module.
func (k Keeper) GetCurrentGroupID(ctx sdk.Context) tss.GroupID {
	return tss.GroupID(sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.CurrentGroupIDStoreKey)))
}

// SetReplacingGroupID sets a replacing groupID of the Bandtss module.
func (k Keeper) SetReplacingGroupID(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Set(types.ReplacingGroupIDStoreKey, sdk.Uint64ToBigEndian(uint64(groupID)))
}

// GetReplacingGroupID retrieves a replacing groupID of the Bandtss module.
func (k Keeper) GetReplacingGroupID(ctx sdk.Context) tss.GroupID {
	return tss.GroupID(sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.ReplacingGroupIDStoreKey)))
}

// SetSigningFee sets a signing fee of the Bandtss module.
func (k Keeper) SetSigningFee(ctx sdk.Context, signingFee types.SigningFee) {
	ctx.KVStore(k.storeKey).Set(types.SigningFeeStoreKey(signingFee.SigningID), k.cdc.MustMarshal(&signingFee))
}

// GetSigningFee retrieves a signing fee of the Bandtss module.
func (k Keeper) GetSigningFee(ctx sdk.Context, signingID tss.SigningID) (types.SigningFee, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningFeeStoreKey(signingID))
	if bz == nil {
		return types.SigningFee{}, types.ErrSigningFeeNotFound.Wrapf("signingID: %d", signingID)
	}

	signingFee := types.SigningFee{}
	k.cdc.MustUnmarshal(bz, &signingFee)
	return signingFee, nil
}

// GetSigningFeeIterator gets an iterator all signingFee.
func (k Keeper) GetSigningFeeIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningFeeStoreKeyPrefix)
}

// GetSigningFees retrieves all signingFee of the store.
func (k Keeper) GetSigningFees(ctx sdk.Context) []types.SigningFee {
	var signingFees []types.SigningFee
	iterator := k.GetSigningFeeIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var signingFee types.SigningFee
		k.cdc.MustUnmarshal(iterator.Value(), &signingFee)
		signingFees = append(signingFees, signingFee)
	}
	return signingFees
}

// DeleteSigningFee removes the signing fee of the signingID
func (k Keeper) DeleteSigningFee(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningFeeStoreKey(signingID))
}
