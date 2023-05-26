package tss_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/suite"
)

type member struct {
	mid tss.MemberID
	tss.Round1Data
	nonceSym        tss.PublicKey
	keySyms         tss.PublicKeys
	secretShares    tss.Scalars
	encSecretShares tss.Scalars
	d               tss.KeyPair
	e               tss.KeyPair
	ownNonce        tss.KeyPair
	ownKey          tss.KeyPair
	lagrange        tss.Scalar
}

type TSSTestSuite struct {
	suite.Suite

	lo       tss.Scalar
	data     []byte
	bytes    []byte
	fakeData []byte
	nonce    []byte

	groupID         tss.GroupID
	groupDKGContext []byte
	groupThreshold  uint64
	groupPubKey     tss.PublicKey
	groupPubNonce   tss.PublicKey

	member1 member
	member2 member
	fakeKey tss.KeyPair
}

func (suite *TSSTestSuite) SetupTest() {
	suite.lo = hexDecode("1c2a36f7f92bf15a859621d3be209a977ed3b8da5c95abf8b804c913f3d9a720")
	suite.data = []byte("data")
	suite.bytes = []byte("bytes")
	suite.fakeData = []byte("fakeData")
	suite.nonce = hexDecode("0000000000000000000000000000000000000000000000000000006e6f6e6365")

	suite.groupID = tss.GroupID(1)
	suite.groupDKGContext = hexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")
	suite.groupThreshold = uint64(2)
	suite.groupPubKey = hexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044")
	suite.groupPubNonce = hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a")

	suite.member1 = member{
		mid: tss.MemberID(1),
		Round1Data: tss.Round1Data{
			OneTimePrivKey: hexDecode("83127264737dd61b4b7f8058a8418874f0e0e52ada48b39a497712a487096304"),
			OneTimePubKey:  hexDecode("0383764b806848430ed195ef8017fb4e768893ea07782e679c31e5ff1b8b453973"),
			OneTimeSig: hexDecode(
				"023d5cdddbdbe503590231e9a8096348cf27d93714021feaef91b3c09553723ba3c5d137db80b4642825e48c425450f14731e7cd3c2397abb4b2c70e65a70b062e",
			),
			A0PrivKey: hexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
			A0PubKey:  hexDecode("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3"),
			A0Sig: hexDecode(
				"033638414d6249831a89965f5f7fc59a77efc9335c4565cbd79f29f86b252d547a8aa2f99b06c196c7a81931b2a099ab1fcf998d115173e9def162b50180ddf2d9",
			),
			Coefficients: tss.Scalars{
				hexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
				hexDecode("2611e629e7043dbd32682e91c73292b087bb9f0747cd4bdd94e92093b6716504"),
			},
			CoefficientsCommit: tss.Points{
				hexDecode("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3"),
				hexDecode("035b8a99ebc56c07b88404407046e6f0d5e5318a87b888ea25d6d12d8175b2d70c"),
			},
		},
		nonceSym: hexDecode("029edf4c76a3dfa484d3182672f40c490e969718a47611c516a875c889ed580767"),
		keySyms: tss.PublicKeys{
			hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
		},
		secretShares:    tss.Scalars{hexDecode("fc93f14f4e3e4e15378e2c65ba1986494a3f54b7c135dd21d67a44435332eb71")},
		encSecretShares: tss.Scalars{hexDecode("d47a459f272be3d22e54af5a0a45ea8318e88f2c3c767962b2b5f9ba53d9922d")},
		d: tss.KeyPair{
			PrivateKey: hexDecode("c51328a8409dab5115f9b081fdaa6f0271ac4482c0bcebf407c5734efdaccb9f"),
			PublicKey:  hexDecode("03cd12d8f9abd0537d125fc6c998567bfd223cbdeb5ba66443f59731ff1a008aa2"),
		},
		e: tss.KeyPair{
			PrivateKey: hexDecode("dc835a4bd0c0e59b1aae9411a10325521b4140e2b02df9516ba8d06071c8d627"),
			PublicKey:  hexDecode("03741dc9ba6f4876636424e02ab325dea615e262cc9b0e14404a1857b762cceba2"),
		},
		ownKey: tss.KeyPair{
			PrivateKey: hexDecode("b248a8a2f6f1644b196402de4026d3b63db36529b2b365995f5b21eebf20acea"),
			PublicKey:  hexDecode("0268c34a74f75ea26f3eba73a44afdaaa5e4704baa6f58d6e1ab831a5608e4dae4"),
		},
		ownNonce: tss.KeyPair{
			PrivateKey: hexDecode("9addd6376764dca38714808557daa8753f0317c6896d5c19d5ee647680c8e0fd"),
			PublicKey:  hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
		},
		lagrange: hexDecode("0000000000000000000000000000000000000000000000000000000000000002"),
	}

	suite.member2 = member{
		mid: tss.MemberID(2),
		Round1Data: tss.Round1Data{
			OneTimePrivKey: hexDecode("e628ea45842af65d017c5c8c198f8c16741093b759f7c2145ec3b4a2c76942ad"),
			OneTimePubKey:  hexDecode("02e20b4d6bd3f10e7c3a9098c5832180b809a826ae49a972d5348758529c5015c5"),
			OneTimeSig: hexDecode(
				"031b25c792eebdefb217d64f632fa39b25b2ff1c3aed6889d82560aaf74daa397cc4ee436588e42749d1a24c831f285abafd52ca56f449f78fc77e8aa455e0143c",
			),
			A0PrivKey: hexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
			A0PubKey:  hexDecode("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb"),
			A0Sig: hexDecode(
				"036f96137c7d88f85a723f5933a697afd039032786f08bab2a223b1c0069f2ad07041319c804f1eb2e7d6b181a8beb791c2173451af5bbeab033e5b43590a5a9c3",
			),
			Coefficients: tss.Scalars{
				hexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
				hexDecode("42e8ead39b0d57a943cf5d7fba99da80a96eac0599bebfea0cfc5a775a6bae09"),
			},
			CoefficientsCommit: tss.Points{
				hexDecode("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb"),
				hexDecode("023d61b24c8785efe8c7459dc706d95b197c0acb31697feb49fec2d3446dc36de4"),
			},
		},
		nonceSym: hexDecode("039d0b05b67af8121b94dcac51d6867b27fc483c55199e30a0d1dca992fcd00651"),
		keySyms: tss.PublicKeys{
			hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
		},
		secretShares:    tss.Scalars{hexDecode("dbc69d7d8fb753f3143e050a4d3fe01c35de8c5fe8937490dd9c5ccbf29567be")},
		encSecretShares: tss.Scalars{hexDecode("b3acf1cd68a4e9b00b0487fe9d6c44560487c6d463d410d1b9d81242f33c0e7a")},
		d: tss.KeyPair{
			PrivateKey: hexDecode("56b6a3783a58558ccc349fb3d8b33efd1184d38e80781b5e5ad6ece6067ae0cc"),
			PublicKey:  hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
		},
		e: tss.KeyPair{
			PrivateKey: hexDecode("a83e277fb7568bfa97226c08f4fbb74c6a4a2adf6fda88b70ee84f69dbba610c"),
			PublicKey:  hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
		},
		ownKey: tss.KeyPair{
			PrivateKey: hexDecode("1b4379a07902f9b18f9b8eefc1f340e8b42ed34fe4f6d125416e3e6cffc77eb6"),
			PublicKey:  hexDecode("034c0386dff08b142f356c0c7ae610c9cba27239a5447cde69c7c953b7b65f89c7"),
		},
		ownNonce: tss.KeyPair{
			PrivateKey: hexDecode("aececfaa8b2bd446b39fc2b13f3073806b57d4a7f6abed87486111be87f48be7"),
			PublicKey:  hexDecode("02aacc8be43d6af147efc41f41754acc7764f31b9d0be33a5acbf9bd46bd3bb4bc"),
		},
		lagrange: hexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140"),
	}

	suite.fakeKey = tss.KeyPair{
		PrivateKey: hexDecode("799c17a7f11d5679d1fe7c29c07202dd6dadc2217aea0a659cab69125c0bd49c"),
		PublicKey:  hexDecode("022894d8b8c15950b49d3856b246c19ad154389d1d5889d4214f8ee40c4a16ab3c"),
	}
}

func TestTSSTestSuite(t *testing.T) {
	suite.Run(t, new(TSSTestSuite))
}

func hexDecode(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}
