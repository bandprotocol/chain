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
			types.ReferenceSourceConfig{RegistryIPFSHash: "", RegistryVersion: "1.0.0"},
			fmt.Errorf("registry ipfs hash cannot be empty"),
		},
		{
			"empty version",
			types.ReferenceSourceConfig{RegistryIPFSHash: "hash", RegistryVersion: ""},
			fmt.Errorf("registry version cannot be empty"),
		},
		{
			"wrong version format",
			types.ReferenceSourceConfig{RegistryIPFSHash: "hash", RegistryVersion: "hash"},
			fmt.Errorf("registry version is not in a valid version format"),
		},
		{
			"pre-release version",
			types.ReferenceSourceConfig{RegistryIPFSHash: "hash", RegistryVersion: "0.0.1-alpha.3"},
			nil,
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
