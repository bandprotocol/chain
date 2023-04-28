package cli_test

// import (
// 	"testing"
// 	"time"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"

// 	"github.com/bandprotocol/chain/v2/x/tss/client/cli"
// 	"github.com/bandprotocol/chain/v2/x/tss/types"
// )

// func genAddresFromString(s string) []byte {
// 	var b [20]byte
// 	copy(b[:], s)
// 	return b[:]
// }

// func TestRequestStoreKey(t *testing.T) {
// 	granterAddress := sdk.AccAddress(genAddresFromString("granter"))
// 	granteeAddress := sdk.AccAddress(genAddresFromString("grantee"))

// 	expTime := time.Unix(0, 0)

// 	_, err := cli.CombineGrantMsgs(granterAddress, granteeAddress, types.MsgGrants, &expTime)
// 	require.NoError(t, err)
// }
