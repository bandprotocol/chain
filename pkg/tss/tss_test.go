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
	suite.lo = hexDecode("0db10b2760c2de165394dae4a5747a5ed1a4f48b81367a7db473d31b1437a38a")
	suite.data = []byte("data")
	suite.fakeData = []byte("fakeData")
	suite.nonce = hexDecode("0000000000000000000000000000000000000000000000000000006e6f6e6365")

	suite.groupID = tss.GroupID(1)
	suite.groupDKGContext = hexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")
	suite.groupThreshold = uint64(2)
	suite.groupPubKey = hexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044")
	suite.groupPubNonce = hexDecode("03adbd937b6c75db1950368510529d98cc75b250bc3abf0b1e1259d8261663eb1e")

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
			PrivateKey: hexDecode("16345ba335ee62fb72857c1624ccfa034b42739faa1a3da7e670aa95e978c65d"),
			PublicKey:  hexDecode("02e769132e689823ed8b791ee96a18ff1e7b381f49310eea083c1ac2f9a7edf676"),
		},
		e: tss.KeyPair{
			PrivateKey: hexDecode("8d4efee63ee784d480b58d801c68a6400be2c8717c09ebde3c948537b35952f9"),
			PublicKey:  hexDecode("021e4c38834733962a416cd5cf9378f4d20761b873b94b2368aea2284224043c64"),
		},
		ownKey: tss.KeyPair{
			PrivateKey: hexDecode("b248a8a2f6f1644b196402de4026d3b63db36529b2b365995f5b21eebf20acea"),
			PublicKey:  hexDecode("0268c34a74f75ea26f3eba73a44afdaaa5e4704baa6f58d6e1ab831a5608e4dae4"),
		},
		ownNonce: tss.KeyPair{
			PrivateKey: hexDecode("a6eaa61cb0e2d31363127904987872f42dd3dce0344ef5cebc3b99a0f0a6eb59"),
			PublicKey:  hexDecode("031d8cf9c2386ff16c8555d108264d0c5a30ca7600c8aa9b2c837610560c1ee0a1"),
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
			PrivateKey: hexDecode("6bbb750dd8ccc743fe806c938e39101d0c6354e79db8169e07b2740ee7649123"),
			PublicKey:  hexDecode("03486905485e94b763153279a1ccfa958cc381b2b20320c6ccb33dbe3cf3c4c1f6"),
		},
		e: tss.KeyPair{
			PrivateKey: hexDecode("7fc182acddac8be21145b7e1cdc39002aa83c365a146c81efe32bca8e0c8e98c"),
			PublicKey:  hexDecode("0363e98cc4b8a587ad618990520d4e5eaa6c497cb8be9207551815083b25184778"),
		},
		ownKey: tss.KeyPair{
			PrivateKey: hexDecode("1b4379a07902f9b18f9b8eefc1f340e8b42ed34fe4f6d125416e3e6cffc77eb6"),
			PublicKey:  hexDecode("034c0386dff08b142f356c0c7ae610c9cba27239a5447cde69c7c953b7b65f89c7"),
		},
		ownNonce: tss.KeyPair{
			PrivateKey: hexDecode("3efbc92a479fe6f2d4da9df61571727548805c51687ff5acd70614f69d908dcd"),
			PublicKey:  hexDecode("03047362ee6171f24f6ba63c73ab660803557c928cde5b96cb21b1fa4a11072674"),
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
