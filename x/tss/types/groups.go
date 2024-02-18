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
	Authority string
}

type CreateGroupResult struct{}

type ReplaceGroupInput struct {
	CurrentGroupID tss.GroupID
	NewGroupID     tss.GroupID
	ExecTime       time.Time
	Authority      string
}

type ReplaceGroupResult struct{}
