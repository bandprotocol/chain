package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding oracle type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.RequestStoreKeyPrefix):
			var rA, rB types.Request
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)
		case bytes.Equal(kvA.Key[:1], types.ReportStoreKeyPrefix):
			var rA, rB types.Report
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)
		case bytes.Equal(kvA.Key[:1], types.DataSourceStoreKeyPrefix):
			var dsA, dsB types.DataSource
			cdc.MustUnmarshal(kvA.Value, &dsA)
			cdc.MustUnmarshal(kvB.Value, &dsB)
			return fmt.Sprintf("%v\n%v", dsA, dsB)
		case bytes.Equal(kvA.Key[:1], types.OracleScriptStoreKeyPrefix):
			var osA, osB types.OracleScript
			cdc.MustUnmarshal(kvA.Value, &osA)
			cdc.MustUnmarshal(kvB.Value, &osB)
			return fmt.Sprintf("%v\n%v", osA, osB)
		case bytes.Equal(kvA.Key[:1], types.ValidatorStatusKeyPrefix):
			var vsA, vsB types.ValidatorStatus
			cdc.MustUnmarshal(kvA.Value, &vsA)
			cdc.MustUnmarshal(kvB.Value, &vsB)
			return fmt.Sprintf("%v\n%v", vsA, vsB)
		case bytes.Equal(kvA.Key[:1], types.ResultStoreKeyPrefix):
			var rA, rB types.Result
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)
		default:
			panic(fmt.Sprintf("invalid oracle key %X", kvA.Key))
		}
	}
}
