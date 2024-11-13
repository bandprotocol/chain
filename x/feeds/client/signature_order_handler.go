package client

import (
	bandtssclient "github.com/bandprotocol/chain/v3/x/bandtss/client"
	"github.com/bandprotocol/chain/v3/x/feeds/client/cli"
)

// FeedsRequestSignatureHandler is the request signature handler.
var FeedsRequestSignatureHandler = bandtssclient.NewRequestSignatureHandler(cli.GetCmdRequestSignature)
