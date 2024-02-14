// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: feeds/v1beta1/params.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

// Params is the data structure that keeps the parameters of the feeds module.
type Params struct {
	// The address of the admin that is allowed to perform operations on modules.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// allow_diff_time is the allowed difference (in seconds) between timestamp and block_time when validator submits the
	// prices.
	AllowDiffTime int64 `protobuf:"varint,2,opt,name=allow_diff_time,json=allowDiffTime,proto3" json:"allow_diff_time,omitempty"`
	// transition_time is the time (in seconds) given for validators to adapt to changing in symbol's interval.
	TransitionTime int64 `protobuf:"varint,3,opt,name=transition_time,json=transitionTime,proto3" json:"transition_time,omitempty"`
	// min_interval is the minimum limit of every symbols' interval (in seconds).
	// If the calculated interval is lower than this, it will be capped at this value.
	MinInterval int64 `protobuf:"varint,4,opt,name=min_interval,json=minInterval,proto3" json:"min_interval,omitempty"`
	// max_interval is the maximum limit of every symbols' interval (in seconds).
	// If the calculated interval of a symbol is higher than this, it will not be recognized as a supported symbol.
	MaxInterval int64 `protobuf:"varint,5,opt,name=max_interval,json=maxInterval,proto3" json:"max_interval,omitempty"`
	// power_threshold is the amount of minimum power required to put symbol in the supported list.
	PowerThreshold int64 `protobuf:"varint,6,opt,name=power_threshold,json=powerThreshold,proto3" json:"power_threshold,omitempty"`
	// max_support_symbol is the maximum number of symbols supported at a time.
	MaxSupportedSymbol int64 `protobuf:"varint,7,opt,name=max_supported_symbol,json=maxSupportedSymbol,proto3" json:"max_supported_symbol,omitempty"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_bbfae8ad171874f3, []int{0}
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

func (m *Params) GetAdmin() string {
	if m != nil {
		return m.Admin
	}
	return ""
}

func (m *Params) GetAllowDiffTime() int64 {
	if m != nil {
		return m.AllowDiffTime
	}
	return 0
}

func (m *Params) GetTransitionTime() int64 {
	if m != nil {
		return m.TransitionTime
	}
	return 0
}

func (m *Params) GetMinInterval() int64 {
	if m != nil {
		return m.MinInterval
	}
	return 0
}

func (m *Params) GetMaxInterval() int64 {
	if m != nil {
		return m.MaxInterval
	}
	return 0
}

func (m *Params) GetPowerThreshold() int64 {
	if m != nil {
		return m.PowerThreshold
	}
	return 0
}

func (m *Params) GetMaxSupportedSymbol() int64 {
	if m != nil {
		return m.MaxSupportedSymbol
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "feeds.v1beta1.Params")
}

func init() { proto.RegisterFile("feeds/v1beta1/params.proto", fileDescriptor_bbfae8ad171874f3) }

var fileDescriptor_bbfae8ad171874f3 = []byte{
	// 361 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0xd1, 0xb1, 0x8e, 0xda, 0x30,
	0x00, 0xc6, 0xf1, 0x04, 0x0a, 0x6d, 0xdd, 0x52, 0xa4, 0x88, 0x21, 0x65, 0x08, 0xb4, 0x43, 0x61,
	0x69, 0x5c, 0xda, 0xad, 0x5b, 0x51, 0x87, 0xbb, 0xed, 0x04, 0x4c, 0xb7, 0x44, 0x4e, 0xec, 0x24,
	0x96, 0x62, 0x3b, 0xb2, 0x0d, 0x84, 0xb7, 0xb8, 0xf1, 0x46, 0x1e, 0x82, 0x87, 0xb8, 0x11, 0xdd,
	0x74, 0xe3, 0x09, 0x96, 0x7b, 0x8c, 0x13, 0x76, 0x38, 0xb6, 0xe4, 0xfb, 0xff, 0xe4, 0x48, 0x31,
	0xe8, 0xa7, 0x84, 0x60, 0x05, 0x57, 0x93, 0x98, 0x68, 0x34, 0x81, 0x25, 0x92, 0x88, 0xa9, 0xb0,
	0x94, 0x42, 0x0b, 0xaf, 0x63, 0x5a, 0x58, 0xb7, 0x7e, 0x2f, 0x13, 0x99, 0x30, 0x05, 0x9e, 0x9e,
	0x2c, 0xea, 0x7f, 0x4d, 0x84, 0x62, 0x42, 0x45, 0x36, 0xd8, 0x17, 0x9b, 0xbe, 0xef, 0x1a, 0xa0,
	0x7d, 0x63, 0x0e, 0xf4, 0x42, 0xd0, 0x42, 0x98, 0x51, 0xee, 0xbb, 0x43, 0x77, 0xfc, 0x71, 0xea,
	0x3f, 0xee, 0x7e, 0xf6, 0x6a, 0xfb, 0x0f, 0x63, 0x49, 0x94, 0x9a, 0x6b, 0x49, 0x79, 0x36, 0xb3,
	0xcc, 0xfb, 0x01, 0xba, 0xa8, 0x28, 0xc4, 0x3a, 0xc2, 0x34, 0x4d, 0x23, 0x4d, 0x19, 0xf1, 0x1b,
	0x43, 0x77, 0xdc, 0x9c, 0x75, 0xcc, 0xfc, 0x9f, 0xa6, 0xe9, 0x82, 0x32, 0xe2, 0x8d, 0x40, 0x57,
	0x4b, 0xc4, 0x15, 0xd5, 0x54, 0x70, 0xeb, 0x9a, 0xc6, 0x7d, 0xb9, 0xcc, 0x06, 0x7e, 0x03, 0x9f,
	0x19, 0xe5, 0x11, 0xe5, 0x9a, 0xc8, 0x15, 0x2a, 0xfc, 0x77, 0x46, 0x7d, 0x62, 0x94, 0x5f, 0xd7,
	0x93, 0x21, 0xa8, 0xba, 0x90, 0x56, 0x4d, 0x50, 0xf5, 0x46, 0x46, 0xa0, 0x5b, 0x8a, 0x35, 0x91,
	0x91, 0xce, 0x25, 0x51, 0xb9, 0x28, 0xb0, 0xdf, 0xb6, 0x9f, 0x33, 0xf3, 0xe2, 0xbc, 0x7a, 0xbf,
	0x40, 0xef, 0x74, 0x96, 0x5a, 0x96, 0xa5, 0x90, 0x9a, 0xe0, 0x48, 0x6d, 0x58, 0x2c, 0x0a, 0xff,
	0xbd, 0xd1, 0x1e, 0x43, 0xd5, 0xfc, 0x9c, 0xe6, 0xa6, 0xfc, 0xfd, 0x70, 0xbf, 0x1d, 0x38, 0x2f,
	0xdb, 0x81, 0x3b, 0xbd, 0x7a, 0x38, 0x04, 0xee, 0xfe, 0x10, 0xb8, 0xcf, 0x87, 0xc0, 0xbd, 0x3b,
	0x06, 0xce, 0xfe, 0x18, 0x38, 0x4f, 0xc7, 0xc0, 0xb9, 0x0d, 0x33, 0xaa, 0xf3, 0x65, 0x1c, 0x26,
	0x82, 0xc1, 0x18, 0x71, 0x6c, 0x7e, 0x73, 0x22, 0x0a, 0x98, 0xe4, 0x88, 0x72, 0xb8, 0xfa, 0x0d,
	0x2b, 0x68, 0xef, 0x53, 0x6f, 0x4a, 0xa2, 0xe2, 0xb6, 0x01, 0x7f, 0x5e, 0x03, 0x00, 0x00, 0xff,
	0xff, 0x72, 0x4c, 0x15, 0xe6, 0xe5, 0x01, 0x00, 0x00,
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
	if this.Admin != that1.Admin {
		return false
	}
	if this.AllowDiffTime != that1.AllowDiffTime {
		return false
	}
	if this.TransitionTime != that1.TransitionTime {
		return false
	}
	if this.MinInterval != that1.MinInterval {
		return false
	}
	if this.MaxInterval != that1.MaxInterval {
		return false
	}
	if this.PowerThreshold != that1.PowerThreshold {
		return false
	}
	if this.MaxSupportedSymbol != that1.MaxSupportedSymbol {
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
	if m.MaxSupportedSymbol != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxSupportedSymbol))
		i--
		dAtA[i] = 0x38
	}
	if m.PowerThreshold != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.PowerThreshold))
		i--
		dAtA[i] = 0x30
	}
	if m.MaxInterval != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MaxInterval))
		i--
		dAtA[i] = 0x28
	}
	if m.MinInterval != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.MinInterval))
		i--
		dAtA[i] = 0x20
	}
	if m.TransitionTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.TransitionTime))
		i--
		dAtA[i] = 0x18
	}
	if m.AllowDiffTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.AllowDiffTime))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Admin) > 0 {
		i -= len(m.Admin)
		copy(dAtA[i:], m.Admin)
		i = encodeVarintParams(dAtA, i, uint64(len(m.Admin)))
		i--
		dAtA[i] = 0xa
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
	l = len(m.Admin)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	if m.AllowDiffTime != 0 {
		n += 1 + sovParams(uint64(m.AllowDiffTime))
	}
	if m.TransitionTime != 0 {
		n += 1 + sovParams(uint64(m.TransitionTime))
	}
	if m.MinInterval != 0 {
		n += 1 + sovParams(uint64(m.MinInterval))
	}
	if m.MaxInterval != 0 {
		n += 1 + sovParams(uint64(m.MaxInterval))
	}
	if m.PowerThreshold != 0 {
		n += 1 + sovParams(uint64(m.PowerThreshold))
	}
	if m.MaxSupportedSymbol != 0 {
		n += 1 + sovParams(uint64(m.MaxSupportedSymbol))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
			m.Admin = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowDiffTime", wireType)
			}
			m.AllowDiffTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AllowDiffTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TransitionTime", wireType)
			}
			m.TransitionTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TransitionTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
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
				m.MinInterval |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
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
				m.MaxInterval |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PowerThreshold", wireType)
			}
			m.PowerThreshold = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PowerThreshold |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxSupportedSymbol", wireType)
			}
			m.MaxSupportedSymbol = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxSupportedSymbol |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
