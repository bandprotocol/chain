package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type HandleCreateSigningInput struct {
	GroupID  tss.GroupID
	Content  tsstypes.Content
	Sender   sdk.AccAddress
	FeeLimit sdk.Coins
}

type HandleCreateSigningResult struct {
	Message []byte
	Signing tsstypes.Signing
}
