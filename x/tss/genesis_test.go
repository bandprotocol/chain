package tss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/tss"
	"github.com/bandprotocol/chain/v2/x/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestExportGenesis(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	addr1 := bandtesting.Alice.Address
	addr2 := bandtesting.Bob.Address

	data := types.GenesisState{
		Params:     types.DefaultParams(),
		GroupCount: 1,
		Groups: []types.Group{
			{
				ID:        1,
				Size_:     1,
				Threshold: 1,
				PubKey:    nil,
				Status:    types.GROUP_STATUS_ROUND_1,
			},
		},
		Members: []types.Member{
			{
				ID:          1,
				GroupID:     1,
				Address:     addr1.String(),
				PubKey:      nil,
				IsMalicious: false,
				IsActive:    true,
			},
		},
		SigningCount: 0,
		Signings:     []types.Signing{},
		DEsGenesis: []types.DEGenesis{
			{
				Address: addr1.String(),
				DE: types.DE{
					PubD: []byte("pubD"),
					PubE: []byte("pubE"),
				},
			},
			{
				Address: addr1.String(),
				DE: types.DE{
					PubD: []byte("pubD2"),
					PubE: []byte("pubE2"),
				},
			},
			{
				Address: addr2.String(),
				DE: types.DE{
					PubD: []byte("pubD3"),
					PubE: []byte("pubE3"),
				},
			},
		},
	}

	tss.InitGenesis(ctx, k, &data)

	exported := tss.ExportGenesis(ctx, k)
	require.Equal(t, data.Params, exported.Params)
	require.Equal(t, data.DEsGenesis, exported.DEsGenesis)

	require.Equal(t, uint64(2), k.GetDECount(ctx, addr1))
	de, err := k.GetDE(ctx, addr1, 1)
	require.NoError(t, err)
	require.Equal(
		t,
		types.DE{
			PubD: []byte("pubD2"),
			PubE: []byte("pubE2"),
		},
		de,
	)
}
