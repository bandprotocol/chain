package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PostHandlerOptions are the options required for constructing PostHandlers.
type PostHandlerOptions struct{}

// NewPostHandler returns a PostHandler chain with decorators.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	postDecorators := []sdk.PostDecorator{}
	return sdk.ChainPostDecorators(postDecorators...), nil
}
