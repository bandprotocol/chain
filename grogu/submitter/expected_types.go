package submitter

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	bothanclient "github.com/bandprotocol/bothan/bothan-api/client/go-client"
)

type RemoteClient interface {
	rpcclient.RemoteClient
}

type BothanClient interface {
	bothanclient.Client
}

type AuthQuerier interface {
	QueryAccount(address sdk.Address) (*auth.QueryAccountResponse, error)
}

type TxQuerier interface {
	QueryTx(hash string) (*sdk.TxResponse, error)
}
