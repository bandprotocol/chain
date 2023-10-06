package proof

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash              common.Hash
	MintStoreMerkleHash              common.Hash
	ParamsToSlashingStoresMerkleHash common.Hash
	GovToIcahostStoresMerkleHash     common.Hash
	AuthToFeegrantStoresMerkleHash   common.Hash
	StakingToUpgradeStoresMerkleHash common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:              common.BytesToHash(m.OracleIAVLStateHash),
		MintStoreMerkleHash:              common.BytesToHash(m.MintStoreMerkleHash),
		ParamsToSlashingStoresMerkleHash: common.BytesToHash(m.ParamsToSlashingStoresMerkleHash),
		GovToIcahostStoresMerkleHash:     common.BytesToHash(m.GovToIcahostStoresMerkleHash),
		AuthToFeegrantStoresMerkleHash:   common.BytesToHash(m.AuthToFeegrantStoresMerkleHash),
		StakingToUpgradeStoresMerkleHash: common.BytesToHash(m.StakingToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:              tmbytes.HexBytes(multiStoreEp.Value),
		MintStoreMerkleHash:              tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		ParamsToSlashingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		GovToIcahostStoresMerkleHash:     tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToFeegrantStoresMerkleHash:   tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		StakingToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
