package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestGetSetRequestToSigningMap(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	rid := types.RequestID(123)
	sid := tss.SigningID(456)

	// Set the request to signing map
	k.SetRequestToSigningMap(ctx, rid, sid)

	// Get the signing ID associated with the request ID
	gotSid, err := k.GetRequestToSigningMap(ctx, rid)
	require.NoError(t, err)
	require.Equal(t, sid, gotSid)
}
