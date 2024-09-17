// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tunnel/v1beta1/params.proto

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
	// min_deposit is the minimum deposit required to create a tunnel
	MinDeposit github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=min_deposit,json=minDeposit,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"min_deposit"`
	// min_interval is the minimum interval in seconds
	MinInterval uint64 `protobuf:"varint,2,opt,name=min_interval,json=minInterval,proto3" json:"min_interval,omitempty"`
	// max_signals defines the maximum number of signals allowed per tunnel.
	MaxSignals uint64 `protobuf:"varint,3,opt,name=max_signals,json=maxSignals,proto3" json:"max_signals,omitempty"`
	// base_packet_fee is the base fee for each packet
	BasePacketFee github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,4,rep,name=base_packet_fee,json=basePacketFee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"base_packet_fee"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_a7b5eedd244355eb, []int{0}
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

func init() {
	proto.RegisterType((*Params)(nil), "tunnel.v1beta1.Params")
}

func init() { proto.RegisterFile("tunnel/v1beta1/params.proto", fileDescriptor_a7b5eedd244355eb) }

var fileDescriptor_a7b5eedd244355eb = []byte{
	// 335 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x91, 0x31, 0x4e, 0xf3, 0x30,
	0x1c, 0xc5, 0x93, 0xb6, 0xaa, 0x3e, 0xb9, 0x1f, 0x20, 0x45, 0x0c, 0xa1, 0x48, 0x4e, 0x61, 0xea,
	0x42, 0x4c, 0xcb, 0xc6, 0x58, 0x10, 0x52, 0xb7, 0xaa, 0x6c, 0x2c, 0x91, 0xe3, 0x9a, 0xd4, 0x6a,
	0x6c, 0x47, 0xb5, 0x5b, 0x95, 0x5b, 0x30, 0x32, 0x76, 0xe6, 0x06, 0xdc, 0xa0, 0x63, 0x47, 0x26,
	0x40, 0xcd, 0xc2, 0x31, 0x90, 0xed, 0x80, 0x38, 0x00, 0x93, 0xad, 0xff, 0x7b, 0xfe, 0xbd, 0x27,
	0xff, 0xc1, 0xb1, 0x5e, 0x08, 0x41, 0x73, 0xb4, 0xec, 0xa5, 0x54, 0xe3, 0x1e, 0x2a, 0xf0, 0x1c,
	0x73, 0x15, 0x17, 0x73, 0xa9, 0x65, 0xb0, 0xef, 0xc4, 0xb8, 0x12, 0xdb, 0x87, 0x99, 0xcc, 0xa4,
	0x95, 0x90, 0xb9, 0x39, 0x57, 0x1b, 0x12, 0xa9, 0xb8, 0x54, 0x28, 0xc5, 0x8a, 0xfe, 0x70, 0x88,
	0x64, 0xc2, 0xe9, 0xa7, 0x2f, 0x35, 0xd0, 0x1c, 0x59, 0x6c, 0x90, 0x83, 0x16, 0x67, 0x22, 0x99,
	0xd0, 0x42, 0x2a, 0xa6, 0x43, 0xbf, 0x53, 0xef, 0xb6, 0xfa, 0x47, 0xb1, 0x03, 0xc4, 0x06, 0xf0,
	0x9d, 0x15, 0x5f, 0x49, 0x26, 0x06, 0xe7, 0x9b, 0xb7, 0xc8, 0x7b, 0x7e, 0x8f, 0xba, 0x19, 0xd3,
	0xd3, 0x45, 0x1a, 0x13, 0xc9, 0x51, 0x95, 0xe6, 0x8e, 0x33, 0x35, 0x99, 0x21, 0xfd, 0x50, 0x50,
	0x65, 0x1f, 0xa8, 0x31, 0xe0, 0x4c, 0x5c, 0x3b, 0x7c, 0x70, 0x02, 0xfe, 0x9b, 0x34, 0x26, 0x34,
	0x9d, 0x2f, 0x71, 0x1e, 0xd6, 0x3a, 0x7e, 0xb7, 0x31, 0x36, 0x0d, 0x86, 0xd5, 0x28, 0x88, 0x40,
	0x8b, 0xe3, 0x55, 0xa2, 0x58, 0x26, 0x70, 0xae, 0xc2, 0xba, 0x75, 0x00, 0x8e, 0x57, 0xb7, 0x6e,
	0x12, 0x28, 0x70, 0x60, 0x6a, 0x25, 0x05, 0x26, 0x33, 0xaa, 0x93, 0x7b, 0x4a, 0xc3, 0xc6, 0xdf,
	0xb7, 0xde, 0x33, 0x90, 0x91, 0x8d, 0xb8, 0xa1, 0xf4, 0xf2, 0xdf, 0xd3, 0x3a, 0xf2, 0x3e, 0xd7,
	0x91, 0x3f, 0x18, 0x6e, 0x76, 0xd0, 0xdf, 0xee, 0xa0, 0xff, 0xb1, 0x83, 0xfe, 0x63, 0x09, 0xbd,
	0x6d, 0x09, 0xbd, 0xd7, 0x12, 0x7a, 0x77, 0xe8, 0x17, 0x3c, 0xc5, 0x62, 0x62, 0xff, 0x9a, 0xc8,
	0x1c, 0x91, 0x29, 0x66, 0x02, 0x2d, 0xfb, 0x68, 0x85, 0xaa, 0xdd, 0xda, 0xa4, 0xb4, 0x69, 0x1d,
	0x17, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3f, 0x54, 0xc7, 0x3d, 0xf2, 0x01, 0x00, 0x00,
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
			dAtA[i] = 0x22
		}
	}
	if m.MaxSignals != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxSignals))
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
	if m.MaxSignals != 0 {
		n += 1 + sovParams(uint64(m.MaxSignals))
	}
	if len(m.BasePacketFee) > 0 {
		for _, e := range m.BasePacketFee {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
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
		case 4:
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