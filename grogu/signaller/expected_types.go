package signaller

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client"

	feeds "github.com/bandprotocol/chain/v3/x/feeds/types"
)

type BothanClient interface {
	bothan.Client
}

type FeedQuerier interface {
	QueryValidValidator(valAddress sdk.ValAddress) (*feeds.QueryValidValidatorResponse, error)
	QueryValidatorPrices(valAddress sdk.ValAddress) (*feeds.QueryValidatorPricesResponse, error)
	QueryParams() (*feeds.QueryParamsResponse, error)
	QueryCurrentFeeds() (*feeds.QueryCurrentFeedsResponse, error)
}
