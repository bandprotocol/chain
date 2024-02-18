package types

import (
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// TSSKeeper defines the expected tss keeper (noalias)
type TSSKeeper interface {
	CreateGroup(ctx sdk.Context, input tsstypes.CreateGroupInput) (*tsstypes.CreateGroupResult, error)
	ReplaceGroup(ctx sdk.Context, input tsstypes.ReplaceGroupInput) (*tsstypes.ReplaceGroupResult, error)
}
