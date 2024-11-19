package types_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestAbsInt64(t *testing.T) {
	require.Equal(t, int64(5), types.AbsInt64(-5))
	require.Equal(t, int64(5), types.AbsInt64(5))
	require.Equal(t, int64(0), types.AbsInt64(0))
}

func TestStringToBytes32(t *testing.T) {
	testcases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "atom-usd",
			in:   "CS:ATOM-USD",
			out:  "00000000000000000000000000000000000000000043533a41544f4d2d555344",
		},
		{
			name: "band-usd",
			in:   "CS:BAND-USD",
			out:  "00000000000000000000000000000000000000000043533a42414e442d555344",
		},
		{
			name: "too long string",
			in:   "this-is-too-long-string-that-cannot-be-converted",
			out:  "",
			err:  fmt.Errorf("string is too long"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := types.StringToBytes32(tc.in)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				return
			} else {
				require.NoError(t, err)

				bz := make([]byte, 32)
				copy(bz, out[0:32])

				outStr := hex.EncodeToString(bz)
				require.Equal(t, tc.out, outStr)
			}
		})
	}
}
