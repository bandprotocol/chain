package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	packetExpireTime = int64(10 * time.Minute)
)

// HasResult checks if the result of this request ID exists in the storage.
func (k Keeper) HasResult(ctx sdk.Context, id types.RequestID) bool {
	return ctx.KVStore(k.storeKey).Has(types.ResultStoreKey(id))
}

// SetResult sets result to the store.
func (k Keeper) SetResult(ctx sdk.Context, reqID types.RequestID, result types.Result) {
	ctx.KVStore(k.storeKey).Set(types.ResultStoreKey(reqID), k.cdc.MustMarshal(&result))
}

// MarshalResult marshal the result
func (k Keeper) MarshalResult(ctx sdk.Context, result types.Result) ([]byte, error) {
	return k.cdc.Marshal(&result)
}

// GetResult returns the result for the given request ID or error if not exists.
func (k Keeper) GetResult(ctx sdk.Context, id types.RequestID) (types.Result, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ResultStoreKey(id))
	if bz == nil {
		return types.Result{}, types.ErrResultNotFound.Wrapf("id: %d", id)
	}
	var result types.Result
	k.cdc.MustUnmarshal(bz, &result)
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
func (k Keeper) ResolveSuccess(
	ctx sdk.Context,
	id types.RequestID,
	requester string,
	feeLimit sdk.Coins,
	result []byte,
	gasUsed uint64,
	gid tss.GroupID,
	encodeType types.EncodeType,
) {
	k.SaveResult(ctx, id, types.RESOLVE_STATUS_SUCCESS, result)

	events := sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, fmt.Sprintf("%d", types.RESOLVE_STATUS_SUCCESS)),
		sdk.NewAttribute(types.AttributeKeyResult, hex.EncodeToString(result)),
		sdk.NewAttribute(types.AttributeKeyGasUsed, fmt.Sprintf("%d", gasUsed)),
	)

	if gid != tss.GroupID(0) {
		// Request sign by tss module
		var signingResult *types.SigningResult
		sid, err := k.tssKeeper.HandleRequestSign(
			ctx,
			gid,
			types.NewRequestingSignature(id, encodeType),
			sdk.MustAccAddressFromBech32(requester),
			feeLimit,
		)
		if err != nil {
			codespace, code, _ := sdkerrors.ABCIInfo(err, false)
			signingResult = &types.SigningResult{
				ErrorCodespace: codespace,
				ErrorCode:      uint64(code),
			}

			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeHandleRequestSignFail,
				sdk.NewAttribute(types.AttributeKeyID, fmt.Sprintf("%d", id)),
				sdk.NewAttribute(types.AttributeKeyTSSGroupID, fmt.Sprintf("%d", gid)),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			))

			events = events.AppendAttributes(
				sdk.NewAttribute(types.AttributeKeySigningErrCodespace, signingResult.ErrorCodespace),
				sdk.NewAttribute(types.AttributeKeySigningErrCode, fmt.Sprintf("%d", signingResult.ErrorCode)),
			)
		} else {
			signingResult = &types.SigningResult{
				SigningID: sid,
			}

			events = events.AppendAttributes(
				sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingResult.SigningID)),
			)
		}

		k.SetSigningResult(ctx, id, *signingResult)
	}

	ctx.EventManager().EmitEvent(events)
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

	if r.IBCChannel != nil {
		sourceChannel := r.IBCChannel.ChannelId
		sourcePort := r.IBCChannel.PortId

		channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
		if !ok {
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSendPacketFail,
				sdk.NewAttribute(types.AttributeKeyReason, "Module does not own channel capability"),
			))
			return
		}

		packetData := types.NewOracleResponsePacketData(
			r.ClientID, id, reportCount, int64(r.RequestTime), ctx.BlockTime().Unix(), status, result,
		)

		if _, err := k.channelKeeper.SendPacket(
			ctx,
			channelCap,
			sourcePort,
			sourceChannel,
			clienttypes.NewHeight(0, 0),
			uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
			packetData.GetBytes(),
		); err != nil {
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSendPacketFail,
				sdk.NewAttribute(types.AttributeKeyReason, fmt.Sprintf("Unable to send packet: %s", err)),
			))
		}
	}
}
