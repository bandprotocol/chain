package proof

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash               common.Hash
	ParamsStoreMerkleHash             common.Hash
	SlashingToStakingStoresMerkleHash common.Hash
	TransferToUpgradeStoresMerkleHash common.Hash
	AuthToMintStoresMerkleHash        common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:               common.BytesToHash(m.OracleIAVLStateHash),
		ParamsStoreMerkleHash:             common.BytesToHash(m.ParamsStoreMerkleHash),
		SlashingToStakingStoresMerkleHash: common.BytesToHash(m.SlashingToStakingStoresMerkleHash),
		TransferToUpgradeStoresMerkleHash: common.BytesToHash(m.TransferToUpgradeStoresMerkleHash),
		AuthToMintStoresMerkleHash:        common.BytesToHash(m.AuthToMintStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:               tmbytes.HexBytes(multiStoreEp.Value),
		ParamsStoreMerkleHash:             tmbytes.HexBytes(multiStoreEp.Path[0].Suffix),
		SlashingToStakingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		TransferToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
		AuthToMintStoresMerkleHash:        tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
	}
}
