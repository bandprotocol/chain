package signaller

import (
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bothanclient "github.com/bandprotocol/bothan/bothan-api/client/go-client"

	feeds "github.com/bandprotocol/chain/v3/x/feeds/types"
)

type BothanClient interface {
	bothanclient.Client
}

type FeedQuerier interface {
	QueryValidValidator(valAddress sdk.ValAddress) (*feeds.QueryValidValidatorResponse, error)
	QueryValidatorPrices(valAddress sdk.ValAddress) (*feeds.QueryValidatorPricesResponse, error)
	QueryParams() (*feeds.QueryParamsResponse, error)
	QueryCurrentFeeds() (*feeds.QueryCurrentFeedsResponse, error)
}

type NodeQuerier interface {
	QueryStatus() (*node.StatusResponse, error)
}
