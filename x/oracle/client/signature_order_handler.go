package client

import (
	"github.com/bandprotocol/chain/v2/x/oracle/client/cli"
	tssclient "github.com/bandprotocol/chain/v2/x/tss/client"
)

// OracleSignatureOrderHandler is the request signature handler.
var OracleSignatureOrderHandler = tssclient.NewSignatureOrderHandler(cli.GetCmdRequestSignature)
