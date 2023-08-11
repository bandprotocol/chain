package client

import (
	"github.com/bandprotocol/chain/v2/x/oracle/client/cli"
	tssclient "github.com/bandprotocol/chain/v2/x/tss/client"
)

// RequestSignatureHandler is the request signature handler.
var RequestSignatureHandler = tssclient.NewRequestSignatureHandler(cli.GetCmdRequestSignature)
