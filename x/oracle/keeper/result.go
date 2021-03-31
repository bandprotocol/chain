package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"

	"github.com/GeoDB-Limited/odin-core/pkg/obi"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
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
func (k Keeper) ResolveSuccess(ctx sdk.Context, id types.RequestID, result []byte, gasUsed uint32) {
	k.SaveResult(ctx, id, types.RESOLVE_STATUS_SUCCESS, result)
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
) {
	r := k.MustGetRequest(ctx, id)
	reportCount := k.GetReportCount(ctx, id)
	k.SetResult(ctx, id, types.NewResult(
		r.ClientID,                         // ClientID
		r.OracleScriptID,                   // OracleScriptID
		r.Calldata,                         // Calldata
		uint64(len(r.RequestedValidators)), // AskCount
		r.MinCount,                         // MinCount
		id,                                 // RequestID
		reportCount,                        // AnsCount
		int64(r.RequestTime),               // RequestTime
		ctx.BlockTime().Unix(),             // ResolveTime
		status,                             // ResolveStatus
		result,                             // Result
	))

	if r.IBCSource != nil {
		sourceChannel := r.IBCSource.SourceChannel
		sourcePort := r.IBCSource.SourcePort
		sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
		if !found {
			panic(fmt.Sprintf("Cannot find channel on port ID (%s) channel ID (%s)", sourcePort, sourceChannel))
		}
		destinationPort := sourceChannelEnd.Counterparty.PortId
		destinationChannel := sourceChannelEnd.Counterparty.ChannelId
		sequence, found := k.channelKeeper.GetNextSequenceSend(
			ctx, sourcePort, sourceChannel,
		)
		if !found {
			panic(fmt.Sprintf("Cannot get sequence number on source port: %s, source channel: %s", sourcePort, sourceChannel))
		}
		channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
		if !ok {
			panic("Module does not own channel capability")
		}

		packetData := types.NewOracleResponsePacketData(
			r.ClientID, id, reportCount, int64(r.RequestTime), ctx.BlockTime().Unix(), status, result,
		)

		packet := channeltypes.NewPacket(
			packetData.GetBytes(),
			sequence,
			sourcePort,
			sourceChannel,
			destinationPort,
			destinationChannel,
			clienttypes.NewHeight(0, 0),
			uint64(ctx.BlockTime().UnixNano()+int64(10*time.Minute)), // TODO: Find what time out will be used on response packet
		)

		if err := k.channelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
			panic(err)
		}
	}
}
