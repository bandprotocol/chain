package emitter

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	SenderAddress   = sdk.AccAddress(genAddresFromString("Sender"))
	ValAddress      = sdk.ValAddress(genAddresFromString("Validator"))
	TreasuryAddress = sdk.AccAddress(genAddresFromString("Treasury"))
	OwnerAddress    = sdk.AccAddress(genAddresFromString("Owner"))
	ReporterAddress = sdk.AccAddress(genAddresFromString("Reporter"))
)

func genAddresFromString(s string) []byte {
	var b [20]byte
	copy(b[:], s)
	return b[:]
}

func testCompareJson(t *testing.T, msg sdk.Msg, expect string) {
	res, err := json.Marshal(msg)
	require.NoError(t, err)
	require.Equal(t, expect, string(res))
}
