package proof

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/types"
)

/*
commit := types.Commit{
	Height: 1,
	Round:  0,
	BlockID: types.BlockID{
		Hash: hexToBytes("442E3E690F13C2EEBDBFC3DBF80B2373CBBCFA2C12CF17FB9A22A17552A5BF2B"),
		PartSetHeader: types.PartSetHeader{
			Total: 1,
			Hash:  hexToBytes("B9F8E456397FF911BDC55D91B0DBEEDFB59711A87C4223DEB5217C93B9CA60FE"),
		},
	},
	Signatures: []types.CommitSig{
		{
			BlockIDFlag:      2,
			ValidatorAddress: hexToBytes("F0C23921727D869745C4F9703CF33996B1D2B715"),
			Timestamp:        parseTime("2021-06-03T11:05:37.834259369Z"),
			Signature:        base64ToBytes("PqaRaT+IuvAlUqQm+HgTpeDbgUytz4Laxh9ZJOeM7ZtuOrzc2YPoml95KD+rx7TDM1YCc4DM4TjnjxICvrBW0g=="),
		},
		{
			BlockIDFlag:      2,
			ValidatorAddress: hexToBytes("BDB6A0728C8DFE2124536F16F2BA428FE767A8F9"),
			Timestamp:        parseTime("2021-06-03T11:05:37.856577545Z"),
			Signature:        base64ToBytes("x8PIR8F4d1BdC9lACchSQdSNsvQ8ZDHAik0eIaO/Jf9UFhomQHjXYmamikU8IFVpAfUhKpGpxqYqRZMek0n8XQ=="),
		},
		{
			BlockIDFlag:      2,
			ValidatorAddress: hexToBytes("F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"),
			Timestamp:        parseTime("2021-06-03T11:05:37.834252296Z"),
			Signature:        base64ToBytes("db7HnsXg0RG5SpR4hshq+qHLhZZdnyxl+nZCm2SOhHEI1kJbG1vWObtkGmZu6lk4IT+znXfPJHC2dZeanYMTfg=="),
		},
	},
}
*/
func TestGetSignaturesAndPrefix(t *testing.T) {
	header := types.Header{
		ChainID: "odin",
	}
	commit := types.Commit{
		Height: 1,
		Round:  0,
		BlockID: types.BlockID{
			Hash: hexToBytes("442E3E690F13C2EEBDBFC3DBF80B2373CBBCFA2C12CF17FB9A22A17552A5BF2B"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("B9F8E456397FF911BDC55D91B0DBEEDFB59711A87C4223DEB5217C93B9CA60FE"),
			},
		},
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F0C23921727D869745C4F9703CF33996B1D2B715"),
				Timestamp:        parseTime("2021-06-03T11:05:37.834259369Z"),
				Signature:        base64ToBytes("PqaRaT+IuvAlUqQm+HgTpeDbgUytz4Laxh9ZJOeM7ZtuOrzc2YPoml95KD+rx7TDM1YCc4DM4TjnjxICvrBW0g=="),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("BDB6A0728C8DFE2124536F16F2BA428FE767A8F9"),
				Timestamp:        parseTime("2021-06-03T11:05:37.856577545Z"),
				Signature:        base64ToBytes("x8PIR8F4d1BdC9lACchSQdSNsvQ8ZDHAik0eIaO/Jf9UFhomQHjXYmamikU8IFVpAfUhKpGpxqYqRZMek0n8XQ=="),
			},
			{
				BlockIDFlag:      2,
				ValidatorAddress: hexToBytes("F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"),
				Timestamp:        parseTime("2021-06-03T11:05:37.834252296Z"),
				Signature:        base64ToBytes("db7HnsXg0RG5SpR4hshq+qHLhZZdnyxl+nZCm2SOhHEI1kJbG1vWObtkGmZu6lk4IT+znXfPJHC2dZeanYMTfg=="),
			},
		},
	}
	sh := types.SignedHeader{
		Header: &header,
		Commit: &commit,
	}
	sig, err := GetSignaturesAndPrefix(&sh)
	require.NoError(t, err)

	for _, x := range sig {
		fmt.Println(hex.EncodeToString(x.R))
		fmt.Println(hex.EncodeToString(x.S))
		fmt.Println(x.V)
		fmt.Println(hex.EncodeToString(x.SignedDataPrefix))
		fmt.Println(hex.EncodeToString(x.SignedDataSuffix))
	}

	expected := []TMSignature{
		{
			R:                hexToBytes("c7c3c847c17877505d0bd94009c85241d48db2f43c6431c08a4d1e21a3bf25ff"),
			S:                hexToBytes("54161a264078d76266a68a453c20556901f5212a91a9c6a62a45931e9349fc5d"),
			V:                28,
			SignedDataPrefix: hexToBytes("69080211010000000000000022480a20"),
			SignedDataSuffix: hexToBytes("122408011220b9f8e456397ff911bdc55d91b0dbeedfb59711a87c4223deb5217c93b9ca60fe2a0c0881efe285061089acb9980332046f64696e"),
		},
		{
			R:                hexToBytes("75bec79ec5e0d111b94a947886c86afaa1cb85965d9f2c65fa76429b648e8471"),
			S:                hexToBytes("08d6425b1b5bd639bb641a666eea5938213fb39d77cf2470b675979a9d83137e"),
			V:                28,
			SignedDataPrefix: hexToBytes("69080211010000000000000022480a20"),
			SignedDataSuffix: hexToBytes("122408011220b9f8e456397ff911bdc55d91b0dbeedfb59711a87c4223deb5217c93b9ca60fe2a0c0881efe285061088dce68d0332046f64696e"),
		},
		{
			R:                hexToBytes("3ea691693f88baf02552a426f87813a5e0db814cadcf82dac61f5924e78ced9b"),
			S:                hexToBytes("6e3abcdcd983e89a5f79283fabc7b4c33356027380cce138e78f1202beb056d2"),
			V:                27,
			SignedDataPrefix: hexToBytes("69080211010000000000000022480a20"),
			SignedDataSuffix: hexToBytes("122408011220b9f8e456397ff911bdc55d91b0dbeedfb59711a87c4223deb5217c93b9ca60fe2a0c0881efe2850610a993e78d0332046f64696e"),
		},
	}
	require.Equal(t, expected, sig)
}

func TestVerifySignature(t *testing.T) {
	signatures := []TMSignature{
		{
			R:                hexToBytes("6916405D52FF02EC26DD78E831E0A179C89B99CBBDB15C9DA802B75A7621D5EB"),
			S:                hexToBytes("69CF40BE7AC1AA176B13BA4D57EB2B8735A5832014F0DC168EA6F580C51BB222"),
			V:                28,
			SignedDataPrefix: hexToBytes("7808021184C002000000000022480A20"),
			SignedDataSuffix: hexToBytes("12240801122044551F853D916A7C630C0C210C921BAC7D05CE0C249DFC6088C0274F058418272A0C08DE9493850610F0FFAEEB02321362616E642D6C616F7A692D746573746E657431"),
		},
		{
			R:                hexToBytes("6A8E3C35DEED991D257BCA9451360BFBE7978D388AF8D2F864A6919FE1083C7E"),
			S:                hexToBytes("14D145DD6BC1A770ACBDF37DAC08DD8076AB888FDA2739BE9B9767B23A387D1E"),
			V:                27,
			SignedDataPrefix: hexToBytes("7808021184C002000000000022480A20"),
			SignedDataSuffix: hexToBytes("12240801122044551F853D916A7C630C0C210C921BAC7D05CE0C249DFC6088C0274F058418272A0C08DE9493850610DAEB8D9C03321362616E642D6C616F7A692D746573746E657431"),
		},
		{
			R:                hexToBytes("EB402F4B863A1DF91E7772D9574640EFFC5447ECEC6EDF6F1CFE2C33D7DC8DD4"),
			S:                hexToBytes("1FEC45523E885DD6E8AD75EA2D81D30657267DF646406240F206A98749EBD0A7"),
			V:                27,
			SignedDataPrefix: hexToBytes("7808021184C002000000000022480A20"),
			SignedDataSuffix: hexToBytes("12240801122044551F853D916A7C630C0C210C921BAC7D05CE0C249DFC6088C0274F058418272A0C08DE9493850610B68FD4E702321362616E642D6C616F7A692D746573746E657431"),
		},
	}

	evmAddresses := []common.Address{
		common.HexToAddress("0x3b759C4d728e50D5cC04c75f596367829d5b5061"),
		common.HexToAddress("0x49897b9D617AD700b84a935616E81f9f4b5305bc"),
		common.HexToAddress("0x7054bd1Fd7535A0DD552361e634196b1574594BB"),
	}

	blockHash := hexToBytes("8C36C3D12A378BD7E4E8F26BDECCA68B48390240DA456EE9C3292B6E36756AC4")

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
