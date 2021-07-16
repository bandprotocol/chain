package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrAuctionIsAlreadyClosed = sdkerrors.Register(ModuleName, 126, "The auction is already closed")
	ErrAuctionIsAlreadyOpened = sdkerrors.Register(ModuleName, 127, "The auction is already opened")
	ErrAuctionIsNotPending    = sdkerrors.Register(ModuleName, 128, "The auction is not pending")
)
