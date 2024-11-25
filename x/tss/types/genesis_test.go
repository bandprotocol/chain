package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

func TestGenesisStateValidate(t *testing.T) {
	validTssPoint2 := tss.Point(
		testutil.HexDecode("02117a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad41"),
	)

	validMemberAddrs := []string{
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
	}

	validGroups := []types.Group{
		{
			ID:            1,
			Size_:         2,
			Threshold:     1,
			PubKey:        validTssPoint,
			Status:        types.GROUP_STATUS_ACTIVE,
			CreatedHeight: 1,
			ModuleOwner:   "test",
		},
	}

	validMembers := []types.Member{
		{ID: 1, GroupID: 1, Address: validMemberAddrs[0], PubKey: validTssPoint},
		{ID: 2, GroupID: 1, Address: validMemberAddrs[1], PubKey: validTssPoint},
	}

	validDEs := []types.DEGenesis{
		{
			Address: validMemberAddrs[0],
			DE: types.DE{
				PubD: validTssPoint,
				PubE: validTssPoint,
			},
		},
		{
			Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			DE: types.DE{
				PubD: validTssPoint2,
				PubE: validTssPoint2,
			},
		},
	}

	testCases := []struct {
		name         string
		genesisState types.GenesisState
		expErr       bool
	}{
		{
			"valid genesisState",
			types.GenesisState{
				Params:  types.DefaultParams(),
				Groups:  validGroups,
				Members: validMembers,
				DEs:     validDEs,
			},
			false,
		},
		{
			"invalid genesisState - member size doesn't match",
			types.GenesisState{
				Params:  types.DefaultParams(),
				Groups:  validGroups,
				Members: []types.Member{validMembers[0]},
				DEs:     validDEs,
			},
			true,
		},
		{
			"invalid genesisState - duplicate member ID",
			types.GenesisState{
				Params: types.DefaultParams(),
				Groups: []types.Group{validGroups[0]},
				Members: []types.Member{
					{ID: 1, GroupID: 1, Address: validMemberAddrs[0], PubKey: validTssPoint},
					{ID: 1, GroupID: 1, Address: validMemberAddrs[1], PubKey: validTssPoint},
				},
				DEs: validDEs,
			},
			true,
		},
		{
			"invalid genesisState - duplicate member Address",
			types.GenesisState{
				Params: types.DefaultParams(),
				Groups: []types.Group{validGroups[0]},
				Members: []types.Member{
					{ID: 1, GroupID: 1, Address: validMemberAddrs[0], PubKey: validTssPoint},
					{ID: 2, GroupID: 1, Address: validMemberAddrs[0], PubKey: validTssPoint},
				},
				DEs: validDEs,
			},
			true,
		},
		{
			"invalid genesisState - duplicate group",
			types.GenesisState{
				Params:  types.DefaultParams(),
				Groups:  []types.Group{validGroups[0], validGroups[0]},
				Members: []types.Member{validMembers[0]},
				DEs:     validDEs,
			},
			true,
		},
		{
			"invalid genesisState - member not in group",
			types.GenesisState{
				Params:  types.DefaultParams(),
				Groups:  []types.Group{},
				Members: []types.Member{validMembers[0]},
				DEs:     validDEs,
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.genesisState.Validate()

			if tc.expErr {
				require.Error(tt, err)
			} else {
				require.NoError(tt, err)
			}
		})
	}
}
