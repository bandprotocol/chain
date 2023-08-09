package proof

import (
	"encoding/hex"
	"testing"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestGetPrefix(t *testing.T) {
	prefix, err := GetPrefix(tmproto.SignedMsgType(2), 25000, 0)
	require.NoError(t, err)
	require.Equal(t, "080211a861000000000000", hex.EncodeToString(prefix))

	prefix, err = GetPrefix(tmproto.SignedMsgType(2), 25000, 1)
	require.NoError(t, err)
	require.Equal(t, "080211a861000000000000190100000000000000", hex.EncodeToString(prefix))
}

/*
{
	commit: {
		height: "25000",
		round: 0,
		block_id: {
			hash: "3489F21785ACE1CE4214CB2B57F3A98DC0B7377D1BA1E1180B6E199E33B0FC5A",
			parts: {
				total: 1,
				hash: "6BF91EFBA26A4CD86EBBD0E54DCFC9BD2C790859CFA96215661A47E4921A6301"
			}
		},
		signatures: [
			{
				block_id_flag: 2,
				validator_address: "5179B0BB203248E03D2A1342896133B5C58E1E44",
				timestamp: "2021-08-25T00:05:33.107055466Z",
				signature: "OUNlGT+BnPU5OBNm0xtsWEmqoxroum+VxixcgGVr+1xqB+SjwKvOrl+FTUkt9plDj7hHYvFS9znd6sSN3Py1zA=="
			},
			{
				block_id_flag: 2,
				validator_address: "BDB6A0728C8DFE2124536F16F2BA428FE767A8F9",
				timestamp: "2021-08-25T00:05:33.128300266Z",
				signature: "hLhYW3EkD+4OZ0lSt57SXXk/GzG0LdN7gPdbmFELV1QexE3XxTiUdN+OXCXMbti1c8yi4Amqgk7oJb3Gk5NZJw=="
			},
			{
				block_id_flag: 2,
				validator_address: "F0C23921727D869745C4F9703CF33996B1D2B715",
				timestamp: "2021-08-25T00:05:33.108916976Z",
				signature: "VlSkT7iTMMNM8thi+UB2MZShRbcu07sK3VdZ4eaP0UUqx5XQKpxXTPEjQ/38Z/3O2KJPiOyBOMf4Iw9utEK3Jg=="
			},
			{
				block_id_flag: 2,
				validator_address: "F23391B5DBF982E37FB7DADEA64AAE21CAE4C172",
				timestamp: "2021-08-25T00:05:33.120372486Z",
				signature: "XXtL57IbANCK19vkjPJ2HOzLWZ5kqrELKQGg3VjwAyVxYO9omlM8Hpg3B1B/yEZtrqHQ3HqInjon0bsdCc7AMA=="
			}
		]
	}
}
*/

func TestGetSignaturesAndPrefix(t *testing.T) {
	header := types.Header{
		ChainID: "bandchain",
	}
	commit := types.Commit{
		Height: 25000,
		Round:  0,
		BlockID: types.BlockID{
			Hash: hexToBytes("3489F21785ACE1CE4214CB2B57F3A98DC0B7377D1BA1E1180B6E199E33B0FC5A"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("6BF91EFBA26A4CD86EBBD0E54DCFC9BD2C790859CFA96215661A47E4921A6301"),
			},
		},
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("5179B0BB203248E03D2A1342896133B5C58E1E44"),
				Timestamp:        parseTime("2021-08-25T00:05:33.107055466Z"),
				Signature: base64ToBytes(
					"OUNlGT+BnPU5OBNm0xtsWEmqoxroum+VxixcgGVr+1xqB+SjwKvOrl+FTUkt9plDj7hHYvFS9znd6sSN3Py1zA==",
				),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("BDB6A0728C8DFE2124536F16F2BA428FE767A8F9"),
				Timestamp:        parseTime("2021-08-25T00:05:33.128300266Z"),
				Signature: base64ToBytes(
					"hLhYW3EkD+4OZ0lSt57SXXk/GzG0LdN7gPdbmFELV1QexE3XxTiUdN+OXCXMbti1c8yi4Amqgk7oJb3Gk5NZJw==",
				),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F0C23921727D869745C4F9703CF33996B1D2B715"),
				Timestamp:        parseTime("2021-08-25T00:05:33.108916976Z"),
				Signature: base64ToBytes(
					"VlSkT7iTMMNM8thi+UB2MZShRbcu07sK3VdZ4eaP0UUqx5XQKpxXTPEjQ/38Z/3O2KJPiOyBOMf4Iw9utEK3Jg==",
				),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"),
				Timestamp:        parseTime("2021-08-25T00:05:33.120372486Z"),
				Signature: base64ToBytes(
					"XXtL57IbANCK19vkjPJ2HOzLWZ5kqrELKQGg3VjwAyVxYO9omlM8Hpg3B1B/yEZtrqHQ3HqInjon0bsdCc7AMA==",
				),
			},
		},
	}
	sh := types.SignedHeader{
		Header: &header,
		Commit: &commit,
	}

	sig, commonVote, err := GetSignaturesAndPrefix(&sh)
	require.NoError(t, err)

	expectedSigs := []TMSignature{
		{
			R:                hexToBytes("84B8585B71240FEE0E674952B79ED25D793F1B31B42DD37B80F75B98510B5754"),
			S:                hexToBytes("1EC44DD7C5389474DF8E5C25CC6ED8B573CCA2E009AA824EE825BDC693935927"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610EAE9963D"),
		},
		{
			R:                hexToBytes("394365193F819CF539381366D31B6C5849AAA31AE8BA6F95C62C5C80656BFB5C"),
			S:                hexToBytes("6A07E4A3C0ABCEAE5F854D492DF699438FB84762F152F739DDEAC48DDCFCB5CC"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610EA928633"),
		},
		{
			R:                hexToBytes("5D7B4BE7B21B00D08AD7DBE48CF2761CECCB599E64AAB10B2901A0DD58F00325"),
			S:                hexToBytes("7160EF689A533C1E983707507FC8466DAEA1D0DC7A889E3A27D1BB1D09CEC030"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD929689061086FAB239"),
		},
		{
			R:                hexToBytes("5654A44FB89330C34CF2D862F940763194A145B72ED3BB0ADD5759E1E68FD145"),
			S:                hexToBytes("2AC795D02A9C574CF12343FDFC67FDCED8A24F88EC8138C7F8230F6EB442B726"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610F0E1F733"),
		},
	}
	expectedCommonVote := CommonEncodedVotePart{
		SignedDataPrefix: hexToBytes("080211A86100000000000022480A20"),
		SignedDataSuffix: hexToBytes("1224080112206BF91EFBA26A4CD86EBBD0E54DCFC9BD2C790859CFA96215661A47E4921A6301"),
	}

	require.Equal(t, expectedSigs, sig)
	require.Equal(t, expectedCommonVote, commonVote)
}

func TestVerifySignature(t *testing.T) {
	signatures := []TMSignature{
		{
			R:                hexToBytes("84B8585B71240FEE0E674952B79ED25D793F1B31B42DD37B80F75B98510B5754"),
			S:                hexToBytes("1EC44DD7C5389474DF8E5C25CC6ED8B573CCA2E009AA824EE825BDC693935927"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610EAE9963D"),
		},
		{
			R:                hexToBytes("394365193F819CF539381366D31B6C5849AAA31AE8BA6F95C62C5C80656BFB5C"),
			S:                hexToBytes("6A07E4A3C0ABCEAE5F854D492DF699438FB84762F152F739DDEAC48DDCFCB5CC"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610EA928633"),
		},
		{
			R:                hexToBytes("5D7B4BE7B21B00D08AD7DBE48CF2761CECCB599E64AAB10B2901A0DD58F00325"),
			S:                hexToBytes("7160EF689A533C1E983707507FC8466DAEA1D0DC7A889E3A27D1BB1D09CEC030"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD929689061086FAB239"),
		},
		{
			R:                hexToBytes("5654A44FB89330C34CF2D862F940763194A145B72ED3BB0ADD5759E1E68FD145"),
			S:                hexToBytes("2AC795D02A9C574CF12343FDFC67FDCED8A24F88EC8138C7F8230F6EB442B726"),
			V:                28,
			EncodedTimestamp: hexToBytes("08CD9296890610F0E1F733"),
		},
	}
	commonVote := CommonEncodedVotePart{
		SignedDataPrefix: hexToBytes("080211A86100000000000022480A20"),
		SignedDataSuffix: hexToBytes("1224080112206BF91EFBA26A4CD86EBBD0E54DCFC9BD2C790859CFA96215661A47E4921A6301"),
	}

	evmAddresses := []common.Address{
		common.HexToAddress("0x652D89a66Eb4eA55366c45b1f9ACfc8e2179E1c5"),
		common.HexToAddress("0x88e1cd00710495EEB93D4f522d16bC8B87Cb00FE"),
		common.HexToAddress("0xaAA22E077492CbaD414098EBD98AA8dc1C7AE8D9"),
		common.HexToAddress("0xB956589b6fC5523eeD0d9eEcfF06262Ce84ff260"),
	}

	blockHash := hexToBytes("3489F21785ACE1CE4214CB2B57F3A98DC0B7377D1BA1E1180B6E199E33B0FC5A")
	commonPart := append(commonVote.SignedDataPrefix, append(blockHash, commonVote.SignedDataSuffix...)...)
	encodedChainIDConstant := hexToBytes("320962616e64636861696e")

	for i, sig := range signatures {
		msg := append(commonPart, []byte{42, uint8(len(sig.EncodedTimestamp))}...)
		msg = append(msg, sig.EncodedTimestamp...)
		msg = append(msg, encodedChainIDConstant...)
		msg = append([]byte{uint8(len(msg))}, msg...)

		sigBytes := append(sig.R, sig.S...)
		sigBytes = append(sigBytes, uint8(sig.V)-27)
		pub, err := crypto.SigToPub(tmhash.Sum(msg), sigBytes)
		require.Nil(t, err)
		require.Equal(t, evmAddresses[i], crypto.PubkeyToAddress(*pub))
	}
}
