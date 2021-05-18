package proof

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/types"
)

/*
{
	"commit": {
        "height": "50000",
        "round": 0,
        "block_id": {
          "hash": "253E2EF603743B5CD0C7E8B8591082190398A16032ED2FF096F854033D106F4E",
          "parts": {
            "total": 1,
            "hash": "2EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A8637"
          }
        },
        "signatures": [
          {
            "block_id_flag": 2,
            "validator_address": "F23391B5DBF982E37FB7DADEA64AAE21CAE4C172",
            "timestamp": "2021-05-14T12:17:12.271008775Z",
            "signature": "mLo7n09R88zVZatlB79S9SrJY1XOdt0SmpRScYLoM+B6L8fR7nyZHu85KFFrDJ7lfEncX5FshEV1na3rS2XEXA=="
          },
          {
            "block_id_flag": 2,
            "validator_address": "F0C23921727D869745C4F9703CF33996B1D2B715",
            "timestamp": "2021-05-14T12:17:12.271817895Z",
            "signature": "s1hhTdXndnssshRV5S5fFdBbnuGpd0Zx2uNl9gXIY005SMb9cUYPg6bHQUv1TdTVmLNo9PdbRXPAQHdQiq/w0Q=="
          },
          {
            "block_id_flag": 2,
            "validator_address": "5179B0BB203248E03D2A1342896133B5C58E1E44",
            "timestamp": "2021-05-14T12:17:12.260335655Z",
            "signature": "AGKoHYRS0gFuZdUsFSAfJS8vVUmbexNtPUoo0WK1OH5JirxQdD/QRoCn+GBbc+41WLPHLcFUGFCIYmQYZElErA=="
          },
          {
            "block_id_flag": 2,
            "validator_address": "BDB6A0728C8DFE2124536F16F2BA428FE767A8F9",
            "timestamp": "2021-05-14T12:17:12.264889675Z",
            "signature": "X1IjZf5Hh/Uw0qDbEBVA/feVKWGnyXsrmCQxpK5OTEUGlisVuq4nkXOevWWpRFoqd4SXJDGOMlPirGMHycTWsw=="
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
		Height: 50000,
		Round:  0,
		BlockID: types.BlockID{
			Hash: hexToBytes("253E2EF603743B5CD0C7E8B8591082190398A16032ED2FF096F854033D106F4E"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("2EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A8637"),
			},
		},
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"),
				Timestamp:        parseTime("2021-05-14T12:17:12.271008775Z"),
				Signature:        base64ToBytes("mLo7n09R88zVZatlB79S9SrJY1XOdt0SmpRScYLoM+B6L8fR7nyZHu85KFFrDJ7lfEncX5FshEV1na3rS2XEXA=="),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F0C23921727D869745C4F9703CF33996B1D2B715"),
				Timestamp:        parseTime("2021-05-14T12:17:12.271817895Z"),
				Signature:        base64ToBytes("s1hhTdXndnssshRV5S5fFdBbnuGpd0Zx2uNl9gXIY005SMb9cUYPg6bHQUv1TdTVmLNo9PdbRXPAQHdQiq/w0Q=="),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("5179B0BB203248E03D2A1342896133B5C58E1E44"),
				Timestamp:        parseTime("2021-05-14T12:17:12.260335655Z"),
				Signature:        base64ToBytes("AGKoHYRS0gFuZdUsFSAfJS8vVUmbexNtPUoo0WK1OH5JirxQdD/QRoCn+GBbc+41WLPHLcFUGFCIYmQYZElErA=="),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("BDB6A0728C8DFE2124536F16F2BA428FE767A8F9"),
				Timestamp:        parseTime("2021-05-14T12:17:12.264889675Z"),
				Signature:        base64ToBytes("X1IjZf5Hh/Uw0qDbEBVA/feVKWGnyXsrmCQxpK5OTEUGlisVuq4nkXOevWWpRFoqd4SXJDGOMlPirGMHycTWsw=="),
			},
		},
	}
	sh := types.SignedHeader{
		Header: &header,
		Commit: &commit,
	}
	sig, err := GetSignaturesAndPrefix(&sh)
	require.NoError(t, err)

	expected := []TMSignature{
		{
			R:                hexToBytes("5F522365FE4787F530D2A0DB101540FDF7952961A7C97B2B982431A4AE4E4C45"),
			S:                hexToBytes("06962B15BAAE2791739EBD65A9445A2A77849724318E3253E2AC6307C9C4D6B3"),
			V:                27,
			SignedDataPrefix: hexToBytes("6D08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0B08C8D4F9840610CBCAA77E320962616E64636861696E"),
		},
		{
			R:                hexToBytes("0062A81D8452D2016E65D52C15201F252F2F55499B7B136D3D4A28D162B5387E"),
			S:                hexToBytes("498ABC50743FD04680A7F8605B73EE3558B3C72DC154185088626418644944AC"),
			V:                28,
			SignedDataPrefix: hexToBytes("6D08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0B08C8D4F9840610A7D0917C320962616E64636861696E"),
		},
		{
			R:                hexToBytes("98BA3B9F4F51F3CCD565AB6507BF52F52AC96355CE76DD129A94527182E833E0"),
			S:                hexToBytes("7A2FC7D1EE7C991EEF3928516B0C9EE57C49DC5F916C8445759DADEB4B65C45C"),
			V:                27,
			SignedDataPrefix: hexToBytes("6E08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0C08C8D4F984061087889D8101320962616E64636861696E"),
		},
		{
			R:                hexToBytes("B358614DD5E7767B2CB21455E52E5F15D05B9EE1A9774671DAE365F605C8634D"),
			S:                hexToBytes("3948C6FD71460F83A6C7414BF54DD4D598B368F4F75B4573C04077508AAFF0D1"),
			V:                27,
			SignedDataPrefix: hexToBytes("6E08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0C08C8D4F9840610A7B9CE8101320962616E64636861696E"),
		},
	}
	require.Equal(t, expected, sig)
}

func TestVerifySignature(t *testing.T) {
	signatures := []TMSignature{
		{
			R:                hexToBytes("5F522365FE4787F530D2A0DB101540FDF7952961A7C97B2B982431A4AE4E4C45"),
			S:                hexToBytes("06962B15BAAE2791739EBD65A9445A2A77849724318E3253E2AC6307C9C4D6B3"),
			V:                27,
			SignedDataPrefix: hexToBytes("6D08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0B08C8D4F9840610CBCAA77E320962616E64636861696E"),
		},
		{
			R:                hexToBytes("0062A81D8452D2016E65D52C15201F252F2F55499B7B136D3D4A28D162B5387E"),
			S:                hexToBytes("498ABC50743FD04680A7F8605B73EE3558B3C72DC154185088626418644944AC"),
			V:                28,
			SignedDataPrefix: hexToBytes("6D08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0B08C8D4F9840610A7D0917C320962616E64636861696E"),
		},
		{
			R:                hexToBytes("98BA3B9F4F51F3CCD565AB6507BF52F52AC96355CE76DD129A94527182E833E0"),
			S:                hexToBytes("7A2FC7D1EE7C991EEF3928516B0C9EE57C49DC5F916C8445759DADEB4B65C45C"),
			V:                27,
			SignedDataPrefix: hexToBytes("6E08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0C08C8D4F984061087889D8101320962616E64636861696E"),
		},
		{
			R:                hexToBytes("B358614DD5E7767B2CB21455E52E5F15D05B9EE1A9774671DAE365F605C8634D"),
			S:                hexToBytes("3948C6FD71460F83A6C7414BF54DD4D598B368F4F75B4573C04077508AAFF0D1"),
			V:                27,
			SignedDataPrefix: hexToBytes("6E08021150C300000000000022480A20"),
			SignedDataSuffix: hexToBytes("1224080112202EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A86372A0C08C8D4F9840610A7B9CE8101320962616E64636861696E"),
		},
	}

	evmAddresses := []common.Address{
		common.HexToAddress("0x652D89a66Eb4eA55366c45b1f9ACfc8e2179E1c5"),
		common.HexToAddress("0x88e1cd00710495EEB93D4f522d16bC8B87Cb00FE"),
		common.HexToAddress("0xaAA22E077492CbaD414098EBD98AA8dc1C7AE8D9"),
		common.HexToAddress("0xB956589b6fC5523eeD0d9eEcfF06262Ce84ff260"),
	}

	blockHash := hexToBytes("253E2EF603743B5CD0C7E8B8591082190398A16032ED2FF096F854033D106F4E")

	for i, sig := range signatures {
		msg := append(sig.SignedDataPrefix, blockHash...)
		msg = append(msg, sig.SignedDataSuffix...)

		sigBytes := append(sig.R, sig.S...)
		sigBytes = append(sigBytes, sig.V-27)
		pub, err := crypto.SigToPub(tmhash.Sum(msg), sigBytes)
		require.Nil(t, err)
		require.Equal(t, evmAddresses[i], crypto.PubkeyToAddress(*pub))
	}
}
