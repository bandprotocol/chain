package proof

import (
	"encoding/base64"
	"encoding/binary"
	"reflect"
	"time"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	gogotypes "github.com/cosmos/gogoproto/types"
)

// cdcEncode implementation is copied from https://github.com/cometbft/cometbft/blob/v0.37.2/types/encoding_helper.go
func cdcEncode(item interface{}) []byte {
	if item != nil && !isTypedNil(item) && !isEmpty(item) {
		switch item := item.(type) {
		case string:
			i := gogotypes.StringValue{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		case int64:
			i := gogotypes.Int64Value{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		case tmbytes.HexBytes:
			i := gogotypes.BytesValue{
				Value: item,
			}
			bz, err := i.Marshal()
			if err != nil {
				return nil
			}
			return bz
		default:
			panic(item)
		}
	}

	return nil
}

// isTypedNil checks if the given interface is a typed nil.
// Go lacks a simple and safe way to see if something is a typed nil.
// See:
//   - https://dave.cheney.net/2017/08/09/typed-nils-in-go-2
//   - https://groups.google.com/forum/#!topic/golang-nuts/wnH302gBa4I/discussion
//   - https://github.com/golang/go/issues/21538
func isTypedNil(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// isEmpty returns true if it has zero length.
func isEmpty(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	default:
		return false
	}
}

func base64ToBytes(s string) []byte {
	decodedString, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decodedString
}

func bytesToBase64(b []byte) string {
	encodedString := base64.StdEncoding.EncodeToString(b)
	return encodedString
}

// func encodeStoreMerkleHash(key string, value []byte) []byte {
// 	bytesKey := []byte(key)
// 	keyBytes := append([]byte{uint8(len(bytesKey))}, bytesKey...)
// 	valueBytes := append([]byte{32}, tmhash.Sum(tmhash.Sum(value))...)
// 	return append(keyBytes, valueBytes...)
// }

// func encodeVarint(value int64) []byte {
// 	buf := make([]byte, binary.MaxVarintLen64)
// 	n := binary.PutVarint(buf, value)
// 	return buf[:n]
// }

func encodeUvarint(value uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	return buf[:n]
}

// func mustParseInt64(b []byte) int64 {
// 	i64, err := strconv.ParseInt(string(b), 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return i64
// }

// func mustParseUint64(b []byte) uint64 {
// 	u64, err := strconv.ParseUint(string(b), 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return u64
// }

// func mustDecodeString(hexstr string) []byte {
// 	b, err := hex.DecodeString(hexstr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b
// }

func parseTime(str string) time.Time {
	var layout string
	if len(str) == 29 {
		layout = "2006-01-02T15:04:05.00000000Z"
	} else {
		layout = "2006-01-02T15:04:05.000000000Z"
	}
	t, err := time.Parse(layout, str)
	if err != nil {
		panic(err)
	}
	return t
}

func encodeTime(t time.Time) []byte {
	bz := []byte{}
	s := t.Unix()
	// skip if default/zero value:
	if s != 0 {
		bz = append(bz, encodeFieldNumberAndType3(1, 0)...)
		bz = append(bz, encodeUvarint(uint64(s))...)
	}
	ns := int32(t.Nanosecond()) // this int64 -> int32 cast is safe (nanos are in [0, 999999999])
	// skip if default/zero value:
	if ns != 0 {
		bz = append(bz, encodeFieldNumberAndType3(2, 0)...)
		bz = append(bz, encodeUvarint(uint64(ns))...)
	}
	return bz
}

// encodeFieldNumberAndType3 Write field key.
func encodeFieldNumberAndType3(num uint32, typ uint8) []byte {
	return encodeUvarint((uint64(num) << 3) | uint64(typ))
}

func convertVarIntToBytes(orig int64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], orig)
	return buf[:n]
}
