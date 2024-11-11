package updater

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"

	bothanclient "github.com/bandprotocol/bothan/bothan-api/client/go-client"

	feeds "github.com/bandprotocol/chain/v3/x/feeds/types"
)

type BothanClient interface {
	bothanclient.Client
}

type FeedQuerier interface {
	QueryCurrentFeeds() (*feeds.QueryCurrentFeedsResponse, error)
	QueryReferenceSourceConfig() (*feeds.QueryReferenceSourceConfigResponse, error)
}

type RemoteClient interface {
	rpcclient.RemoteClient
}
