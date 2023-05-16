package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AuthzKeeper defines the expected authz keeper. for query and testing only don't use to create/remove grant on deliver tx
type AuthzKeeper interface {
	GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	SaveGrant(
		ctx sdk.Context,
		grantee, granter sdk.AccAddress,
		authorization authz.Authorization,
		expiration *time.Time,
	) error
}
