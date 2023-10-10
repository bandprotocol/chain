package proof

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash             common.Hash
	MintStoreMerkleHash             common.Hash
	IbcToIcahostStoresMerkleHash    common.Hash
	FeegrantToGroupStoresMerkleHash common.Hash
	AuthToEvidenceStoresMerkleHash  common.Hash
	ParamsToUpgradeStoresMerkleHash common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:             common.BytesToHash(m.OracleIAVLStateHash),
		MintStoreMerkleHash:             common.BytesToHash(m.MintStoreMerkleHash),
		IbcToIcahostStoresMerkleHash:    common.BytesToHash(m.IbcToIcahostStoresMerkleHash),
		FeegrantToGroupStoresMerkleHash: common.BytesToHash(m.FeegrantToGroupStoresMerkleHash),
		AuthToEvidenceStoresMerkleHash:  common.BytesToHash(m.AuthToEvidenceStoresMerkleHash),
		ParamsToUpgradeStoresMerkleHash: common.BytesToHash(m.ParamsToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:             tmbytes.HexBytes(multiStoreEp.Value),
		MintStoreMerkleHash:             tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		IbcToIcahostStoresMerkleHash:    tmbytes.HexBytes(multiStoreEp.Path[1].Prefix[1:]),
		FeegrantToGroupStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToEvidenceStoresMerkleHash:  tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		ParamsToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
