package bandibctesting

import (
	"testing"
	"time"

	tmtypes "github.com/cometbft/cometbft/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/require"
)

var (
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	ValSenders map[string]*authtypes.BaseAccount
	ValSigners map[string]cryptotypes.PrivKey
)

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, n int) *ibctesting.Coordinator {
	chains := make(map[string]*ibctesting.TestChain)
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: globalStartTime,
	}

	for i := 1; i <= n; i++ {
		chainID := ibctesting.GetChainID(i)
		chains[chainID] = NewTestChain(t, coord, chainID)
	}

	coord.Chains = chains

	return coord
}

// NewTestChain initializes a new test chain with a default of 4 validators
// Use this function if the tests do not need custom control over the validator set
func NewTestChain(t *testing.T, coord *ibctesting.Coordinator, chainID string) *ibctesting.TestChain {
	// generate validators private/public key
	var (
		validatorsPerChain = uint64(2)
		validators         []*tmtypes.Validator
		signersByAddress   = make(map[string]tmtypes.PrivValidator, validatorsPerChain)
		valSenders         = make(map[string]*authtypes.BaseAccount, validatorsPerChain)
		valSigners         = make(map[string]cryptotypes.PrivKey, validatorsPerChain)
	)

	for i := uint64(0); i < validatorsPerChain; i++ {
		privVal := NewPV()
		pubKey, err := privVal.GetPubKey()
		require.NoError(t, err)
		validators = append(validators, tmtypes.NewValidator(pubKey, 1))
		signersByAddress[pubKey.Address().String()] = privVal
		valSigners[pubKey.Address().String()] = privVal.PrivKey

		valSenders[pubKey.Address().String()] = authtypes.NewBaseAccount(
			pubKey.Address().Bytes(),
			privVal.PrivKey.PubKey(),
			i,
			0,
		)

		ValSenders = valSenders
		ValSigners = valSigners
	}

	// construct validator set;
	// Note that the validators are sorted by voting power
	// or, if equal, by address lexical order
	valSet := tmtypes.NewValidatorSet(validators)

	return ibctesting.NewTestChainWithValSet(t, coord, chainID, valSet, signersByAddress)
}
