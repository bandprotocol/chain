// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: band/tunnel/v1beta1/params.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// Params is the data structure that keeps the parameters of the module.
type Params struct {
	// min_deposit is the minimum deposit required to create a tunnel.
	MinDeposit github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=min_deposit,json=minDeposit,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"min_deposit"`
	// min_interval is the minimum interval in seconds.
	MinInterval uint64 `protobuf:"varint,2,opt,name=min_interval,json=minInterval,proto3" json:"min_interval,omitempty"`
	// max_interval is the maximum interval in seconds.
	MaxInterval uint64 `protobuf:"varint,3,opt,name=max_interval,json=maxInterval,proto3" json:"max_interval,omitempty"`
	// min_deviation_bps is the minimum deviation in basis points.
	MinDeviationBPS uint64 `protobuf:"varint,4,opt,name=min_deviation_bps,json=minDeviationBps,proto3" json:"min_deviation_bps,omitempty"`
	// max_deviation_bps is the maximum deviation in basis points.
	MaxDeviationBPS uint64 `protobuf:"varint,5,opt,name=max_deviation_bps,json=maxDeviationBps,proto3" json:"max_deviation_bps,omitempty"`
	// max_signals defines the maximum number of signals allowed per tunnel.
	MaxSignals uint64 `protobuf:"varint,6,opt,name=max_signals,json=maxSignals,proto3" json:"max_signals,omitempty"`
	// base_packet_fee is the base fee for each packet.
	BasePacketFee github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,7,rep,name=base_packet_fee,json=basePacketFee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"base_packet_fee"`
	// axelar_ibc_channel specifies the IBC channel used by the tunnel to communicate with the axelar chain.
	AxelarIBCChannel string `protobuf:"bytes,8,opt,name=axelar_ibc_channel,json=axelarIbcChannel,proto3" json:"axelar_ibc_channel,omitempty"`
	// axelar_gmp_account is the account address on axelar chain that processes and verifies Axelar GMP transactions.
	AxelarGMPAccount string `protobuf:"bytes,9,opt,name=axelar_gmp_account,json=axelarGmpAccount,proto3" json:"axelar_gmp_account,omitempty"`
	// axelar_fee_recipient is the account address on axelar chain that receive fee from tunnel.
	AxelarFeeRecipient string `protobuf:"bytes,10,opt,name=axelar_fee_recipient,json=axelarFeeRecipient,proto3" json:"axelar_fee_recipient,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_842b3bf03f22bf82, []int{0}
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

func (m *Params) GetMinDeposit() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.MinDeposit
	}
	return nil
}

func (m *Params) GetMinInterval() uint64 {
	if m != nil {
		return m.MinInterval
	}
	return 0
}

func (m *Params) GetMaxInterval() uint64 {
	if m != nil {
		return m.MaxInterval
	}
	return 0
}

func (m *Params) GetMinDeviationBPS() uint64 {
	if m != nil {
		return m.MinDeviationBPS
	}
	return 0
}

func (m *Params) GetMaxDeviationBPS() uint64 {
	if m != nil {
		return m.MaxDeviationBPS
	}
	return 0
}

func (m *Params) GetMaxSignals() uint64 {
	if m != nil {
		return m.MaxSignals
	}
	return 0
}

func (m *Params) GetBasePacketFee() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.BasePacketFee
	}
	return nil
}

func (m *Params) GetAxelarIBCChannel() string {
	if m != nil {
		return m.AxelarIBCChannel
	}
	return ""
}

func (m *Params) GetAxelarGMPAccount() string {
	if m != nil {
		return m.AxelarGMPAccount
	}
	return ""
}

func (m *Params) GetAxelarFeeRecipient() string {
	if m != nil {
		return m.AxelarFeeRecipient
	}
	return ""
}

func init() {
	proto.RegisterType((*Params)(nil), "band.tunnel.v1beta1.Params")
}

func init() { proto.RegisterFile("band/tunnel/v1beta1/params.proto", fileDescriptor_842b3bf03f22bf82) }

var fileDescriptor_842b3bf03f22bf82 = []byte{
	// 489 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0xb1, 0x6e, 0xdb, 0x30,
	0x10, 0x86, 0xad, 0xc6, 0x71, 0x1b, 0xa6, 0x85, 0x53, 0xc5, 0x83, 0x9a, 0x41, 0x72, 0x3b, 0x79,
	0xa9, 0x98, 0x34, 0x5b, 0x97, 0x22, 0x72, 0x91, 0xc0, 0x43, 0x00, 0x43, 0xd9, 0xba, 0x08, 0x27,
	0x9a, 0xb1, 0x89, 0x48, 0x24, 0x61, 0xd2, 0x86, 0xfa, 0x16, 0x1d, 0xfa, 0x00, 0x9d, 0xfb, 0x24,
	0x19, 0x33, 0x76, 0x72, 0x0b, 0x7b, 0xe9, 0x63, 0x14, 0x24, 0xe5, 0x54, 0xf0, 0x9c, 0x49, 0xc2,
	0xfd, 0x1f, 0xbf, 0x3b, 0x1c, 0x48, 0xd4, 0xcf, 0x81, 0x4f, 0xb0, 0x5e, 0x70, 0x4e, 0x0b, 0xbc,
	0x3c, 0xcb, 0xa9, 0x86, 0x33, 0x2c, 0x61, 0x0e, 0xa5, 0x8a, 0xe5, 0x5c, 0x68, 0xe1, 0x1f, 0x1b,
	0x22, 0x76, 0x44, 0x5c, 0x13, 0x27, 0xbd, 0xa9, 0x98, 0x0a, 0x9b, 0x63, 0xf3, 0xe7, 0xd0, 0x93,
	0x90, 0x08, 0x55, 0x0a, 0x85, 0x73, 0x50, 0xf4, 0x51, 0x46, 0x04, 0xe3, 0x2e, 0x7f, 0xf7, 0x7d,
	0x1f, 0x75, 0xc6, 0xd6, 0xed, 0x17, 0xe8, 0xb0, 0x64, 0x3c, 0x9b, 0x50, 0x29, 0x14, 0xd3, 0x81,
	0xd7, 0xdf, 0x1b, 0x1c, 0x7e, 0x78, 0x13, 0x3b, 0x41, 0x6c, 0x04, 0xdb, 0x5e, 0xf1, 0x50, 0x30,
	0x9e, 0x9c, 0xde, 0xaf, 0xa2, 0xd6, 0xcf, 0xdf, 0xd1, 0x60, 0xca, 0xf4, 0x6c, 0x91, 0xc7, 0x44,
	0x94, 0xb8, 0xee, 0xe6, 0x3e, 0xef, 0xd5, 0xe4, 0x0e, 0xeb, 0xaf, 0x92, 0x2a, 0x7b, 0x40, 0xa5,
	0xa8, 0x64, 0xfc, 0xb3, 0xd3, 0xfb, 0x6f, 0xd1, 0x4b, 0xd3, 0x8d, 0x71, 0x4d, 0xe7, 0x4b, 0x28,
	0x82, 0x67, 0x7d, 0x6f, 0xd0, 0x4e, 0xcd, 0x04, 0xa3, 0xba, 0x64, 0x11, 0xa8, 0xfe, 0x23, 0x7b,
	0x35, 0x02, 0xd5, 0x23, 0xf2, 0x09, 0xbd, 0x76, 0x33, 0x2f, 0x19, 0x68, 0x26, 0x78, 0x96, 0x4b,
	0x15, 0xb4, 0x0d, 0x97, 0x1c, 0xaf, 0x57, 0x51, 0xf7, 0xda, 0x34, 0xac, 0xb3, 0x64, 0x7c, 0x93,
	0x76, 0xcb, 0x66, 0x41, 0x2a, 0x2b, 0x80, 0x6a, 0x47, 0xb0, 0xdf, 0x10, 0x40, 0xb5, 0x23, 0x68,
	0x16, 0xa4, 0xf2, 0x23, 0x64, 0x06, 0xca, 0x14, 0x9b, 0x72, 0x28, 0x54, 0xd0, 0xb1, 0x33, 0xa2,
	0x12, 0xaa, 0x1b, 0x57, 0xf1, 0x15, 0xea, 0x9a, 0xdd, 0x65, 0x12, 0xc8, 0x1d, 0xd5, 0xd9, 0x2d,
	0xa5, 0xc1, 0xf3, 0xa7, 0x5f, 0xed, 0x2b, 0x23, 0x19, 0xdb, 0x16, 0x97, 0x94, 0xfa, 0x09, 0xf2,
	0xa1, 0xa2, 0x05, 0xcc, 0x33, 0x96, 0x93, 0x8c, 0xcc, 0xc0, 0x5c, 0x95, 0xe0, 0x45, 0xdf, 0x1b,
	0x1c, 0x24, 0xbd, 0xf5, 0x2a, 0x3a, 0xba, 0xb0, 0xe9, 0x28, 0x19, 0x0e, 0x5d, 0x96, 0x1e, 0x39,
	0x7e, 0x94, 0x93, 0xba, 0xd2, 0x70, 0x4c, 0x4b, 0x99, 0x01, 0x21, 0x62, 0xc1, 0x75, 0x70, 0xb0,
	0xeb, 0xb8, 0xba, 0x1e, 0x5f, 0xb8, 0x6c, 0xeb, 0xb8, 0x2a, 0x65, 0x5d, 0xf1, 0x4f, 0x51, 0xaf,
	0x76, 0xdc, 0x52, 0x9a, 0xcd, 0x29, 0x61, 0x92, 0x51, 0xae, 0x03, 0x64, 0x2c, 0x69, 0xed, 0xbf,
	0xa4, 0x34, 0xdd, 0x26, 0x1f, 0xdb, 0x7f, 0x7f, 0x44, 0x5e, 0x32, 0xba, 0x5f, 0x87, 0xde, 0xc3,
	0x3a, 0xf4, 0xfe, 0xac, 0x43, 0xef, 0xdb, 0x26, 0x6c, 0x3d, 0x6c, 0xc2, 0xd6, 0xaf, 0x4d, 0xd8,
	0xfa, 0x82, 0x1b, 0x2b, 0x31, 0xcf, 0xc0, 0x5e, 0x63, 0x22, 0x0a, 0x4c, 0x66, 0xc0, 0x38, 0x5e,
	0x9e, 0xe3, 0x6a, 0xfb, 0x76, 0xec, 0x7e, 0xf2, 0x8e, 0x25, 0xce, 0xff, 0x05, 0x00, 0x00, 0xff,
	0xff, 0x6b, 0x5a, 0xbe, 0x7d, 0x57, 0x03, 0x00, 0x00,
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
	if len(this.MinDeposit) != len(that1.MinDeposit) {
		return false
	}
	for i := range this.MinDeposit {
		if !this.MinDeposit[i].Equal(&that1.MinDeposit[i]) {
			return false
		}
	}
	if this.MinInterval != that1.MinInterval {
		return false
	}
	if this.MaxInterval != that1.MaxInterval {
		return false
	}
	if this.MinDeviationBPS != that1.MinDeviationBPS {
		return false
	}
	if this.MaxDeviationBPS != that1.MaxDeviationBPS {
		return false
	}
	if this.MaxSignals != that1.MaxSignals {
		return false
	}
	if len(this.BasePacketFee) != len(that1.BasePacketFee) {
		return false
	}
	for i := range this.BasePacketFee {
		if !this.BasePacketFee[i].Equal(&that1.BasePacketFee[i]) {
			return false
		}
	}
	if this.AxelarIBCChannel != that1.AxelarIBCChannel {
		return false
	}
	if this.AxelarGMPAccount != that1.AxelarGMPAccount {
		return false
	}
	if this.AxelarFeeRecipient != that1.AxelarFeeRecipient {
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
	if len(m.AxelarFeeRecipient) > 0 {
		i -= len(m.AxelarFeeRecipient)
		copy(dAtA[i:], m.AxelarFeeRecipient)
		i = encodeVarintParams(dAtA, i, uint64(len(m.AxelarFeeRecipient)))
		i--
		dAtA[i] = 0x52
	}
	if len(m.AxelarGMPAccount) > 0 {
		i -= len(m.AxelarGMPAccount)
		copy(dAtA[i:], m.AxelarGMPAccount)
		i = encodeVarintParams(dAtA, i, uint64(len(m.AxelarGMPAccount)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.AxelarIBCChannel) > 0 {
		i -= len(m.AxelarIBCChannel)
		copy(dAtA[i:], m.AxelarIBCChannel)
		i = encodeVarintParams(dAtA, i, uint64(len(m.AxelarIBCChannel)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.BasePacketFee) > 0 {
		for iNdEx := len(m.BasePacketFee) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BasePacketFee[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintParams(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x3a
		}
	}
	if m.MaxSignals != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxSignals))
		i--
		dAtA[i] = 0x30
	}
	if m.MaxDeviationBPS != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxDeviationBPS))
		i--
		dAtA[i] = 0x28
	}
	if m.MinDeviationBPS != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MinDeviationBPS))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxInterval != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxInterval))
		i--
		dAtA[i] = 0x18
	}
	if m.MinInterval != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MinInterval))
		i--
		dAtA[i] = 0x10
	}
	if len(m.MinDeposit) > 0 {
		for iNdEx := len(m.MinDeposit) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.MinDeposit[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintParams(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
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
	if len(m.MinDeposit) > 0 {
		for _, e := range m.MinDeposit {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
	}
	if m.MinInterval != 0 {
		n += 1 + sovParams(uint64(m.MinInterval))
	}
	if m.MaxInterval != 0 {
		n += 1 + sovParams(uint64(m.MaxInterval))
	}
	if m.MinDeviationBPS != 0 {
		n += 1 + sovParams(uint64(m.MinDeviationBPS))
	}
	if m.MaxDeviationBPS != 0 {
		n += 1 + sovParams(uint64(m.MaxDeviationBPS))
	}
	if m.MaxSignals != 0 {
		n += 1 + sovParams(uint64(m.MaxSignals))
	}
	if len(m.BasePacketFee) > 0 {
		for _, e := range m.BasePacketFee {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
	}
	l = len(m.AxelarIBCChannel)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	l = len(m.AxelarGMPAccount)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	l = len(m.AxelarFeeRecipient)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinDeposit", wireType)
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
			m.MinDeposit = append(m.MinDeposit, types.Coin{})
			if err := m.MinDeposit[len(m.MinDeposit)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinInterval", wireType)
			}
			m.MinInterval = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MinInterval |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxInterval", wireType)
			}
			m.MaxInterval = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxInterval |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinDeviationBPS", wireType)
			}
			m.MinDeviationBPS = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MinDeviationBPS |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxDeviationBPS", wireType)
			}
			m.MaxDeviationBPS = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxDeviationBPS |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxSignals", wireType)
			}
			m.MaxSignals = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxSignals |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BasePacketFee", wireType)
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
			m.BasePacketFee = append(m.BasePacketFee, types.Coin{})
			if err := m.BasePacketFee[len(m.BasePacketFee)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AxelarIBCChannel", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AxelarIBCChannel = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AxelarGMPAccount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AxelarGMPAccount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AxelarFeeRecipient", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AxelarFeeRecipient = string(dAtA[iNdEx:postIndex])
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
