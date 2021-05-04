package oraclekeeper

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
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

// HasResult checks if the result of this request ID exists in the storage.
func (k Keeper) HasResult(ctx sdk.Context, id oracletypes.RequestID) bool {
	return ctx.KVStore(k.storeKey).Has(oracletypes.ResultStoreKey(id))
}

// SetResult sets result to the store.
func (k Keeper) SetResult(ctx sdk.Context, reqID oracletypes.RequestID, result oracletypes.Result) {
	store := ctx.KVStore(k.storeKey)
	store.Set(oracletypes.ResultStoreKey(reqID), obi.MustEncode(result))
}

// GetResult returns the result for the given request ID or error if not exists.
func (k Keeper) GetResult(ctx sdk.Context, id oracletypes.RequestID) (oracletypes.Result, error) {
	bz := ctx.KVStore(k.storeKey).Get(oracletypes.ResultStoreKey(id))
	if bz == nil {
		return oracletypes.Result{}, sdkerrors.Wrapf(oracletypes.ErrResultNotFound, "id: %d", id)
	}
	var result oracletypes.Result
	obi.MustDecode(bz, &result)
	return result, nil
}

// MustGetResult returns the result for the given request ID. Panics on error.
func (k Keeper) MustGetResult(ctx sdk.Context, id oracletypes.RequestID) oracletypes.Result {
	result, err := k.GetResult(ctx, id)
	if err != nil {
		panic(err)
	}
	return result
}

// ResolveSuccess resolves the given request as success with the given result.
func (k Keeper) ResolveSuccess(ctx sdk.Context, id oracletypes.RequestID, result []byte, gasUsed uint32) {
	k.SaveResult(ctx, id, oracletypes.RESOLVE_STATUS_SUCCESS, result)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, fmt.Sprintf("%d", oracletypes.RESOLVE_STATUS_SUCCESS)),
		sdk.NewAttribute(oracletypes.AttributeKeyResult, hex.EncodeToString(result)),
		sdk.NewAttribute(oracletypes.AttributeKeyGasUsed, fmt.Sprintf("%d", gasUsed)),
	))
}

// ResolveFailure resolves the given request as failure with the given reason.
func (k Keeper) ResolveFailure(ctx sdk.Context, id oracletypes.RequestID, reason string) {
	k.SaveResult(ctx, id, oracletypes.RESOLVE_STATUS_FAILURE, []byte{})
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, fmt.Sprintf("%d", oracletypes.RESOLVE_STATUS_FAILURE)),
		sdk.NewAttribute(oracletypes.AttributeKeyReason, reason),
	))
}

// ResolveExpired resolves the given request as expired.
func (k Keeper) ResolveExpired(ctx sdk.Context, id oracletypes.RequestID) {
	k.SaveResult(ctx, id, oracletypes.RESOLVE_STATUS_EXPIRED, []byte{})
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, fmt.Sprintf("%d", oracletypes.RESOLVE_STATUS_EXPIRED)),
	))
}

// SaveResult saves the result packets for the request with the given resolve status and result.
func (k Keeper) SaveResult(
	ctx sdk.Context, id oracletypes.RequestID, status oracletypes.ResolveStatus, result []byte,
) {
	r := k.MustGetRequest(ctx, id)
	reportCount := k.GetReportCount(ctx, id)
	k.SetResult(ctx, id, oracletypes.NewResult(
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

		packetData := oracletypes.NewOracleResponsePacketData(
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
