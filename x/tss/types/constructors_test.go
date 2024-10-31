package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestValidateGroup(t *testing.T) {
	testcases := []struct {
		name        string
		group       types.Group
		expectedErr error
	}{
		{
			"valid group",
			types.Group{
				ID:            1,
				Size_:         3,
				Threshold:     2,
				PubKey:        validTssPoint,
				Status:        types.GROUP_STATUS_ACTIVE,
				CreatedHeight: 1,
				ModuleOwner:   "module",
			},
			nil,
		},
		{
			"valid - public key is nil",
			types.Group{
				ID:            1,
				Size_:         3,
				Threshold:     2,
				PubKey:        nil,
				Status:        types.GROUP_STATUS_ROUND_1,
				CreatedHeight: 1,
				ModuleOwner:   "module",
			},
			nil,
		},
		{
			"invalid public key",
			types.Group{
				ID:            1,
				Size_:         3,
				Threshold:     2,
				PubKey:        tss.Point(testutil.HexDecode("0002")),
				Status:        types.GROUP_STATUS_ROUND_1,
				CreatedHeight: 1,
				ModuleOwner:   "module",
			},
			types.ErrInvalidGroup,
		},
		{
			"invalid group - incorrect threshold",
			types.Group{
				ID:            1,
				Size_:         1,
				Threshold:     2,
				PubKey:        validTssPoint,
				Status:        types.GROUP_STATUS_ACTIVE,
				CreatedHeight: 1,
				ModuleOwner:   "module",
			},
			types.ErrInvalidGroup,
		},
		{
			"invalid group - incorrect groupID",
			types.Group{
				ID:            0,
				Size_:         1,
				Threshold:     2,
				PubKey:        validTssPoint,
				Status:        types.GROUP_STATUS_ACTIVE,
				CreatedHeight: 1,
				ModuleOwner:   "module",
			},
			types.ErrInvalidGroup,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.group.Validate()
			require.ErrorIs(t, tc.expectedErr, err)
		})
	}
}
