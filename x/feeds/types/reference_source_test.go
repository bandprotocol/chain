package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestReferenceSourceConfig_Validate(t *testing.T) {
	tests := []struct {
		name                  string
		referenceSourceConfig types.ReferenceSourceConfig
		wantErr               error
	}{
		{"default reference source config", types.DefaultReferenceSourceConfig(), nil},
	}

	for _, tt := range tests {
		tt := tt
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
