package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"

	"github.com/bandprotocol/chain/pkg/obi"
	"github.com/bandprotocol/chain/x/oracle/types"
)

// HasResult checks if the result of this request ID exists in the storage.
func (k Keeper) HasResult(ctx sdk.Context, id types.RequestID) bool {
	return ctx.KVStore(k.storeKey).Has(types.ResultStoreKey(id))
}

// SetResult sets result to the store.
func (k Keeper) SetResult(ctx sdk.Context, reqID types.RequestID, result types.Result) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ResultStoreKey(reqID), obi.MustEncode(result))
}

// GetResult returns the result for the given request ID or error if not exists.
func (k Keeper) GetResult(ctx sdk.Context, id types.RequestID) (types.Result, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ResultStoreKey(id))
	if bz == nil {
		return types.Result{}, sdkerrors.Wrapf(types.ErrResultNotFound, "id: %d", id)
	}
	var result types.Result
	obi.MustDecode(bz, &result)
	return result, nil
}

// MustGetResult returns the result for the given request ID. Panics on error.
func (k Keeper) MustGetResult(ctx sdk.Context, id types.RequestID) types.Result {
	result, err := k.GetResult(ctx, id)
	if err != nil {
		panic(err)
	}
	return result
}

// ResolveSuccess resolves the given request as success with the given result.
func (k Keeper) ResolveSuccess(ctx sdk.Context, id types.RequestID, result []byte, gasUsed uint32, ibcSource *types.IBCSource) {
	_, rep := k.SaveResult(ctx, id, types.RESOLVE_STATUS_SUCCESS, result)

	if ibcSource != nil {
		sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, ibcSource.SourcePort, ibcSource.SourceChannel)
		if !found {
			// TODO: Better error handler
			panic("unknown channel")
		}
		destinationPort := sourceChannelEnd.Counterparty.PortId
		destinationChannel := sourceChannelEnd.Counterparty.ChannelId
		sequence, found := k.channelKeeper.GetNextSequenceSend(
			ctx, ibcSource.SourcePort, ibcSource.SourceChannel,
		)
		channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(ibcSource.SourcePort, ibcSource.SourceChannel))
		if !ok {
			// TODO: Better error handler
			panic("module does not own channel capability")
		}

		err := k.channelKeeper.SendPacket(ctx, channelCap, channeltypes.NewPacket(
			rep.GetBytes(),
			sequence,
			ibcSource.SourcePort,
			ibcSource.SourceChannel,
			destinationPort,
			destinationChannel,
			clienttypes.NewHeight(0, 10000), // Arbitary height
			0,                               // Arbitrarily timeout for now
		))
		if err != nil {
			panic(err)
		}
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, fmt.Sprintf("%d", types.RESOLVE_STATUS_SUCCESS)),
		sdk.NewAttribute(types.AttributeKeyResult, hex.EncodeToString(result)),
		sdk.NewAttribute(types.AttributeKeyGasUsed, fmt.Sprintf("%d", gasUsed)),
	))
}

// ResolveFailure resolves the given request as failure with the given reason.
func (k Keeper) ResolveFailure(ctx sdk.Context, id types.RequestID, reason string) {
	k.SaveResult(ctx, id, types.RESOLVE_STATUS_FAILURE, []byte{})
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, fmt.Sprintf("%d", types.RESOLVE_STATUS_FAILURE)),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	))
}

// ResolveExpired resolves the given request as expired.
func (k Keeper) ResolveExpired(ctx sdk.Context, id types.RequestID) {
	k.SaveResult(ctx, id, types.RESOLVE_STATUS_EXPIRED, []byte{})
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, fmt.Sprintf("%d", types.RESOLVE_STATUS_EXPIRED)),
	))
}

// SaveResult saves the result packets for the request with the given resolve status and result.
func (k Keeper) SaveResult(
	ctx sdk.Context, id types.RequestID, status types.ResolveStatus, result []byte,
) (types.OracleRequestPacketData, types.OracleResponsePacketData) {
	r := k.MustGetRequest(ctx, id)
	reqPacket := types.NewOracleRequestPacketData(
		r.ClientID,                         // ClientID
		r.OracleScriptID,                   // OracleScriptID
		r.Calldata,                         // Calldata
		uint64(len(r.RequestedValidators)), // AskCount
		r.MinCount,                         // Mincount
	)
	resPacket := types.NewOracleResponsePacketData(
		r.ClientID,                // ClientID
		id,                        // RequestID
		k.GetReportCount(ctx, id), // AnsCount
		int64(r.RequestTime),      // RequestTime
		ctx.BlockTime().Unix(),    // ResolveTime
		status,                    // ResolveStatus
		result,                    // Result
	)
	k.SetResult(ctx, id, types.NewResult(reqPacket, resPacket))
	return reqPacket, resPacket
}
