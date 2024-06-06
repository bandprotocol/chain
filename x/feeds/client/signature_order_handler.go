package client

import (
	bandtssclient "github.com/bandprotocol/chain/v2/x/bandtss/client"
	"github.com/bandprotocol/chain/v2/x/feeds/client/cli"
)

// FeedsRequestSignatureHandler is the request signature handler.
var FeedsRequestSignatureHandler = bandtssclient.NewRequestSignatureHandler(cli.GetCmdRequestSignature)
