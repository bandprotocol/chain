package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrAuctionHasClosed = sdkerrors.Register(ModuleName, 126, "The auction has already closed")
	ErrAuctionHasStarted = sdkerrors.Register(ModuleName, 127, "The auction has already started")
)
