package submitter

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type RemoteClient interface {
	rpcclient.RemoteClient
}

type AuthQuerier interface {
	QueryAccount(address sdk.Address) (*auth.QueryAccountResponse, error)
}

type TxQuerier interface {
	QueryTx(hash string) (*sdk.TxResponse, error)
}
