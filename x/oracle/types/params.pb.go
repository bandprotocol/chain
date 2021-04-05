// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: oracle/v1/params.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params is the data structure that keeps the parameters of the oracle module.
type Params struct {
	// MaxRawRequestCount is the maximum number of data source raw requests a
	// request can make.
	MaxRawRequestCount uint64 `protobuf:"varint,1,opt,name=max_raw_request_count,json=maxRawRequestCount,proto3" json:"max_raw_request_count,omitempty"`
	// MaxAskCount is the maximum number of validators a request can target.
	MaxAskCount uint64 `protobuf:"varint,2,opt,name=max_ask_count,json=maxAskCount,proto3" json:"max_ask_count,omitempty"`
	// ExpirationBlockCount is the number of blocks a request stays valid before
	// it gets expired due to insufficient reports.
	ExpirationBlockCount uint64 `protobuf:"varint,3,opt,name=expiration_block_count,json=expirationBlockCount,proto3" json:"expiration_block_count,omitempty"`
	// BaseOwasmGas is the base amount of Cosmos-SDK gas charged for owasm
	// execution.
	BaseOwasmGas uint64 `protobuf:"varint,4,opt,name=base_owasm_gas,json=baseOwasmGas,proto3" json:"base_owasm_gas,omitempty"`
	// PerValidatorRequestGas is the amount of Cosmos-SDK gas charged per
	// requested validator.
	PerValidatorRequestGas uint64 `protobuf:"varint,5,opt,name=per_validator_request_gas,json=perValidatorRequestGas,proto3" json:"per_validator_request_gas,omitempty"`
	// SamplingTryCount the number of validator sampling tries to pick the highest
	// voting power subset of validators to perform an oracle task.
	SamplingTryCount uint64 `protobuf:"varint,6,opt,name=sampling_try_count,json=samplingTryCount,proto3" json:"sampling_try_count,omitempty"`
	// OracleRewardPercentage is the percentage of block rewards allocated to
	// active oracle validators.
	OracleRewardPercentage uint64 `protobuf:"varint,7,opt,name=oracle_reward_percentage,json=oracleRewardPercentage,proto3" json:"oracle_reward_percentage,omitempty"`
	// InactivePenaltyDuration is the duration period where a validator cannot
	// activate back after missing an oracle report.
	InactivePenaltyDuration uint64 `protobuf:"varint,8,opt,name=inactive_penalty_duration,json=inactivePenaltyDuration,proto3" json:"inactive_penalty_duration,omitempty"`
	// MaxDataSize is the maximum number of bytes that can be present in the report as the result
	MaxDataSize uint64 `protobuf:"varint,9,opt,name=max_data_size,json=maxDataSize,proto3" json:"max_data_size,omitempty"`
	// MaxCalldataSize is the maximum number of bytes that can be present in the calldata
	MaxCalldataSize uint64 `protobuf:"varint,10,opt,name=max_calldata_size,json=maxCalldataSize,proto3" json:"max_calldata_size,omitempty"`
	// TODO: maybe use DecCoins
	// DataProviderRewardPerByte is the amount of tokens, user gets for the byte of data provided
	DataProviderRewardPerByte github_com_cosmos_cosmos_sdk_types.DecCoin `protobuf:"bytes,11,opt,name=data_provider_reward_per_byte,json=dataProviderRewardPerByte,proto3,customtype=github.com/cosmos/cosmos-sdk/types.DecCoin" json:"data_provider_reward_per_byte"`
	// TODO: maybe use Coins
	// DataRequesterBasicFee is the amount of tokens user has to pay in DataProvidersPool for the data provided
	DataRequesterBasicFee github_com_cosmos_cosmos_sdk_types.Coin `protobuf:"bytes,12,opt,name=data_requester_basic_fee,json=dataRequesterBasicFee,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Coin" json:"data_requester_basic_fee"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_d7000dc69c8e604b, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetMaxRawRequestCount() uint64 {
	if m != nil {
		return m.MaxRawRequestCount
	}
	return 0
}

func (m *Params) GetMaxAskCount() uint64 {
	if m != nil {
		return m.MaxAskCount
	}
	return 0
}

func (m *Params) GetExpirationBlockCount() uint64 {
	if m != nil {
		return m.ExpirationBlockCount
	}
	return 0
}

func (m *Params) GetBaseOwasmGas() uint64 {
	if m != nil {
		return m.BaseOwasmGas
	}
	return 0
}

func (m *Params) GetPerValidatorRequestGas() uint64 {
	if m != nil {
		return m.PerValidatorRequestGas
	}
	return 0
}

func (m *Params) GetSamplingTryCount() uint64 {
	if m != nil {
		return m.SamplingTryCount
	}
	return 0
}

func (m *Params) GetOracleRewardPercentage() uint64 {
	if m != nil {
		return m.OracleRewardPercentage
	}
	return 0
}

func (m *Params) GetInactivePenaltyDuration() uint64 {
	if m != nil {
		return m.InactivePenaltyDuration
	}
	return 0
}

func (m *Params) GetMaxDataSize() uint64 {
	if m != nil {
		return m.MaxDataSize
	}
	return 0
}

func (m *Params) GetMaxCalldataSize() uint64 {
	if m != nil {
		return m.MaxCalldataSize
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "oracle.v1.Params")
}

func init() { proto.RegisterFile("oracle/v1/params.proto", fileDescriptor_d7000dc69c8e604b) }

var fileDescriptor_d7000dc69c8e604b = []byte{
	// 587 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0xbd, 0x6e, 0x13, 0x4f,
	0x14, 0xc5, 0xbd, 0xff, 0x7f, 0x30, 0xc9, 0x24, 0x7c, 0xad, 0x92, 0xb0, 0x8e, 0x60, 0x1d, 0x45,
	0x48, 0x44, 0x11, 0xd9, 0x95, 0x03, 0x05, 0xa4, 0xc3, 0xb1, 0x48, 0x01, 0x12, 0x96, 0x41, 0x14,
	0x34, 0xa3, 0xeb, 0xd9, 0x8b, 0x19, 0x65, 0x77, 0x67, 0x99, 0x19, 0x7f, 0xa5, 0xe5, 0x01, 0x40,
	0x54, 0x94, 0x79, 0x9c, 0x94, 0x29, 0x11, 0x45, 0x84, 0xe2, 0x86, 0xc7, 0x40, 0x33, 0xb3, 0xeb,
	0xa4, 0xa0, 0xa0, 0xb2, 0x75, 0xcf, 0x39, 0x3e, 0xbf, 0xb9, 0x9e, 0x21, 0xeb, 0x42, 0x02, 0x4b,
	0x31, 0x1e, 0xb5, 0xe2, 0x02, 0x24, 0x64, 0x2a, 0x2a, 0xa4, 0xd0, 0xc2, 0x5f, 0x72, 0xf3, 0x68,
	0xd4, 0xda, 0x58, 0x1d, 0x88, 0x81, 0xb0, 0xd3, 0xd8, 0x7c, 0x73, 0x86, 0x8d, 0x90, 0x09, 0x95,
	0x09, 0x15, 0xf7, 0x41, 0x99, 0x74, 0x1f, 0x35, 0xb4, 0x62, 0x26, 0x78, 0xee, 0xf4, 0xad, 0x2f,
	0x75, 0x52, 0xef, 0xda, 0x5f, 0xf4, 0x5b, 0x64, 0x2d, 0x83, 0x09, 0x95, 0x30, 0xa6, 0x12, 0x3f,
	0x0d, 0x51, 0x69, 0xca, 0xc4, 0x30, 0xd7, 0x81, 0xb7, 0xe9, 0x6d, 0x2f, 0xf4, 0xfc, 0x0c, 0x26,
	0x3d, 0x18, 0xf7, 0x9c, 0x74, 0x60, 0x14, 0x7f, 0x8b, 0xdc, 0x30, 0x11, 0x50, 0x47, 0xa5, 0xf5,
	0x3f, 0x6b, 0x5d, 0xce, 0x60, 0xf2, 0x5c, 0x1d, 0x39, 0xcf, 0x13, 0xb2, 0x8e, 0x93, 0x82, 0x4b,
	0xd0, 0x5c, 0xe4, 0xb4, 0x9f, 0x0a, 0x56, 0x99, 0xff, 0xb7, 0xe6, 0xd5, 0x4b, 0xb5, 0x6d, 0x44,
	0x97, 0x7a, 0x40, 0x6e, 0x1a, 0x64, 0x2a, 0xc6, 0xa0, 0x32, 0x3a, 0x00, 0x15, 0x2c, 0x58, 0xf7,
	0x8a, 0x99, 0xbe, 0x36, 0xc3, 0x43, 0x50, 0xfe, 0x33, 0xd2, 0x28, 0x50, 0xd2, 0x11, 0xa4, 0x3c,
	0x01, 0x2d, 0xe4, 0x1c, 0xdc, 0x04, 0xae, 0xd9, 0xc0, 0x7a, 0x81, 0xf2, 0x5d, 0xa5, 0x97, 0xf0,
	0x26, 0xfa, 0x88, 0xf8, 0x0a, 0xb2, 0x22, 0xe5, 0xf9, 0x80, 0x6a, 0x39, 0x2d, 0x91, 0xea, 0x36,
	0x73, 0xbb, 0x52, 0xde, 0xca, 0xa9, 0xc3, 0x79, 0x4a, 0x02, 0xb7, 0x69, 0x2a, 0x71, 0x0c, 0x32,
	0xa1, 0x05, 0x4a, 0x86, 0xb9, 0x86, 0x01, 0x06, 0xd7, 0x5d, 0x8f, 0xd3, 0x7b, 0x56, 0xee, 0xce,
	0x55, 0x7f, 0x9f, 0x34, 0x78, 0x0e, 0x4c, 0xf3, 0x11, 0xd2, 0x02, 0x73, 0x48, 0xf5, 0x94, 0x26,
	0x43, 0x77, 0xde, 0x60, 0xd1, 0x46, 0xef, 0x56, 0x86, 0xae, 0xd3, 0x3b, 0xa5, 0x5c, 0xad, 0x37,
	0x01, 0x0d, 0x54, 0xf1, 0x63, 0x0c, 0x96, 0xe6, 0xeb, 0xed, 0x80, 0x86, 0x37, 0xfc, 0x18, 0xfd,
	0x1d, 0x72, 0xc7, 0x78, 0x18, 0xa4, 0xe9, 0xa5, 0x8f, 0x58, 0xdf, 0xad, 0x0c, 0x26, 0x07, 0xe5,
	0xdc, 0x7a, 0xbf, 0x79, 0xe4, 0xbe, 0x35, 0x15, 0x52, 0x8c, 0x78, 0x82, 0xf2, 0xca, 0x69, 0x68,
	0x7f, 0xaa, 0x31, 0x58, 0xde, 0xf4, 0xb6, 0x97, 0xf7, 0xee, 0x45, 0xee, 0xd6, 0x44, 0x66, 0xd9,
	0x51, 0x79, 0x6b, 0xa2, 0x0e, 0xb2, 0x03, 0xc1, 0xf3, 0xf6, 0xde, 0xe9, 0x79, 0xb3, 0xf6, 0xf3,
	0xbc, 0xb9, 0x33, 0xe0, 0xfa, 0xe3, 0xb0, 0x1f, 0x31, 0x91, 0xc5, 0xe5, 0x2d, 0x73, 0x1f, 0xbb,
	0x2a, 0x39, 0x8a, 0xf5, 0xb4, 0x40, 0x55, 0x65, 0x7a, 0x0d, 0x53, 0xdb, 0x2d, 0x5b, 0xe7, 0x3b,
	0x6a, 0x4f, 0x35, 0xfa, 0x9f, 0x3d, 0x12, 0x58, 0xa8, 0xf2, 0xbf, 0x33, 0x28, 0xa0, 0x38, 0xa3,
	0x1f, 0x10, 0x83, 0x15, 0xcb, 0xd3, 0xf8, 0x2b, 0x8f, 0x85, 0x89, 0x4b, 0x98, 0x87, 0xff, 0x00,
	0x63, 0x49, 0xd6, 0x4c, 0x57, 0xaf, 0xaa, 0x6a, 0x9b, 0xa6, 0x17, 0x88, 0xfb, 0x8b, 0xdf, 0x4f,
	0x9a, 0xb5, 0xdf, 0x27, 0x4d, 0xaf, 0xfd, 0xf2, 0xf4, 0x22, 0xf4, 0xce, 0x2e, 0x42, 0xef, 0xd7,
	0x45, 0xe8, 0x7d, 0x9d, 0x85, 0xb5, 0xb3, 0x59, 0x58, 0xfb, 0x31, 0x0b, 0x6b, 0xef, 0x5b, 0x57,
	0x3a, 0x0e, 0x51, 0x74, 0xda, 0xbb, 0xaf, 0x78, 0xc6, 0x35, 0x26, 0xb1, 0x48, 0x78, 0xbe, 0xcb,
	0x84, 0xc4, 0x78, 0x12, 0x97, 0x2f, 0xd5, 0x56, 0xf6, 0xeb, 0xf6, 0x95, 0x3d, 0xfe, 0x13, 0x00,
	0x00, 0xff, 0xff, 0xb7, 0xf5, 0x90, 0x72, 0xc0, 0x03, 0x00, 0x00,
}

func (this *Params) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxRawRequestCount != that1.MaxRawRequestCount {
		return false
	}
	if this.MaxAskCount != that1.MaxAskCount {
		return false
	}
	if this.ExpirationBlockCount != that1.ExpirationBlockCount {
		return false
	}
	if this.BaseOwasmGas != that1.BaseOwasmGas {
		return false
	}
	if this.PerValidatorRequestGas != that1.PerValidatorRequestGas {
		return false
	}
	if this.SamplingTryCount != that1.SamplingTryCount {
		return false
	}
	if this.OracleRewardPercentage != that1.OracleRewardPercentage {
		return false
	}
	if this.InactivePenaltyDuration != that1.InactivePenaltyDuration {
		return false
	}
	if this.MaxDataSize != that1.MaxDataSize {
		return false
	}
	if this.MaxCalldataSize != that1.MaxCalldataSize {
		return false
	}
	if !this.DataProviderRewardPerByte.Equal(that1.DataProviderRewardPerByte) {
		return false
	}
	if !this.DataRequesterBasicFee.Equal(that1.DataRequesterBasicFee) {
		return false
	}
	return true
}
func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.DataRequesterBasicFee.Size()
		i -= size
		if _, err := m.DataRequesterBasicFee.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x62
	{
		size := m.DataProviderRewardPerByte.Size()
		i -= size
		if _, err := m.DataProviderRewardPerByte.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x5a
	if m.MaxCalldataSize != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxCalldataSize))
		i--
		dAtA[i] = 0x50
	}
	if m.MaxDataSize != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxDataSize))
		i--
		dAtA[i] = 0x48
	}
	if m.InactivePenaltyDuration != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.InactivePenaltyDuration))
		i--
		dAtA[i] = 0x40
	}
	if m.OracleRewardPercentage != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.OracleRewardPercentage))
		i--
		dAtA[i] = 0x38
	}
	if m.SamplingTryCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.SamplingTryCount))
		i--
		dAtA[i] = 0x30
	}
	if m.PerValidatorRequestGas != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.PerValidatorRequestGas))
		i--
		dAtA[i] = 0x28
	}
	if m.BaseOwasmGas != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.BaseOwasmGas))
		i--
		dAtA[i] = 0x20
	}
	if m.ExpirationBlockCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.ExpirationBlockCount))
		i--
		dAtA[i] = 0x18
	}
	if m.MaxAskCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxAskCount))
		i--
		dAtA[i] = 0x10
	}
	if m.MaxRawRequestCount != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxRawRequestCount))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MaxRawRequestCount != 0 {
		n += 1 + sovParams(uint64(m.MaxRawRequestCount))
	}
	if m.MaxAskCount != 0 {
		n += 1 + sovParams(uint64(m.MaxAskCount))
	}
	if m.ExpirationBlockCount != 0 {
		n += 1 + sovParams(uint64(m.ExpirationBlockCount))
	}
	if m.BaseOwasmGas != 0 {
		n += 1 + sovParams(uint64(m.BaseOwasmGas))
	}
	if m.PerValidatorRequestGas != 0 {
		n += 1 + sovParams(uint64(m.PerValidatorRequestGas))
	}
	if m.SamplingTryCount != 0 {
		n += 1 + sovParams(uint64(m.SamplingTryCount))
	}
	if m.OracleRewardPercentage != 0 {
		n += 1 + sovParams(uint64(m.OracleRewardPercentage))
	}
	if m.InactivePenaltyDuration != 0 {
		n += 1 + sovParams(uint64(m.InactivePenaltyDuration))
	}
	if m.MaxDataSize != 0 {
		n += 1 + sovParams(uint64(m.MaxDataSize))
	}
	if m.MaxCalldataSize != 0 {
		n += 1 + sovParams(uint64(m.MaxCalldataSize))
	}
	l = m.DataProviderRewardPerByte.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.DataRequesterBasicFee.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRawRequestCount", wireType)
			}
			m.MaxRawRequestCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxRawRequestCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxAskCount", wireType)
			}
			m.MaxAskCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxAskCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpirationBlockCount", wireType)
			}
			m.ExpirationBlockCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpirationBlockCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseOwasmGas", wireType)
			}
			m.BaseOwasmGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BaseOwasmGas |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerValidatorRequestGas", wireType)
			}
			m.PerValidatorRequestGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerValidatorRequestGas |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SamplingTryCount", wireType)
			}
			m.SamplingTryCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SamplingTryCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleRewardPercentage", wireType)
			}
			m.OracleRewardPercentage = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OracleRewardPercentage |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InactivePenaltyDuration", wireType)
			}
			m.InactivePenaltyDuration = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InactivePenaltyDuration |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxDataSize", wireType)
			}
			m.MaxDataSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxDataSize |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxCalldataSize", wireType)
			}
			m.MaxCalldataSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxCalldataSize |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataProviderRewardPerByte", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DataProviderRewardPerByte.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataRequesterBasicFee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DataRequesterBasicFee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
