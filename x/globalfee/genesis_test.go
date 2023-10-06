package globalfee

import (
	"testing"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	gotJSON := AppModuleBasic{}.DefaultGenesis(encCfg.Codec)
	assert.JSONEq(t, `{"params":{"minimum_gas_prices":[]}}`, string(gotJSON))
}

func TestValidateGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	specs := map[string]struct {
		src    string
		expErr bool
	}{
		"all good": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}]}}`,
		},
		"empty minimum": {
			src: `{"params":{"minimum_gas_prices":[]}}`,
		},
		"minimum not set": {
			src: `{"params":{}}`,
		},
		"zero amount not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"0"}]}}`,
			expErr: true,
		},
		"duplicate denoms not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"negative amounts not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"-1"}]}}`,
			expErr: true,
		},
		"denom must be sorted": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ZLX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"sorted denoms is allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ZLX", "amount":"2"}]}}`,
			expErr: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := AppModuleBasic{}.ValidateGenesis(encCfg.Codec, nil, []byte(spec.src))
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
