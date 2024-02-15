package proof

import (
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/types"
	"github.com/ethereum/go-ethereum/common"
)

// BlockHeaderMerklePartsEthereum is an Ethereum version of BlockHeaderMerkleParts for solidity ABI-encoding.
type BlockHeaderMerklePartsEthereum struct {
	VersionAndChainIdHash             common.Hash
	Height                            uint64
	TimeSecond                        uint64
	TimeNanoSecond                    uint32
	LastBlockIdAndOther               common.Hash
	NextValidatorHashAndConsensusHash common.Hash
	LastResultsHash                   common.Hash
	EvidenceAndProposerHash           common.Hash
}

func (bp *BlockHeaderMerkleParts) encodeToEthFormat() BlockHeaderMerklePartsEthereum {
	return BlockHeaderMerklePartsEthereum{
		VersionAndChainIdHash:             common.BytesToHash(bp.VersionAndChainIdHash),
		Height:                            bp.Height,
		TimeSecond:                        bp.TimeSecond,
		TimeNanoSecond:                    bp.TimeNanoSecond,
		LastBlockIdAndOther:               common.BytesToHash(bp.LastBlockIdAndOther),
		NextValidatorHashAndConsensusHash: common.BytesToHash(bp.NextValidatorHashAndConsensusHash),
		LastResultsHash:                   common.BytesToHash(bp.LastResultsHash),
		EvidenceAndProposerHash:           common.BytesToHash(bp.EvidenceAndProposerHash),
	}
}

// GetBlockHeaderMerkleParts converts Tendermint block header struct into BlockHeaderMerkleParts for gas-optimized proof verification.
func GetBlockHeaderMerkleParts(block *types.Header) BlockHeaderMerkleParts {
	// based on https://github.com/tendermint/tendermint/blob/master/types/block.go#L448
	hbz, err := block.Version.Marshal()
	if err != nil {
		panic(err)
	}

	pbbi := block.LastBlockID.ToProto()
	bzbi, err := pbbi.Marshal()
	if err != nil {
		panic(err)
	}

	return BlockHeaderMerkleParts{
		VersionAndChainIdHash: merkle.HashFromByteSlices([][]byte{
			hbz,
			cdcEncode(block.ChainID),
		}),
		Height:         uint64(block.Height),
		TimeSecond:     uint64(block.Time.Unix()),
		TimeNanoSecond: uint32(block.Time.Nanosecond()),
		LastBlockIdAndOther: merkle.HashFromByteSlices([][]byte{
			bzbi,
			cdcEncode(block.LastCommitHash),
			cdcEncode(block.DataHash),
			cdcEncode(block.ValidatorsHash),
		}),
		NextValidatorHashAndConsensusHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.NextValidatorsHash),
			cdcEncode(block.ConsensusHash),
		}),
		LastResultsHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.LastResultsHash),
		}),
		EvidenceAndProposerHash: merkle.HashFromByteSlices([][]byte{
			cdcEncode(block.EvidenceHash),
			cdcEncode(block.ProposerAddress),
		}),
	}
}
