package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestAbsInt64(t *testing.T) {
	require.Equal(t, int64(5), types.AbsInt64(-5))
	require.Equal(t, int64(5), types.AbsInt64(5))
	require.Equal(t, int64(0), types.AbsInt64(0))
}
