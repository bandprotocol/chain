package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestPriceService_Validate(t *testing.T) {
	tests := []struct {
		name         string
		priceService types.PriceService
		wantErr      error
	}{
		{"default price service", types.DefaultPriceService(), nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.priceService.Validate()
			if tt.wantErr == nil {
				require.NoError(t, got)
				return
			}
			require.Equal(t, tt.wantErr, got)
		})
	}
}
