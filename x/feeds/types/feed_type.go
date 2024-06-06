package types

import tsslib "github.com/bandprotocol/chain/v2/pkg/tss"

var (
	FeedTypeDefaultPrefix = tsslib.Hash([]byte("Default"))[:4]
	FeedTypeTickPrefix    = tsslib.Hash([]byte("Tick"))[:4]
)
