package types

import (
	"fmt"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

// IBCRoute defines the IBC route for the tunnel module
var _ RouteI = &IBCRoute{}

// IsChannelIDFormat checks if a channelID is in the format required on the SDK for
// parsing channel identifiers. The channel identifier must be in the form: `channel-{N}
var IsChannelIDFormat = regexp.MustCompile(`^channel-[0-9]{1,20}$`).MatchString

// NewIBCRoute creates a new IBCRoute instance.
func NewIBCRoute(channelID string) *IBCRoute {
	return &IBCRoute{
		ChannelID: channelID,
	}
}

// Route defines the IBC route for the tunnel module
func (r *IBCRoute) ValidateBasic() error {
	// Validate the ChannelID format
	if !IsChannelIDFormat(r.ChannelID) {
		return fmt.Errorf("channel identifier is not in the format: `channel-{N}`")
	}
	return nil
}

// NewIBCPacketContent creates a new IBCPacketContent instance.
func NewIBCPacketContent(channelID string, sequence uint64) *IBCPacketContent {
	return &IBCPacketContent{
		ChannelID: channelID,
		Sequence:  sequence,
	}
}

// NewIBCPacketResult creates a new IBCPacketResult instance.
func NewIBCPacketResult(
	tunnelID uint64,
	sequence uint64,
	prices []feedstypes.Price,
	created_at int64,
) IBCPacketResult {
	return IBCPacketResult{
		TunnelID:  tunnelID,
		Sequence:  sequence,
		Prices:    prices,
		CreatedAt: created_at,
	}
}

// GetBytes returns the IBCPacketResult bytes
func (p IBCPacketResult) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
