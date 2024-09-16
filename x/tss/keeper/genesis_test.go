package keeper_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestExportGenesis(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	addr1 := bandtesting.Alice.Address
	addr2 := bandtesting.Bob.Address

	data := types.GenesisState{
		Params: types.DefaultParams(),
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
		DEs: []types.DEGenesis{
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

	k.InitGenesis(ctx, &data)

	exported := k.ExportGenesis(ctx)
	require.Equal(t, data.Params, exported.Params)
	require.Equal(t, uint64(1), k.GetGroupCount(ctx))

	sort.Slice(exported.DEs, func(i, j int) bool {
		if exported.DEs[i].Address != exported.DEs[j].Address {
			return exported.DEs[i].Address < exported.DEs[j].Address
		}
		return i < j
	})

	sort.Slice(data.DEs, func(i, j int) bool {
		if data.DEs[i].Address != data.DEs[j].Address {
			return data.DEs[i].Address < data.DEs[j].Address
		}
		return i < j
	})
	require.Equal(t, data.DEs, exported.DEs)

	require.Equal(t, types.DEQueue{Head: 0, Tail: 2}, k.GetDEQueue(ctx, addr1))
	existingDEs := []types.DE{}
	for i := 0; i < 2; i++ {
		de, err := k.GetDE(ctx, addr1, uint64(i))
		require.NoError(t, err)
		existingDEs = append(existingDEs, de)
	}
	require.Equal(t, existingDEs, []types.DE{
		{
			PubD: []byte("pubD"),
			PubE: []byte("pubE"),
		},
		{
			PubD: []byte("pubD2"),
			PubE: []byte("pubE2"),
		},
	})

	k.SetDEQueue(ctx, addr1, types.DEQueue{Head: 1, Tail: 2})

	exported = k.ExportGenesis(ctx)
	require.Len(t, exported.DEs, 2)
	require.Contains(t, exported.DEs, types.DEGenesis{
		Address: addr1.String(),
		DE: types.DE{
			PubD: []byte("pubD2"),
			PubE: []byte("pubE2"),
		},
	})
	require.Contains(t, exported.DEs, types.DEGenesis{
		Address: addr2.String(),
		DE: types.DE{
			PubD: []byte("pubD3"),
			PubE: []byte("pubE3"),
		},
	})
}
