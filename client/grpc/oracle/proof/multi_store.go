package proof

import (
	"github.com/ethereum/go-ethereum/common"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	ics23 "github.com/cosmos/ics23/go"
)

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash              common.Hash
	ParamsStoreMerkleHash            common.Hash
	IcahostToMintStoresMerkleHash    common.Hash
	RestakeToStakingStoresMerkleHash common.Hash
	TransferToUpgradeStoreMerkleHash common.Hash
	AuthToIbcStoresMerkleHash        common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:              common.BytesToHash(m.OracleIAVLStateHash),
		ParamsStoreMerkleHash:            common.BytesToHash(m.ParamsStoreMerkleHash),
		IcahostToMintStoresMerkleHash:    common.BytesToHash(m.IcahostToMintStoresMerkleHash),
		RestakeToStakingStoresMerkleHash: common.BytesToHash(m.RestakeToStakingStoresMerkleHash),
		TransferToUpgradeStoreMerkleHash: common.BytesToHash(m.TransferToUpgradeStoreMerkleHash),
		AuthToIbcStoresMerkleHash:        common.BytesToHash(m.AuthToIbcStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:              tmbytes.HexBytes(multiStoreEp.Value),
		ParamsStoreMerkleHash:            tmbytes.HexBytes(multiStoreEp.Path[0].Suffix),
		IcahostToMintStoresMerkleHash:    tmbytes.HexBytes(multiStoreEp.Path[1].Prefix[1:]),
		RestakeToStakingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
		TransferToUpgradeStoreMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[3].Suffix),
		AuthToIbcStoresMerkleHash:        tmbytes.HexBytes(multiStoreEp.Path[4].Prefix[1:]),
	}
}
