package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	p2 := DefaultParams()
	require.Equal(t, p1, p2)

	p1.AllowedDenoms = []string{"uband"}
	require.NotEqual(t, p1, p2)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		expErr bool
	}{
		{
			"default params",
			DefaultParams(),
			false,
		},
		{
			"invalid denom",
			Params{
				AllowedDenoms: []string{""},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.params.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
