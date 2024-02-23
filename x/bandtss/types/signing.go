package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
