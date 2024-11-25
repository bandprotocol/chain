package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestTotalFees_Validate(t *testing.T) {
	tests := []struct {
		name      string
		totalFees types.TotalFees
		expErr    bool
	}{
		{
			name: "invalid total packet fee",
			totalFees: types.TotalFees{
				TotalPacketFee: sdk.Coins{(sdk.Coin{Denom: "uband", Amount: math.NewInt(-100)})},
			},
			expErr: true,
		},
		{
			name: "all good",
			totalFees: types.TotalFees{
				TotalPacketFee: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.totalFees.Validate()
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
