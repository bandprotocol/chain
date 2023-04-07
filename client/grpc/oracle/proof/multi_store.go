package proof

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash               common.Hash
	ParamsStoreMerkleHash             common.Hash
	SlashingToStakingStoresMerkleHash common.Hash
	GovToMintStoresMerkleHash         common.Hash
	AuthToFeegrantStoresMerkleHash    common.Hash
	TransferToUpgradeStoresMerkleHash common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:               common.BytesToHash(m.OracleIAVLStateHash),
		ParamsStoreMerkleHash:             common.BytesToHash(m.ParamsStoreMerkleHash),
		SlashingToStakingStoresMerkleHash: common.BytesToHash(m.SlashingToStakingStoresMerkleHash),
		GovToMintStoresMerkleHash:         common.BytesToHash(m.GovToMintStoresMerkleHash),
		AuthToFeegrantStoresMerkleHash:    common.BytesToHash(m.AuthToFeegrantStoresMerkleHash),
		TransferToUpgradeStoresMerkleHash: common.BytesToHash(m.TransferToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:               tmbytes.HexBytes(multiStoreEp.Value),
		ParamsStoreMerkleHash:             tmbytes.HexBytes(multiStoreEp.Path[0].Suffix),
		SlashingToStakingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		GovToMintStoresMerkleHash:         tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToFeegrantStoresMerkleHash:    tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		TransferToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
