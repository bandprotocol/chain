package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestIBCHookRoute_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		route  types.IBCHookRoute
		expErr bool
		errMsg string
	}{
		{
			name: "invalid channel ID",
			route: types.IBCHookRoute{
				ChannelID: "invalid-channel",
			},
			expErr: true,
			errMsg: "channel identifier is not in the format: `channel-{N}`",
		},
		{
			name: "empty channel ID",
			route: types.IBCHookRoute{
				ChannelID: "",
			},
			expErr: true,
			errMsg: "channel identifier is not in the format: `channel-{N}`",
		},
		{
			name: "all good",
			route: types.IBCHookRoute{
				ChannelID: "channel-1",
			},
			expErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.route.ValidateBasic()
			if tt.expErr {
				require.Error(t, err)
				require.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
