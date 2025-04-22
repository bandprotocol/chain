package proof

import (
	"github.com/ethereum/go-ethereum/common"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	ics23 "github.com/cosmos/ics23/go"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash                   common.Hash
	MintStoreMerkleHash                   common.Hash
	ParamsToRestakeStoresMerkleHash       common.Hash
	RollingseedToTransferStoresMerkleHash common.Hash
	TSSToUpgradeStoresMerkleHash          common.Hash
	AuthToIcahostStoresMerkleHash         common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:                   common.BytesToHash(m.OracleIAVLStateHash),
		MintStoreMerkleHash:                   common.BytesToHash(m.MintStoreMerkleHash),
		ParamsToRestakeStoresMerkleHash:       common.BytesToHash(m.ParamsToRestakeStoresMerkleHash),
		RollingseedToTransferStoresMerkleHash: common.BytesToHash(m.RollingseedToTransferStoresMerkleHash),
		TSSToUpgradeStoresMerkleHash:          common.BytesToHash(m.TSSToUpgradeStoresMerkleHash),
		AuthToIcahostStoresMerkleHash:         common.BytesToHash(m.AuthToIcahostStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:                   tmbytes.HexBytes(multiStoreEp.Value),
		MintStoreMerkleHash:                   tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		ParamsToRestakeStoresMerkleHash:       tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		RollingseedToTransferStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
		TSSToUpgradeStoresMerkleHash:          tmbytes.HexBytes(multiStoreEp.Path[3].Suffix),
		AuthToIcahostStoresMerkleHash:         tmbytes.HexBytes(multiStoreEp.Path[4].Prefix[1:]),
	}
}
