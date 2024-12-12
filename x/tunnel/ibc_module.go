package tunnel

import (
	"math"
	"strings"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

var (
	_ porttypes.IBCModule             = (*IBCModule)(nil)
	_ porttypes.PacketDataUnmarshaler = (*IBCModule)(nil)
	_ porttypes.UpgradableModule      = (*IBCModule)(nil)
)

// IBCModule implements the ICS26 interface for tunnel given the tunnel keeper.
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(keeper keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: keeper,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	err := ValidateTunnelChannelParams(ctx, im.keeper, order, portID, channelID)
	if err != nil {
		return "", err
	}

	if !keeper.IsValidPortID(portID) {
		return "", types.ErrInvalidPortID
	}

	// If version is empty, set it to the current version
	if strings.TrimSpace(version) == "" {
		version = types.Version
	}

	if version != types.Version {
		return "", types.ErrInvalidVersion.Wrapf("got %s, expected %s", version, types.Version)
	}

	// openInit must claim the channelCapability that IBC passes into the callback
	if err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	err := ValidateTunnelChannelParams(ctx, im.keeper, order, portID, channelID)
	if err != nil {
		return "", err
	}

	if !keeper.IsValidPortID(portID) {
		return "", types.ErrInvalidPortID
	}

	if counterpartyVersion != types.Version {
		return "", types.ErrInvalidVersion.Wrapf(
			"invalid counterparty version: got: %s, expected %s",
			counterpartyVersion,
			types.Version,
		)
	}

	// module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// If module can already authenticate the capability then module already owns it so we don't need to claim
	// Otherwise, module does not have channel capability and we must claim it from IBC
	if !im.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		// Only claim channel capability passed back by IBC module if we do not already own it
		if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return "", err
		}
	}

	return types.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return types.ErrInvalidVersion.Wrapf(
			"invalid counterparty version: %s, expected %s",
			counterpartyVersion,
			types.Version,
		)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for tunnel channels
	return sdkerrors.ErrInvalidRequest.Wrap("user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// Disallow incoming packets for tunnel channels
	ack := channeltypes.NewErrorAcknowledgement(
		sdkerrors.ErrInvalidRequest.Wrap("tunnel does not accept incoming packets"),
	)
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// do nothing for out-going packet
	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// do nothing for out-going packet
	return nil
}

// OnChanUpgradeInit implements the IBCModule interface
func (im IBCModule) OnChanUpgradeInit(
	ctx sdk.Context,
	portID, channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) (string, error) {
	if err := ValidateTunnelChannelParams(ctx, im.keeper, proposedOrder, portID, channelID); err != nil {
		return "", err
	}

	if proposedVersion != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "expected %s, got %s", types.Version, proposedVersion)
	}

	return proposedVersion, nil
}

// OnChanUpgradeTry implements the IBCModule interface
func (im IBCModule) OnChanUpgradeTry(
	ctx sdk.Context,
	portID, channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	counterpartyVersion string,
) (string, error) {
	if err := ValidateTunnelChannelParams(ctx, im.keeper, proposedOrder, portID, channelID); err != nil {
		return "", err
	}

	if counterpartyVersion != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "expected %s, got %s", types.Version, counterpartyVersion)
	}

	return counterpartyVersion, nil
}

// OnChanUpgradeAck implements the IBCModule interface
func (IBCModule) OnChanUpgradeAck(ctx sdk.Context, portID, channelID, counterpartyVersion string) error {
	if counterpartyVersion != types.Version {
		return errorsmod.Wrapf(types.ErrInvalidVersion, "expected %s, got %s", types.Version, counterpartyVersion)
	}

	return nil
}

// OnChanUpgradeOpen implements the IBCModule interface
func (IBCModule) OnChanUpgradeOpen(
	ctx sdk.Context,
	portID, channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) {
}

// UnmarshalPacketData attempts to unmarshal the provided packet data bytes
// into a TunnelPricesPacketData. This function implements the optional
// PacketDataUnmarshaler interface required for ADR 008 support.
func (IBCModule) UnmarshalPacketData(bz []byte) (interface{}, error) {
	var packetData types.TunnelPricesPacketData
	if err := types.ModuleCdc.UnmarshalJSON(bz, &packetData); err != nil {
		return nil, err
	}

	return packetData, nil
}

// ValidateTunnelChannelParams does validation of a newly created tunnel channel. A tunnel
// channel must be ORDERED, use the correct port (by default 'tunnel'), and use the current
// supported version. Only 2^32 channels are allowed to be created.
func ValidateTunnelChannelParams(
	ctx sdk.Context,
	keeper keeper.Keeper,
	order channeltypes.Order,
	portID string,
	channelID string,
) error {
	// NOTE: for escrow address security only 2^32 channels are allowed to be created
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/7737
	channelSequence, err := channeltypes.ParseChannelSequence(channelID)
	if err != nil {
		return err
	}
	if channelSequence > uint64(math.MaxUint32) {
		return types.ErrMaxTunnelChannels.Wrapf(
			"channel sequence %d is greater than max allowed tunnel channels %d",
			channelSequence,
			uint64(math.MaxUint32),
		)
	}
	if order != channeltypes.UNORDERED {
		return channeltypes.ErrInvalidChannelOrdering.Wrapf(
			"expected %s channel, got %s",
			channeltypes.UNORDERED,
			order,
		)
	}

	return nil
}
