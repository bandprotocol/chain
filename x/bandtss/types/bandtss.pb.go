// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: bandtss/v1beta1/bandtss.proto

package types

import (
	fmt "fmt"
	github_com_bandprotocol_chain_v2_pkg_tss "github.com/bandprotocol/chain/v2/pkg/tss"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// MemberStatus is an enumeration of the possible statuses of a member.
type MemberStatus int32

const (
	// MEMBER_STATUS_UNSPECIFIED is unknown status.
	MEMBER_STATUS_UNSPECIFIED MemberStatus = 0
	// MEMBER_STATUS_ACTIVE is the active status.
	MEMBER_STATUS_ACTIVE MemberStatus = 1
	// MEMBER_STATUS_INACTIVE is the inactive status.
	MEMBER_STATUS_INACTIVE MemberStatus = 2
	// MEMBER_STATUS_JAIL is the status when member is jailed.
	MEMBER_STATUS_JAIL MemberStatus = 3
)

var MemberStatus_name = map[int32]string{
	0: "MEMBER_STATUS_UNSPECIFIED",
	1: "MEMBER_STATUS_ACTIVE",
	2: "MEMBER_STATUS_INACTIVE",
	3: "MEMBER_STATUS_JAIL",
}

var MemberStatus_value = map[string]int32{
	"MEMBER_STATUS_UNSPECIFIED": 0,
	"MEMBER_STATUS_ACTIVE":      1,
	"MEMBER_STATUS_INACTIVE":    2,
	"MEMBER_STATUS_JAIL":        3,
}

func (x MemberStatus) String() string {
	return proto.EnumName(MemberStatus_name, int32(x))
}

func (MemberStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2effaef066b71284, []int{0}
}

// Status maintains whether a member is an active member.
type Status struct {
	// address is the address of the member.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// status represents the current status of the member
	Status MemberStatus `protobuf:"varint,2,opt,name=status,proto3,enum=bandtss.v1beta1.MemberStatus" json:"status,omitempty"`
	// since is a block timestamp when a member has been activated/deactivated/jailed
	Since time.Time `protobuf:"bytes,3,opt,name=since,proto3,stdtime" json:"since"`
	// last_active is a latest block timestamp when a member is active
	LastActive time.Time `protobuf:"bytes,4,opt,name=last_active,json=lastActive,proto3,stdtime" json:"last_active"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_2effaef066b71284, []int{0}
}
func (m *Status) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Status.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(m, src)
}
func (m *Status) XXX_Size() int {
	return m.Size()
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Status) GetStatus() MemberStatus {
	if m != nil {
		return m.Status
	}
	return MEMBER_STATUS_UNSPECIFIED
}

func (m *Status) GetSince() time.Time {
	if m != nil {
		return m.Since
	}
	return time.Time{}
}

func (m *Status) GetLastActive() time.Time {
	if m != nil {
		return m.LastActive
	}
	return time.Time{}
}

type SigningFee struct {
	// signing_ids is a list of signing IDs.
	SigningID github_com_bandprotocol_chain_v2_pkg_tss.SigningID `protobuf:"varint,1,opt,name=signing_id,json=signingId,proto3,casttype=github.com/bandprotocol/chain/v2/pkg/tss.SigningID" json:"signing_id,omitempty"`
	// fee is the total tokens that will be paid for this signing
	Fee github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=fee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"fee"`
	// requester is the address of requester who paid for the Bandtss fee.
	Requester string `protobuf:"bytes,3,opt,name=requester,proto3" json:"requester,omitempty"`
}

func (m *SigningFee) Reset()         { *m = SigningFee{} }
func (m *SigningFee) String() string { return proto.CompactTextString(m) }
func (*SigningFee) ProtoMessage()    {}
func (*SigningFee) Descriptor() ([]byte, []int) {
	return fileDescriptor_2effaef066b71284, []int{1}
}
func (m *SigningFee) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SigningFee) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SigningFee.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SigningFee) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SigningFee.Merge(m, src)
}
func (m *SigningFee) XXX_Size() int {
	return m.Size()
}
func (m *SigningFee) XXX_DiscardUnknown() {
	xxx_messageInfo_SigningFee.DiscardUnknown(m)
}

var xxx_messageInfo_SigningFee proto.InternalMessageInfo

func (m *SigningFee) GetSigningID() github_com_bandprotocol_chain_v2_pkg_tss.SigningID {
	if m != nil {
		return m.SigningID
	}
	return 0
}

func (m *SigningFee) GetFee() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Fee
	}
	return nil
}

func (m *SigningFee) GetRequester() string {
	if m != nil {
		return m.Requester
	}
	return ""
}

func init() {
	proto.RegisterEnum("bandtss.v1beta1.MemberStatus", MemberStatus_name, MemberStatus_value)
	proto.RegisterType((*Status)(nil), "bandtss.v1beta1.Status")
	proto.RegisterType((*SigningFee)(nil), "bandtss.v1beta1.SigningFee")
}

func init() { proto.RegisterFile("bandtss/v1beta1/bandtss.proto", fileDescriptor_2effaef066b71284) }

var fileDescriptor_2effaef066b71284 = []byte{
	// 528 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x3f, 0x6f, 0xd3, 0x40,
	0x1c, 0xf5, 0x35, 0x21, 0x90, 0x0b, 0x82, 0xe8, 0x54, 0x55, 0x4e, 0x44, 0xec, 0xa8, 0x53, 0x84,
	0x84, 0xaf, 0x0d, 0x62, 0xe9, 0x96, 0xa4, 0xae, 0x64, 0x44, 0x22, 0xe4, 0xa4, 0x0c, 0x48, 0x28,
	0xf2, 0x9f, 0xab, 0x6b, 0x35, 0xf1, 0x85, 0xdc, 0x25, 0x82, 0x91, 0x8d, 0xb1, 0x1f, 0x01, 0x89,
	0x05, 0xf1, 0x49, 0x3a, 0x76, 0x64, 0x4a, 0x91, 0xb3, 0xb0, 0xb3, 0x31, 0x21, 0xdf, 0x9d, 0x43,
	0xc3, 0x82, 0x98, 0xec, 0xe7, 0xf7, 0x7e, 0xef, 0xee, 0xbd, 0x9f, 0x0c, 0x1b, 0xbe, 0x97, 0x84,
	0x9c, 0x31, 0xbc, 0x3c, 0xf4, 0x09, 0xf7, 0x0e, 0xb1, 0xc2, 0xd6, 0x6c, 0x4e, 0x39, 0x45, 0x0f,
	0x73, 0xa8, 0xe8, 0xfa, 0x6e, 0x44, 0x23, 0x2a, 0x38, 0x9c, 0xbd, 0x49, 0x59, 0xdd, 0x8c, 0x28,
	0x8d, 0x26, 0x04, 0x0b, 0xe4, 0x2f, 0xce, 0x30, 0x8f, 0xa7, 0x84, 0x71, 0x6f, 0x3a, 0x53, 0x02,
	0x23, 0xa0, 0x6c, 0x4a, 0x19, 0xf6, 0x3d, 0x46, 0x36, 0x47, 0x05, 0x34, 0x4e, 0x14, 0x5f, 0x93,
	0xfc, 0x58, 0x3a, 0x4b, 0x20, 0xa9, 0xfd, 0x14, 0xc0, 0xd2, 0x90, 0x7b, 0x7c, 0xc1, 0x90, 0x0e,
	0xef, 0x7a, 0x61, 0x38, 0x27, 0x8c, 0xe9, 0xa0, 0x09, 0x5a, 0x65, 0x37, 0x87, 0xe8, 0x19, 0x2c,
	0x31, 0xa1, 0xd1, 0x77, 0x9a, 0xa0, 0xf5, 0xa0, 0xdd, 0xb0, 0xfe, 0xba, 0xb8, 0xd5, 0x27, 0x53,
	0x9f, 0xcc, 0xa5, 0x91, 0xab, 0xc4, 0xe8, 0x08, 0xde, 0x61, 0x71, 0x12, 0x10, 0xbd, 0xd0, 0x04,
	0xad, 0x4a, 0xbb, 0x6e, 0xc9, 0x1c, 0x56, 0x9e, 0xc3, 0x1a, 0xe5, 0x39, 0xba, 0xf7, 0xae, 0x56,
	0xa6, 0x76, 0x79, 0x63, 0x02, 0x57, 0x8e, 0x20, 0x1b, 0x56, 0x26, 0x1e, 0xe3, 0x63, 0x2f, 0xe0,
	0xf1, 0x92, 0xe8, 0xc5, 0xff, 0x70, 0x80, 0xd9, 0x60, 0x47, 0xcc, 0x1d, 0x15, 0x7f, 0x7c, 0x32,
	0xc1, 0xfe, 0x4f, 0x00, 0xe1, 0x30, 0x8e, 0x92, 0x38, 0x89, 0x4e, 0x08, 0x41, 0x3e, 0x84, 0x4c,
	0xa2, 0x71, 0x1c, 0x8a, 0xac, 0xc5, 0x6e, 0x2f, 0x5d, 0x99, 0x65, 0xa5, 0x71, 0x8e, 0x7f, 0xad,
	0xcc, 0x76, 0x14, 0xf3, 0xf3, 0x85, 0x6f, 0x05, 0x74, 0x2a, 0xb6, 0x26, 0x8e, 0x0c, 0xe8, 0x04,
	0x07, 0xe7, 0x5e, 0x9c, 0xe0, 0x65, 0x1b, 0xcf, 0x2e, 0x22, 0x9c, 0x75, 0xb0, 0x99, 0x72, 0xcb,
	0xca, 0xd6, 0x09, 0xd1, 0x1b, 0x58, 0x38, 0x23, 0x44, 0xdf, 0x69, 0x16, 0x5a, 0x95, 0x76, 0xcd,
	0x52, 0x9d, 0x67, 0x0b, 0xda, 0x74, 0xd6, 0xa3, 0x71, 0xd2, 0x3d, 0xc8, 0xae, 0xfd, 0xf5, 0xc6,
	0x6c, 0xdd, 0x3a, 0x4e, 0x6d, 0x53, 0x3e, 0x9e, 0xb0, 0xf0, 0x02, 0xf3, 0xf7, 0x33, 0xc2, 0xc4,
	0x00, 0x73, 0x33, 0x5f, 0xf4, 0x08, 0x96, 0xe7, 0xe4, 0xed, 0x82, 0x30, 0x4e, 0xe6, 0xa2, 0xde,
	0xb2, 0xfb, 0xe7, 0x83, 0x4c, 0xfd, 0xf8, 0x03, 0x80, 0xf7, 0x6f, 0xef, 0x05, 0x35, 0x60, 0xad,
	0x6f, 0xf7, 0xbb, 0xb6, 0x3b, 0x1e, 0x8e, 0x3a, 0xa3, 0xd3, 0xe1, 0xf8, 0x74, 0x30, 0x7c, 0x69,
	0xf7, 0x9c, 0x13, 0xc7, 0x3e, 0xae, 0x6a, 0x48, 0x87, 0xbb, 0xdb, 0x74, 0xa7, 0x37, 0x72, 0x5e,
	0xd9, 0x55, 0x80, 0xea, 0x70, 0x6f, 0x9b, 0x71, 0x06, 0x8a, 0xdb, 0x41, 0x7b, 0x10, 0x6d, 0x73,
	0xcf, 0x3b, 0xce, 0x8b, 0x6a, 0xa1, 0x5e, 0xfc, 0xf8, 0xd9, 0xd0, 0xba, 0x83, 0x2f, 0xa9, 0x01,
	0xae, 0x52, 0x03, 0x5c, 0xa7, 0x06, 0xf8, 0x9e, 0x1a, 0xe0, 0x72, 0x6d, 0x68, 0xd7, 0x6b, 0x43,
	0xfb, 0xb6, 0x36, 0xb4, 0xd7, 0x07, 0xff, 0xec, 0xf8, 0x5d, 0xfe, 0xc7, 0xc8, 0x0a, 0xfc, 0x92,
	0x90, 0x3c, 0xfd, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xfd, 0xbf, 0x21, 0x1d, 0x59, 0x03, 0x00, 0x00,
}

func (this *Status) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Status)
	if !ok {
		that2, ok := that.(Status)
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
	if this.Address != that1.Address {
		return false
	}
	if this.Status != that1.Status {
		return false
	}
	if !this.Since.Equal(that1.Since) {
		return false
	}
	if !this.LastActive.Equal(that1.LastActive) {
		return false
	}
	return true
}
func (this *SigningFee) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SigningFee)
	if !ok {
		that2, ok := that.(SigningFee)
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
	if this.SigningID != that1.SigningID {
		return false
	}
	if len(this.Fee) != len(that1.Fee) {
		return false
	}
	for i := range this.Fee {
		if !this.Fee[i].Equal(&that1.Fee[i]) {
			return false
		}
	}
	if this.Requester != that1.Requester {
		return false
	}
	return true
}
func (m *Status) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Status) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Status) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LastActive, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastActive):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintBandtss(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x22
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.Since, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.Since):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintBandtss(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x1a
	if m.Status != 0 {
		i = encodeVarintBandtss(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintBandtss(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SigningFee) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SigningFee) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SigningFee) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Requester) > 0 {
		i -= len(m.Requester)
		copy(dAtA[i:], m.Requester)
		i = encodeVarintBandtss(dAtA, i, uint64(len(m.Requester)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Fee) > 0 {
		for iNdEx := len(m.Fee) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Fee[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBandtss(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.SigningID != 0 {
		i = encodeVarintBandtss(dAtA, i, uint64(m.SigningID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintBandtss(dAtA []byte, offset int, v uint64) int {
	offset -= sovBandtss(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Status) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovBandtss(uint64(l))
	}
	if m.Status != 0 {
		n += 1 + sovBandtss(uint64(m.Status))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.Since)
	n += 1 + l + sovBandtss(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastActive)
	n += 1 + l + sovBandtss(uint64(l))
	return n
}

func (m *SigningFee) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SigningID != 0 {
		n += 1 + sovBandtss(uint64(m.SigningID))
	}
	if len(m.Fee) > 0 {
		for _, e := range m.Fee {
			l = e.Size()
			n += 1 + l + sovBandtss(uint64(l))
		}
	}
	l = len(m.Requester)
	if l > 0 {
		n += 1 + l + sovBandtss(uint64(l))
	}
	return n
}

func sovBandtss(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozBandtss(x uint64) (n int) {
	return sovBandtss(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Status) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBandtss
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
			return fmt.Errorf("proto: Status: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Status: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
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
				return ErrInvalidLengthBandtss
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthBandtss
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= MemberStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Since", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
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
				return ErrInvalidLengthBandtss
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBandtss
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.Since, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastActive", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
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
				return ErrInvalidLengthBandtss
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBandtss
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LastActive, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBandtss(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBandtss
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
func (m *SigningFee) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBandtss
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
			return fmt.Errorf("proto: SigningFee: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SigningFee: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigningID", wireType)
			}
			m.SigningID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SigningID |= github_com_bandprotocol_chain_v2_pkg_tss.SigningID(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
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
				return ErrInvalidLengthBandtss
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBandtss
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Fee = append(m.Fee, types.Coin{})
			if err := m.Fee[len(m.Fee)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Requester", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBandtss
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
				return ErrInvalidLengthBandtss
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthBandtss
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Requester = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBandtss(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBandtss
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
func skipBandtss(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBandtss
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
					return 0, ErrIntOverflowBandtss
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
					return 0, ErrIntOverflowBandtss
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
				return 0, ErrInvalidLengthBandtss
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupBandtss
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthBandtss
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthBandtss        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBandtss          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupBandtss = fmt.Errorf("proto: unexpected end of group")
)
