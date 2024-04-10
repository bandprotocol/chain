package keeper

import (
	"fmt"

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

// SetMember sets a status of a member of the current group in the store.
func (k Keeper) SetMember(ctx sdk.Context, member types.Member) {
	address := sdk.MustAccAddressFromBech32(member.Address)
	ctx.KVStore(k.storeKey).Set(types.MemberStoreKey(address), k.cdc.MustMarshal(&member))
}

// GetMembersIterator gets an iterator all statuses of address.
func (k Keeper) GetMembersIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MemberStoreKeyPrefix)
}

// HasMember checks that address is in the store or not.
func (k Keeper) HasMember(ctx sdk.Context, address sdk.AccAddress) bool {
	return ctx.KVStore(k.storeKey).Has(types.MemberStoreKey(address))
}

// GetMember retrieves a member by address.
func (k Keeper) GetMember(ctx sdk.Context, address sdk.AccAddress) (types.Member, error) {
	if !k.HasMember(ctx, address) {
		return types.Member{}, types.ErrMemberNotFound.Wrapf("address: %s", address)
	}
	bz := ctx.KVStore(k.storeKey).Get(types.MemberStoreKey(address))

	member := types.Member{}
	err := k.cdc.Unmarshal(bz, &member)
	if err != nil {
		return types.Member{}, err
	}
	return member, nil
}

// GetMembers retrieves all statuses of the store.
func (k Keeper) GetMembers(ctx sdk.Context) []types.Member {
	var statuses []types.Member
	iterator := k.GetMembersIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var status types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &status)
		statuses = append(statuses, status)
	}
	return statuses
}

// DeleteMember removes the status of the address of the group
func (k Keeper) DeleteMember(ctx sdk.Context, address sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(types.MemberStoreKey(address))
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

// SetReplacementSigningID sets a signingID of group replacement singing request.
func (k Keeper) SetReplacement(ctx sdk.Context, replacement types.Replacement) {
	ctx.KVStore(k.storeKey).Set(types.ReplacementStoreKey, k.cdc.MustMarshal(&replacement))
}

// GetReplacementSigningID retrieves a signingID of group replacement singing request.
func (k Keeper) GetReplacement(ctx sdk.Context) (types.Replacement, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ReplacementStoreKey)

	replacement := types.Replacement{}
	err := k.cdc.Unmarshal(bz, &replacement)
	if err != nil {
		return types.Replacement{}, err
	}
	return replacement, nil
}
