package types

import (
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CreateGroupInput struct {
	Members   []string
	Threshold uint64
	Fee       sdk.Coins
}

type CreateGroupResult struct {
	Group      Group
	DKGContext []byte
}

type ReplaceGroupInput struct {
	CurrentGroup Group
	NewGroup     Group
	ExecTime     time.Time
	FeePayer     sdk.AccAddress
}

type ReplaceGroupResult struct {
	Replacement Replacement
}

type UpdateGroupFeeInput struct {
	GroupID tss.GroupID
	Fee     sdk.Coins
}

type UpdateGroupFeeResult struct {
	Group Group
}
