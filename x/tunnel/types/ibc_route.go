package types

import (
	"fmt"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBCRoute defines the IBC route for the tunnel module
var _ RouteI = &IBCRoute{}

// IsChannelIDFormat checks if a channelID is in the format required on the SDK for
// parsing channel identifiers. The channel identifier must be in the form: `channel-{N}
var IsChannelIDFormat = regexp.MustCompile(`^channel-[0-9]{1,20}$`).MatchString

// Route defines the IBC route for the tunnel module
func (r *IBCRoute) ValidateBasic() error {
	// Validate the ChannelID format
	if !IsChannelIDFormat(r.ChannelID) {
		return fmt.Errorf("channel identifier is not in the format: `channel-{N}`")
	}
	return nil
}

func (r *IBCRoute) Fee() (sdk.Coins, error) {
	return sdk.NewCoins(), nil
}
