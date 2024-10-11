package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestReferenceSourceConfig_Validate(t *testing.T) {
	tests := []struct {
		name                  string
		referenceSourceConfig types.ReferenceSourceConfig
		wantErr               error
	}{
		{"default reference source config", types.DefaultReferenceSourceConfig(), nil},
		{
			"empty IPFS hash",
			types.ReferenceSourceConfig{IPFSHash: "", Version: "1.0.0"},
			fmt.Errorf("ipfs hash cannot be empty"),
		},
		{
			"empty version",
			types.ReferenceSourceConfig{IPFSHash: "hash", Version: ""},
			fmt.Errorf("version cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.referenceSourceConfig.Validate()
			if tt.wantErr == nil {
				require.NoError(t, got)
				return
			}
			require.Equal(t, tt.wantErr, got)
		})
	}
}
