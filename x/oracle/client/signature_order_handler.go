package client

import (
	bandtssclient "github.com/bandprotocol/chain/v2/x/bandtss/client"
	"github.com/bandprotocol/chain/v2/x/oracle/client/cli"
)

// OracleRequestSignatureHandler is the request signature handler.
var OracleRequestSignatureHandler = bandtssclient.NewRequestSignatureHandler(cli.GetCmdRequestSignature)
