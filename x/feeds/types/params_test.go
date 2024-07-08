package types_test

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()
	require.Equal(t, p1, p2)

	p1.MaxInterval += 10
	require.NotEqual(t, p1, p2)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		wantErr error
	}{
		{"default params", types.DefaultParams(), nil},
		{"invalid Admin", func() types.Params {
			params := types.DefaultParams()
			params.Admin = "" // Invalid value
			return params
		}(), fmt.Errorf("admin cannot be empty")},
		{"invalid AllowableBlockTimeDiscrepancy", func() types.Params {
			params := types.DefaultParams()
			params.AllowableBlockTimeDiscrepancy = 0 // Invalid value
			return params
		}(), fmt.Errorf("allowable block time discrepancy must be positive: 0")},
		{"invalid GracePeriod", func() types.Params {
			params := types.DefaultParams()
			params.GracePeriod = -1 // Invalid value
			return params
		}(), fmt.Errorf("grace period must be positive: -1")},
		{"invalid MinInterval", func() types.Params {
			params := types.DefaultParams()
			params.MinInterval = 0 // Invalid value
			return params
		}(), fmt.Errorf("min interval must be positive: 0")},
		{"invalid MaxInterval", func() types.Params {
			params := types.DefaultParams()
			params.MaxInterval = 0 // Invalid value
			return params
		}(), fmt.Errorf("max interval must be positive: 0")},
		{"invalid PowerStepThreshold", func() types.Params {
			params := types.DefaultParams()
			params.PowerStepThreshold = -10 // Invalid value
			return params
		}(), fmt.Errorf("power threshold must be positive: -10")},
		{"invalid MaxSupportedFeeds", func() types.Params {
			params := types.DefaultParams()
			params.MaxSupportedFeeds = 0 // Invalid value
			return params
		}(), fmt.Errorf("max supported feeds must be positive: 0")},
		{"invalid CooldownTime", func() types.Params {
			params := types.DefaultParams()
			params.CooldownTime = -5 // Invalid value
			return params
		}(), fmt.Errorf("cooldown time must be positive: -5")},
		{"invalid MinDeviationBasisPoint", func() types.Params {
			params := types.DefaultParams()
			params.MinDeviationBasisPoint = 0 // Invalid value
			return params
		}(), fmt.Errorf("min deviation basis point must be positive: 0")},
		{"invalid MaxDeviationBasisPoint", func() types.Params {
			params := types.DefaultParams()
			params.MaxDeviationBasisPoint = 0 // Invalid value
			return params
		}(), fmt.Errorf("max deviation basis point must be positive: 0")},
		{"invalid SupportedFeedsUpdateInterval", func() types.Params {
			params := types.DefaultParams()
			params.SupportedFeedsUpdateInterval = 0 // Invalid value
			return params
		}(), fmt.Errorf("supported feeds update interval must be positive: 0")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.Validate()
			if tt.wantErr == nil {
				require.NoError(t, got)
				return
			}
			require.Equal(t, tt.wantErr.Error(), got.Error())
		})
	}
}
