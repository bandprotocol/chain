package proof

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash                 common.Hash
	MintStoreMerkleHash                 common.Hash
	ParamsToRollingseedStoresMerkleHash common.Hash
	SlashingToTssStoresMerkleHash       common.Hash
	UpgradeStoreMerkleHash              common.Hash
	AuthToIcahostStoresMerkleHash       common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:                 common.BytesToHash(m.OracleIAVLStateHash),
		MintStoreMerkleHash:                 common.BytesToHash(m.MintStoreMerkleHash),
		ParamsToRollingseedStoresMerkleHash: common.BytesToHash(m.ParamsToRollingseedStoresMerkleHash),
		SlashingToTssStoresMerkleHash:       common.BytesToHash(m.SlashingToTssStoresMerkleHash),
		UpgradeStoreMerkleHash:              common.BytesToHash(m.UpgradeStoreMerkleHash),
		AuthToIcahostStoresMerkleHash:       common.BytesToHash(m.AuthToIcahostStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:                 tmbytes.HexBytes(multiStoreEp.Value),
		MintStoreMerkleHash:                 tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		ParamsToRollingseedStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		SlashingToTssStoresMerkleHash:       tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
		UpgradeStoreMerkleHash:              tmbytes.HexBytes(multiStoreEp.Path[3].Suffix),
		AuthToIcahostStoresMerkleHash:       tmbytes.HexBytes(multiStoreEp.Path[4].Prefix[1:]),
	}
}
