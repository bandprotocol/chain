package tss_test

import (
	"sort"
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

	sort.Slice(exported.DEsGenesis, func(i, j int) bool {
		if exported.DEsGenesis[i].Address != exported.DEsGenesis[j].Address {
			return exported.DEsGenesis[i].Address < exported.DEsGenesis[j].Address
		}
		return i < j
	})

	sort.Slice(data.DEsGenesis, func(i, j int) bool {
		if data.DEsGenesis[i].Address != data.DEsGenesis[j].Address {
			return data.DEsGenesis[i].Address < data.DEsGenesis[j].Address
		}
		return i < j
	})
	require.Equal(t, data.DEsGenesis, exported.DEsGenesis)

	require.Equal(t, uint64(2), k.GetDECount(ctx, addr1))
	hasDE := k.HasDE(ctx, addr1, types.DE{PubD: []byte("pubD2"), PubE: []byte("pubE2")})
	require.True(t, hasDE)
}
