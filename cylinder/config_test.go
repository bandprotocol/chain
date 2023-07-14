package cylinder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func TestIsAllowedGroup(t *testing.T) {
	tests := []struct {
		name    string
		cfg     cylinder.Config
		groupID tss.GroupID
		expect  bool
	}{
		{
			"allow - first",
			cylinder.Config{
				GroupIDs: []uint64{1, 2, 3},
			},
			1,
			true,
		},
		{
			"allow - any place",
			cylinder.Config{
				GroupIDs: []uint64{1, 2, 3},
			},
			2,
			true,
		},
		{
			"not allow - empty list",
			cylinder.Config{
				GroupIDs: nil,
			},
			4,
			false,
		},
		{
			"not allow - not in list",
			cylinder.Config{
				GroupIDs: []uint64{1, 2, 3},
			},
			4,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allow := test.cfg.IsAllowedGroup(test.groupID)
			assert.Equal(t, test.expect, allow)
		})
	}
}
