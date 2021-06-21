package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// OnRecvPacket processes a cross chain oracle request. Data source fees
// are collected from an escrowAddress corresponding to the given requestKey.
func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data types.OracleRequestPacketData) (types.RequestID, error) {
	if err := data.ValidateBasic(); err != nil {
		return 0, err
	}

	escrowAddress := types.GetEscrowAddress(data.RequestKey, packet.DestinationPort, packet.DestinationChannel)
	ibcChannel := types.NewIBCChannel(packet.DestinationPort, packet.DestinationChannel)

	return k.PrepareRequest(ctx, &data, escrowAddress, &ibcChannel)
}
