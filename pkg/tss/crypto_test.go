package tss_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestEncryptAndDecrypt() {
	// Prepare
	expectedValue := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(expectedValue)
	suite.Require().NoError(err)

	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	// Encrypt and decrypt the value using the key.
	ec, err := tss.Encrypt(ev, key)
	suite.Require().NoError(err)
	value, err := tss.Decrypt(ec, key)
	suite.Require().NoError(err)

	// Ensure the decrypted value matches the original value.
	suite.Require().Equal(expectedValue, hex.EncodeToString(value))
}

func (suite *TSSTestSuite) TestHash() {
	// Hash
	data := []byte("data")
	hash := tss.Hash(data)

	// Ensure the hash matches the expected value.
	suite.Require().Equal("8f54f1c2d0eb5771cd5bf67a6689fcd6eed9444d91a39e5ef32a9b4ae5ca14ff", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestI2OSP() {
	tests := []struct {
		name    string
		x       int
		xLen    int
		want    []byte
		wantErr string
	}{
		{
			// negative int #1
			x:       -1,
			xLen:    2,
			wantErr: "I2OSP: x and xLen must be non-negative integer to be converted",
		},
		{
			// negative int #2
			x:       1,
			xLen:    -2,
			wantErr: "I2OSP: x and xLen must be non-negative integer to be converted",
		},
		{
			// integer too large #1
			x:       1,
			xLen:    0,
			wantErr: "I2OSP: integer too large",
		},
		{
			// integer too large #2
			x:       256,
			xLen:    0,
			wantErr: "I2OSP: integer too large",
		},
		{
			// integer too large #3
			x:       1 << 24,
			xLen:    3,
			wantErr: "I2OSP: integer too large",
		},
		{
			// zero length string
			x:    0,
			xLen: 0,
			want: []byte{},
		},
		{
			// I2OSP(0, 2)
			x:    0,
			xLen: 2,
			want: []byte{0x00, 0x00},
		},
		{
			// I2OSP(1, 2)
			x:    1,
			xLen: 2,
			want: []byte{0x00, 0x01},
		},
		{
			// I2OSP(255, 2)
			x:    255,
			xLen: 2,
			want: []byte{0x00, 0xff},
		},
		{
			// I2OSP(256, 2)
			x:    256,
			xLen: 2,
			want: []byte{0x01, 0x00},
		},
		{
			// I2OSP(65535, 2)
			x:    65535,
			xLen: 2,
			want: []byte{0xff, 0xff},
		},
		{
			// I2OSP(1234, 5)
			x:    1234,
			xLen: 5,
			want: []byte{0x00, 0x00, 0x00, 0x04, 0xd2},
		},
	}
	for _, tt := range tests {
		buf, err := tss.I2OSP(tt.x, tt.xLen)
		if tt.wantErr == "" {
			suite.Require().NoError(err)
			suite.Require().Equal(tt.want, buf)
		} else {
			suite.Require().EqualError(err, tt.wantErr)
		}
	}
}

func (suite *TSSTestSuite) TestOS2IP() {
	tests := []struct {
		x    []byte
		want uint
	}{
		{
			// zero length string
			x:    []byte{},
			want: 0,
		},
		{
			// OS2IP(I2OSP(0, 2))
			x:    []byte{0x00, 0x00},
			want: 0,
		},
		{
			// OS2IP(I2OSP(1, 2))
			x:    []byte{0x00, 0x01},
			want: 1,
		},
		{
			// OS2IP(I2OSP(255, 2))
			x:    []byte{0x00, 0xff},
			want: 255,
		},
		{
			// OS2IP(I2OSP(256, 2))
			x:    []byte{0x01, 0x00},
			want: 256,
		},
		{
			// OS2IP(I2OSP(65535, 2))
			x:    []byte{0xff, 0xff},
			want: 65535,
		},
		{
			// OS2IP(I2OSP(1234, 5))
			x:    []byte{0x00, 0x00, 0x00, 0x04, 0xd2},
			want: 1234,
		},
	}
	for _, tt := range tests {
		suite.Require().Equal(tt.want, tss.OS2IP(tt.x))
	}
}

func (suite *TSSTestSuite) TestExpandMessageXMD() {
	hashSha256 := func(data ...[]byte) []byte {
		var combined []byte
		for _, d := range data {
			combined = append(combined, d...)
		}
		hash := sha256.Sum256(combined)
		return hash[:]
	}

	badHashFunction := func(_ ...[]byte) []byte {
		return []byte{1, 2, 3}
	}

	tests := []struct {
		H                func(data ...[]byte) []byte
		msg              []byte
		DST              []byte
		len_in_bytes     int
		wantUniformBytes string
		wantErr          string
	}{
		{
			H:            badHashFunction,
			msg:          []byte("abc"),
			DST:          []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes: 32, // 0x20
			wantErr:      "ExpandMessageXMD: the initial len of b must be >= b_in_bytes",
		},
		{
			H:            hashSha256,
			msg:          []byte("abc"),
			DST:          make([]byte, 255+1),
			len_in_bytes: 32, // 0x20
			wantErr:      "ExpandMessageXMD: input is not within the permissible limits",
		},
		{
			H:            hashSha256,
			msg:          []byte("abc"),
			DST:          []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes: 65535 + 1,
			wantErr:      "ExpandMessageXMD: input is not within the permissible limits",
		},
		{
			H:            hashSha256,
			msg:          []byte("abc"),
			DST:          []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes: 32*255 + 1,
			wantErr:      "ExpandMessageXMD: input is not within the permissible limits",
		},
		{
			H:                hashSha256,
			msg:              []byte{},
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "68a985b87eb6b46952128911f2a4412bbc302a9d759667f87f7a21d803f07235",
		},
		{
			H:                hashSha256,
			msg:              []byte("abc"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "d8ccab23b5985ccea865c6c97b6e5b8350e794e603b4b97902f53a8a0d605615",
		},
		{
			H:                hashSha256,
			msg:              []byte("abcdef0123456789"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "eff31487c770a893cfb36f912fbfcbff40d5661771ca4b2cb4eafe524333f5c1",
		},
		{
			H:                hashSha256,
			msg:              []byte("q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "b23a1d2b4d97b2ef7785562a7e8bac7eed54ed6e97e29aa51bfe3f12ddad1ff9",
		},
		{
			H:                hashSha256,
			msg:              []byte("a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "4623227bcc01293b8c130bf771da8c298dede7383243dc0993d2d94823958c4c",
		},
		{
			H:                hashSha256,
			msg:              []byte{},
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "af84c27ccfd45d41914fdff5df25293e221afc53d8ad2ac06d5e3e29485dadbee0d121587713a3e0dd4d5e69e93eb7cd4f5df4cd103e188cf60cb02edc3edf18eda8576c412b18ffb658e3dd6ec849469b979d444cf7b26911a08e63cf31f9dcc541708d3491184472c2c29bb749d4286b004ceb5ee6b9a7fa5b646c993f0ced",
		},
		{
			H:                hashSha256,
			msg:              []byte("abc"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "abba86a6129e366fc877aab32fc4ffc70120d8996c88aee2fe4b32d6c7b6437a647e6c3163d40b76a73cf6a5674ef1d890f95b664ee0afa5359a5c4e07985635bbecbac65d747d3d2da7ec2b8221b17b0ca9dc8a1ac1c07ea6a1e60583e2cb00058e77b7b72a298425cd1b941ad4ec65e8afc50303a22c0f99b0509b4c895f40",
		},
		{
			H:                hashSha256,
			msg:              []byte("abcdef0123456789"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "ef904a29bffc4cf9ee82832451c946ac3c8f8058ae97d8d629831a74c6572bd9ebd0df635cd1f208e2038e760c4994984ce73f0d55ea9f22af83ba4734569d4bc95e18350f740c07eef653cbb9f87910d833751825f0ebefa1abe5420bb52be14cf489b37fe1a72f7de2d10be453b2c9d9eb20c7e3f6edc5a60629178d9478df",
		},
		{
			H:                hashSha256,
			msg:              []byte("q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "80be107d0884f0d881bb460322f0443d38bd222db8bd0b0a5312a6fedb49c1bbd88fd75d8b9a09486c60123dfa1d73c1cc3169761b17476d3c6b7cbbd727acd0e2c942f4dd96ae3da5de368d26b32286e32de7e5a8cb2949f866a0b80c58116b29fa7fabb3ea7d520ee603e0c25bcaf0b9a5e92ec6a1fe4e0391d1cdbce8c68a",
		},
		{
			H:                hashSha256,
			msg:              []byte("q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "80be107d0884f0d881bb460322f0443d38bd222db8bd0b0a5312a6fedb49c1bbd88fd75d8b9a09486c60123dfa1d73c1cc3169761b17476d3c6b7cbbd727acd0e2c942f4dd96ae3da5de368d26b32286e32de7e5a8cb2949f866a0b80c58116b29fa7fabb3ea7d520ee603e0c25bcaf0b9a5e92ec6a1fe4e0391d1cdbce8c68a",
		},
		{
			H:                hashSha256,
			msg:              []byte("a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "546aff5444b5b79aa6148bd81728704c32decb73a3ba76e9e75885cad9def1d06d6792f8a7d12794e90efed817d96920d728896a4510864370c207f99bd4a608ea121700ef01ed879745ee3e4ceef777eda6d9e5e38b90c86ea6fb0b36504ba4a45d22e86f6db5dd43d98a294bebb9125d5b794e9d2a81181066eb954966a487",
		},
	}
	for _, tt := range tests {
		uniformBytes, err := tss.ExpandMessageXMD(tt.H, tt.msg, tt.DST, tt.len_in_bytes)
		if tt.wantErr == "" {
			suite.Require().NoError(err)
			suite.Require().Equal(tt.wantUniformBytes, fmt.Sprintf("%x", uniformBytes))
		} else {
			suite.Require().EqualError(err, tt.wantErr)
		}
	}
}
