package client

import (
	"github.com/bandprotocol/chain/v2/x/oracle/client/cli"
	tssclient "github.com/bandprotocol/chain/v2/x/tss/client"
)

// RequestingSignatureHandler is the request signature handler.
var RequestingSignatureHandler = tssclient.NewRequestingSignatureHandler(cli.GetCmdRequestSignature)
