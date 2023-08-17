package band

import (
	"fmt"
	"io/ioutil"

	tmtypes "github.com/cometbft/cometbft/types"
)

const (
	flagGenesisTime   = "genesis-time"
	flagChainID       = "chain-id"
	flagInitialHeight = "initial-height"
)

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*tmtypes.GenesisDoc, error) {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc file: %w", err)
	}
	genDoc, err := tmtypes.GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading GenesisDoc at %s: %w", genDocFile, err)
	}

	genDoc.ConsensusParams.Block.MaxBytes = 3000000 // 3M bytes
	genDoc.ConsensusParams.Block.MaxGas = 50000000  // 50M gas

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return genDoc, nil
}
