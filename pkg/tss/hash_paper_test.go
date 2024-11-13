package tss_test

import (
	"crypto/sha256"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

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
		want string
	}{
		{
			// zero length string
			x:    []byte{},
			want: "0",
		},
		{
			// OS2IP(I2OSP(0, 2))
			x:    []byte{0x00, 0x00},
			want: "0",
		},
		{
			// OS2IP(I2OSP(1, 2))
			x:    []byte{0x00, 0x01},
			want: "1",
		},
		{
			// OS2IP(I2OSP(255, 2))
			x:    []byte{0x00, 0xff},
			want: "255",
		},
		{
			// OS2IP(I2OSP(256, 2))
			x:    []byte{0x01, 0x00},
			want: "256",
		},
		{
			// OS2IP(I2OSP(65535, 2))
			x:    []byte{0xff, 0xff},
			want: "65535",
		},
		{
			// OS2IP(I2OSP(1234, 5))
			x:    []byte{0x00, 0x00, 0x00, 0x04, 0xd2},
			want: "1234",
		},
	}
	for _, tt := range tests {
		suite.Require().Equal(tt.want, tss.OS2IP(tt.x).String())
	}
}

// Using test vectors from https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-hash-to-curve-16#name-expand_message_xmdsha-256
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
			H: hashSha256,
			msg: []byte(
				"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     32, // 0x20
			wantUniformBytes: "b23a1d2b4d97b2ef7785562a7e8bac7eed54ed6e97e29aa51bfe3f12ddad1ff9",
		},
		{
			H: hashSha256,
			msg: []byte(
				"a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			),
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
			H: hashSha256,
			msg: []byte(
				"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "80be107d0884f0d881bb460322f0443d38bd222db8bd0b0a5312a6fedb49c1bbd88fd75d8b9a09486c60123dfa1d73c1cc3169761b17476d3c6b7cbbd727acd0e2c942f4dd96ae3da5de368d26b32286e32de7e5a8cb2949f866a0b80c58116b29fa7fabb3ea7d520ee603e0c25bcaf0b9a5e92ec6a1fe4e0391d1cdbce8c68a",
		},
		{
			H: hashSha256,
			msg: []byte(
				"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			),
			DST:              []byte("QUUX-V01-CS02-with-expander-SHA256-128"),
			len_in_bytes:     128, // 0x80
			wantUniformBytes: "80be107d0884f0d881bb460322f0443d38bd222db8bd0b0a5312a6fedb49c1bbd88fd75d8b9a09486c60123dfa1d73c1cc3169761b17476d3c6b7cbbd727acd0e2c942f4dd96ae3da5de368d26b32286e32de7e5a8cb2949f866a0b80c58116b29fa7fabb3ea7d520ee603e0c25bcaf0b9a5e92ec6a1fe4e0391d1cdbce8c68a",
		},
		{
			H: hashSha256,
			msg: []byte(
				"a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			),
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

// Using test vectors from https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-hash-to-curve-16#name-secp256k1
func (suite *TSSTestSuite) TestH_M1_L48() {
	p := crypto.S256().Params().P
	hashSha256 := func(data ...[]byte) []byte {
		var combined []byte
		for _, d := range data {
			combined = append(combined, d...)
		}
		hash := sha256.Sum256(combined)
		return hash[:]
	}
	tests := []struct {
		msg    []byte
		count  int
		DST    string
		wantU0 string
		wantU1 string
	}{
		{
			msg:    []byte{},
			count:  2,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_RO_",
			wantU0: "6b0f9910dd2ba71c78f2ee9f04d73b5f4c5f7fc773a701abea1e573cab002fb3",
			wantU1: "1ae6c212e08fe1a5937f6202f929a2cc8ef4ee5b9782db68b0d5799fd8f09e16",
		},
		{
			msg:    []byte("abc"),
			count:  2,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_RO_",
			wantU0: "128aab5d3679a1f7601e3bdf94ced1f43e491f544767e18a4873f397b08a2b61",
			wantU1: "5897b65da3b595a813d0fdcc75c895dc531be76a03518b044daaa0f2e4689e00",
		},
		{
			msg:    []byte("abcdef0123456789"),
			count:  2,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_RO_",
			wantU0: "ea67a7c02f2cd5d8b87715c169d055a22520f74daeb080e6180958380e2f98b9",
			wantU1: "7434d0d1a500d38380d1f9615c021857ac8d546925f5f2355319d823a478da18",
		},
		{
			msg: []byte(
				"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			),
			count:  2,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_RO_",
			wantU0: "eda89a5024fac0a8207a87e8cc4e85aa3bce10745d501a30deb87341b05bcdf5",
			wantU1: "dfe78cd116818fc2c16f3837fedbe2639fab012c407eac9dfe9245bf650ac51d",
		},
		{
			msg: []byte(
				"a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			),
			count:  2,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_RO_",
			wantU0: "8d862e7e7e23d7843fe16d811d46d7e6480127a6b78838c277bca17df6900e9f",
			wantU1: "68071d2530f040f081ba818d3c7188a94c900586761e9115efa47ae9bd847938",
		},
		{
			msg:    []byte{},
			count:  1,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_NU_",
			wantU0: "0137fcd23bc3da962e8808f97474d097a6c8aa2881fceef4514173635872cf3b",
		},
		{
			msg:    []byte("abc"),
			count:  1,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_NU_",
			wantU0: "e03f894b4d7caf1a50d6aa45cac27412c8867a25489e32c5ddeb503229f63a2e",
		},
		{
			msg:    []byte("abcdef0123456789"),
			count:  1,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_NU_",
			wantU0: "e7a6525ae7069ff43498f7f508b41c57f80563c1fe4283510b322446f32af41b",
		},
		{
			msg: []byte(
				"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			),
			count:  1,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_NU_",
			wantU0: "d97cf3d176a2f26b9614a704d7d434739d194226a706c886c5c3c39806bc323c",
		},
		{
			msg: []byte(
				"a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			),
			count:  1,
			DST:    "QUUX-V01-CS02-with-secp256k1_XMD:SHA-256_SSWU_NU_",
			wantU0: "a9ffbeee1d6e41ac33c248fb3364612ff591b502386c1bf6ac4aaf1ea51f8c3b",
		},
	}
	for _, tt := range tests {
		result, err := tss.H_M1_L48(hashSha256, tt.count, p, tt.msg, tt.DST)
		suite.Require().NoError(err)
		if tt.count == 1 {
			u0 := result[0][0].Text(16)
			if len(u0)%2 != 0 {
				u0 = "0" + u0
			}
			suite.Require().Equal(tt.wantU0, u0)
		} else if tt.count == 2 {
			suite.Require().Equal(tt.wantU0, result[0][0].Text(16))
			suite.Require().Equal(tt.wantU1, result[1][0].Text(16))
		}
	}
}
