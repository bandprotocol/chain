package tss_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/suite"
)

type TSSTestSuite struct {
	suite.Suite

	gid        tss.GroupID
	mid        tss.MemberID
	kpI        tss.KeyPair
	kpJ        tss.KeyPair
	fakeKp     tss.KeyPair
	dkgContext []byte
	challenge  []byte
	nonce      []byte
	threshold  uint64

	scalars tss.Scalars
	points  tss.Points
}

func (suite *TSSTestSuite) SetupTest() {
	privKey, _ := hex.DecodeString("7fc4175e7eb9661496cc38526f0eb4abccfd89d15f3371c3729e11c3ba1d6a14")
	pubKey, _ := hex.DecodeString("03936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa")
	suite.kpI = tss.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}

	privKey, _ = hex.DecodeString("fbbbca56f0b3887bfe5efc86f0355a60e2c0e0886885b6ae7d5a7197e4262d1f")
	pubKey, _ = hex.DecodeString("03f70e80bac0b32b2599fa54d83b5471e90fac27bb09528f0337b49d464d64426f")
	suite.kpJ = tss.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}

	privKey, _ = hex.DecodeString("9238fa38e2e2f618e582673232a3d2adb5726a66ece5058bf0bad1707e8518da")
	pubKey, _ = hex.DecodeString("0349bc89d629be7b35648f3b6fe7b70069ddddfecd0b3f3a6c103d59ee69245b03")
	suite.fakeKp = tss.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}

	suite.challenge = []byte("challenge")
	suite.nonce = []byte("nonce")
	suite.dkgContext, _ = hex.DecodeString("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")
	suite.gid = 1
	suite.mid = 1
	suite.threshold = 2

	scalar1, _ := hex.DecodeString("d419936233d055c4560f4f51986281c992e09fa27c72c54a40ea83bdffdae763")
	point1, _ := hex.DecodeString("023487463ba3c7dbf9de9dc5bc73393f99ba0d86270ce2e4218d60e4a01d8cd11c")

	scalar2, _ := hex.DecodeString("d11e4043e29c3939a28e276c09d6ac57b724bc3562b63e1d832dc13bce5098e0")
	point2, _ := hex.DecodeString("03bd0e1b7c880ce80d4340540240972522b44bba2afcf50bfbe30e0352f225eba9")

	suite.scalars = tss.Scalars{scalar1, scalar2}
	suite.points = tss.Points{point1, point2}
}

func TestTSSTestSuite(t *testing.T) {
	suite.Run(t, new(TSSTestSuite))
}
