package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// OnRecvPacket processes a cross chain oracle request. Data source fees
// are collected from the relayer account.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.OracleRequestPacketData,
	relayer sdk.AccAddress,
) (types.RequestID, error) {
	if err := data.ValidateBasic(); err != nil {
		return 0, err
	}
	ibcChannel := types.NewIBCChannel(packet.DestinationPort, packet.DestinationChannel)

	return k.PrepareRequest(ctx, &data, relayer, &ibcChannel)
}
