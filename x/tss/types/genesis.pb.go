// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: band/tss/v1beta1/genesis.proto

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

// GenesisState defines the tss module's genesis state.
type GenesisState struct {
	// params defines all the paramiters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// groups is an array containing information about each group.
	Groups []Group `protobuf:"bytes,2,rep,name=groups,proto3" json:"groups"`
	// members is an array containing information about each member of groups.
	Members []Member `protobuf:"bytes,3,rep,name=members,proto3" json:"members"`
	// des is an array containing the des of all the addressres.
	DEs []DEGenesis `protobuf:"bytes,4,rep,name=des,proto3" json:"des"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_26d9273eff41c101, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetGroups() []Group {
	if m != nil {
		return m.Groups
	}
	return nil
}

func (m *GenesisState) GetMembers() []Member {
	if m != nil {
		return m.Members
	}
	return nil
}

func (m *GenesisState) GetDEs() []DEGenesis {
	if m != nil {
		return m.DEs
	}
	return nil
}

// Params defines the set of module parameters.
type Params struct {
	// max_group_size is the maximum of the member capacity of the group.
	MaxGroupSize uint64 `protobuf:"varint,1,opt,name=max_group_size,json=maxGroupSize,proto3" json:"max_group_size,omitempty"`
	// max_d_e_size is the maximum of the de capacity of the member.
	MaxDESize uint64 `protobuf:"varint,2,opt,name=max_d_e_size,json=maxDESize,proto3" json:"max_d_e_size,omitempty"`
	// creation_period is the number of blocks allowed to creating tss group.
	CreationPeriod uint64 `protobuf:"varint,3,opt,name=creation_period,json=creationPeriod,proto3" json:"creation_period,omitempty"`
	// signing_period is the number of blocks allowed to sign.
	SigningPeriod uint64 `protobuf:"varint,4,opt,name=signing_period,json=signingPeriod,proto3" json:"signing_period,omitempty"`
	// max_signing_attempt is the maximum number of signing retry process per signingID.
	MaxSigningAttempt uint64 `protobuf:"varint,5,opt,name=max_signing_attempt,json=maxSigningAttempt,proto3" json:"max_signing_attempt,omitempty"`
	// max_memo_length is the maximum length of the memo in the direct originator.
	MaxMemoLength uint64 `protobuf:"varint,6,opt,name=max_memo_length,json=maxMemoLength,proto3" json:"max_memo_length,omitempty"`
	// max_message_length is the maximum length of the message in the TextSignatureOrder.
	MaxMessageLength uint64 `protobuf:"varint,7,opt,name=max_message_length,json=maxMessageLength,proto3" json:"max_message_length,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_26d9273eff41c101, []int{1}
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

func (m *Params) GetMaxGroupSize() uint64 {
	if m != nil {
		return m.MaxGroupSize
	}
	return 0
}

func (m *Params) GetMaxDESize() uint64 {
	if m != nil {
		return m.MaxDESize
	}
	return 0
}

func (m *Params) GetCreationPeriod() uint64 {
	if m != nil {
		return m.CreationPeriod
	}
	return 0
}

func (m *Params) GetSigningPeriod() uint64 {
	if m != nil {
		return m.SigningPeriod
	}
	return 0
}

func (m *Params) GetMaxSigningAttempt() uint64 {
	if m != nil {
		return m.MaxSigningAttempt
	}
	return 0
}

func (m *Params) GetMaxMemoLength() uint64 {
	if m != nil {
		return m.MaxMemoLength
	}
	return 0
}

func (m *Params) GetMaxMessageLength() uint64 {
	if m != nil {
		return m.MaxMessageLength
	}
	return 0
}

// DEGenesis defines an account address and de pair used in the tss module's genesis state.
type DEGenesis struct {
	// address is the address of the de holder.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// de defines the difference de this balance holds.
	DE DE `protobuf:"bytes,2,opt,name=de,proto3" json:"de"`
}

func (m *DEGenesis) Reset()         { *m = DEGenesis{} }
func (m *DEGenesis) String() string { return proto.CompactTextString(m) }
func (*DEGenesis) ProtoMessage()    {}
func (*DEGenesis) Descriptor() ([]byte, []int) {
	return fileDescriptor_26d9273eff41c101, []int{2}
}
func (m *DEGenesis) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DEGenesis) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DEGenesis.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DEGenesis) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DEGenesis.Merge(m, src)
}
func (m *DEGenesis) XXX_Size() int {
	return m.Size()
}
func (m *DEGenesis) XXX_DiscardUnknown() {
	xxx_messageInfo_DEGenesis.DiscardUnknown(m)
}

var xxx_messageInfo_DEGenesis proto.InternalMessageInfo

func (m *DEGenesis) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *DEGenesis) GetDE() DE {
	if m != nil {
		return m.DE
	}
	return DE{}
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "band.tss.v1beta1.GenesisState")
	proto.RegisterType((*Params)(nil), "band.tss.v1beta1.Params")
	proto.RegisterType((*DEGenesis)(nil), "band.tss.v1beta1.DEGenesis")
}

func init() { proto.RegisterFile("band/tss/v1beta1/genesis.proto", fileDescriptor_26d9273eff41c101) }

var fileDescriptor_26d9273eff41c101 = []byte{
	// 524 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0x3d, 0x6f, 0xd3, 0x40,
	0x18, 0xc7, 0x63, 0x27, 0x24, 0xca, 0xa5, 0xa4, 0xe5, 0x88, 0x84, 0x09, 0x92, 0x53, 0x45, 0xbc,
	0x74, 0x28, 0x36, 0x4d, 0x05, 0x42, 0xdd, 0x1a, 0x25, 0x74, 0xa1, 0x52, 0x95, 0x6c, 0x2c, 0xd6,
	0xc5, 0x7e, 0xe4, 0x58, 0xea, 0xf9, 0x8c, 0xef, 0x5a, 0x99, 0x7e, 0x0a, 0x3e, 0x02, 0x3b, 0x2b,
	0x1f, 0xa2, 0x63, 0xc5, 0xc4, 0x54, 0xa1, 0x64, 0x61, 0xe7, 0x0b, 0xa0, 0x7b, 0xce, 0x66, 0x68,
	0xba, 0xd9, 0xcf, 0xef, 0xf7, 0xbf, 0xe7, 0x5e, 0x1e, 0xe2, 0x2e, 0x58, 0x1a, 0xf9, 0x4a, 0x4a,
	0xff, 0xf2, 0x60, 0x01, 0x8a, 0x1d, 0xf8, 0x31, 0xa4, 0x20, 0x13, 0xe9, 0x65, 0xb9, 0x50, 0x82,
	0xee, 0x68, 0xee, 0x29, 0x29, 0xbd, 0x92, 0xf7, 0x7b, 0xb1, 0x88, 0x05, 0x42, 0x5f, 0x7f, 0x19,
	0xaf, 0xff, 0x34, 0x14, 0x92, 0x0b, 0x19, 0x18, 0x60, 0x7e, 0x4a, 0xd4, 0xdf, 0x68, 0xa1, 0x97,
	0x43, 0x36, 0xfc, 0x6b, 0x91, 0xad, 0x13, 0xd3, 0x70, 0xae, 0x98, 0x02, 0xfa, 0x8e, 0x34, 0x33,
	0x96, 0x33, 0x2e, 0x1d, 0x6b, 0xd7, 0xda, 0xeb, 0x8c, 0x1c, 0xef, 0xee, 0x06, 0xbc, 0x33, 0xe4,
	0xe3, 0xc6, 0xf5, 0xed, 0xa0, 0x36, 0x2b, 0x6d, 0xfa, 0x96, 0x34, 0xe3, 0x5c, 0x5c, 0x64, 0xd2,
	0xb1, 0x77, 0xeb, 0x7b, 0x9d, 0xd1, 0x93, 0xcd, 0xdc, 0x89, 0xe6, 0x55, 0xcc, 0xc8, 0xf4, 0x3d,
	0x69, 0x71, 0xe0, 0x0b, 0xc8, 0xa5, 0x53, 0xc7, 0xdc, 0x3d, 0xfd, 0x4e, 0x51, 0x28, 0x83, 0x95,
	0x4e, 0x8f, 0x48, 0x3d, 0x02, 0xe9, 0x34, 0x30, 0xf5, 0x6c, 0x33, 0x35, 0x99, 0x96, 0xe7, 0x1a,
	0x77, 0x74, 0x70, 0x75, 0x3b, 0xa8, 0x4f, 0xa6, 0x72, 0xa6, 0x43, 0xc3, 0xef, 0x36, 0x69, 0x9a,
	0x53, 0xd0, 0xe7, 0xa4, 0xcb, 0x59, 0x11, 0xe0, 0x76, 0x02, 0x99, 0x5c, 0x01, 0x9e, 0xbb, 0x31,
	0xdb, 0xe2, 0xac, 0xc0, 0x0d, 0xcf, 0x93, 0x2b, 0xa0, 0x03, 0xa2, 0xff, 0x83, 0x28, 0x00, 0xe3,
	0xd8, 0xe8, 0xb4, 0x39, 0x2b, 0x26, 0x53, 0x14, 0x5e, 0x91, 0xed, 0x30, 0x07, 0xa6, 0x12, 0x91,
	0x06, 0x19, 0xe4, 0x89, 0x88, 0x9c, 0x3a, 0x3a, 0xdd, 0xaa, 0x7c, 0x86, 0x55, 0xfa, 0x82, 0x74,
	0x65, 0x12, 0xa7, 0x49, 0x1a, 0x57, 0x5e, 0x03, 0xbd, 0x87, 0x65, 0xb5, 0xd4, 0x3c, 0xf2, 0x58,
	0x37, 0xac, 0x54, 0xa6, 0x14, 0xf0, 0x4c, 0x39, 0x0f, 0xd0, 0x7d, 0xc4, 0x59, 0x31, 0x37, 0xe4,
	0xd8, 0x00, 0xfa, 0x92, 0x6c, 0x6b, 0x9f, 0x03, 0x17, 0xc1, 0x39, 0xa4, 0xb1, 0x5a, 0x3a, 0x4d,
	0xb3, 0x2e, 0x67, 0xc5, 0x29, 0x70, 0xf1, 0x11, 0x8b, 0x74, 0x9f, 0x50, 0xe3, 0x49, 0xc9, 0x62,
	0xa8, 0xd4, 0x16, 0xaa, 0x3b, 0xa8, 0x22, 0x30, 0xf6, 0x51, 0xe3, 0xcf, 0xb7, 0x81, 0x35, 0xfc,
	0x4c, 0xda, 0xff, 0x2f, 0x93, 0x8e, 0x48, 0x8b, 0x45, 0x51, 0x0e, 0xd2, 0x0c, 0x48, 0x7b, 0xec,
	0xfc, 0xfc, 0xf1, 0xba, 0x57, 0xce, 0xdb, 0xb1, 0x21, 0x73, 0x95, 0x27, 0x69, 0x3c, 0xab, 0x44,
	0xfa, 0x86, 0xd8, 0x91, 0xb9, 0xb3, 0xce, 0xa8, 0x77, 0xdf, 0x4b, 0x8d, 0x49, 0xf9, 0x44, 0xf6,
	0x64, 0x3a, 0xb3, 0x23, 0x18, 0x7f, 0xb8, 0x5e, 0xb9, 0xd6, 0xcd, 0xca, 0xb5, 0x7e, 0xaf, 0x5c,
	0xeb, 0xeb, 0xda, 0xad, 0xdd, 0xac, 0xdd, 0xda, 0xaf, 0xb5, 0x5b, 0xfb, 0xb4, 0x1f, 0x27, 0x6a,
	0x79, 0xb1, 0xf0, 0x42, 0xc1, 0x7d, 0xbd, 0x12, 0x8e, 0x71, 0x28, 0xce, 0xfd, 0x70, 0xc9, 0x92,
	0xd4, 0xbf, 0x3c, 0xf4, 0x0b, 0x1c, 0x75, 0xf5, 0x25, 0x03, 0xb9, 0x68, 0x22, 0x3e, 0xfc, 0x17,
	0x00, 0x00, 0xff, 0xff, 0xa2, 0xf2, 0x37, 0x6c, 0x66, 0x03, 0x00, 0x00,
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
	if this.MaxGroupSize != that1.MaxGroupSize {
		return false
	}
	if this.MaxDESize != that1.MaxDESize {
		return false
	}
	if this.CreationPeriod != that1.CreationPeriod {
		return false
	}
	if this.SigningPeriod != that1.SigningPeriod {
		return false
	}
	if this.MaxSigningAttempt != that1.MaxSigningAttempt {
		return false
	}
	if this.MaxMemoLength != that1.MaxMemoLength {
		return false
	}
	if this.MaxMessageLength != that1.MaxMessageLength {
		return false
	}
	return true
}
func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DEs) > 0 {
		for iNdEx := len(m.DEs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DEs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Members) > 0 {
		for iNdEx := len(m.Members) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Members[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Groups) > 0 {
		for iNdEx := len(m.Groups) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Groups[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
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
	if m.MaxMessageLength != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxMessageLength))
		i--
		dAtA[i] = 0x38
	}
	if m.MaxMemoLength != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxMemoLength))
		i--
		dAtA[i] = 0x30
	}
	if m.MaxSigningAttempt != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxSigningAttempt))
		i--
		dAtA[i] = 0x28
	}
	if m.SigningPeriod != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.SigningPeriod))
		i--
		dAtA[i] = 0x20
	}
	if m.CreationPeriod != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CreationPeriod))
		i--
		dAtA[i] = 0x18
	}
	if m.MaxDESize != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxDESize))
		i--
		dAtA[i] = 0x10
	}
	if m.MaxGroupSize != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxGroupSize))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *DEGenesis) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DEGenesis) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DEGenesis) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.DE.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.Groups) > 0 {
		for _, e := range m.Groups {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Members) > 0 {
		for _, e := range m.Members {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DEs) > 0 {
		for _, e := range m.DEs {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MaxGroupSize != 0 {
		n += 1 + sovGenesis(uint64(m.MaxGroupSize))
	}
	if m.MaxDESize != 0 {
		n += 1 + sovGenesis(uint64(m.MaxDESize))
	}
	if m.CreationPeriod != 0 {
		n += 1 + sovGenesis(uint64(m.CreationPeriod))
	}
	if m.SigningPeriod != 0 {
		n += 1 + sovGenesis(uint64(m.SigningPeriod))
	}
	if m.MaxSigningAttempt != 0 {
		n += 1 + sovGenesis(uint64(m.MaxSigningAttempt))
	}
	if m.MaxMemoLength != 0 {
		n += 1 + sovGenesis(uint64(m.MaxMemoLength))
	}
	if m.MaxMessageLength != 0 {
		n += 1 + sovGenesis(uint64(m.MaxMessageLength))
	}
	return n
}

func (m *DEGenesis) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = m.DE.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Groups", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Groups = append(m.Groups, Group{})
			if err := m.Groups[len(m.Groups)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Members", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Members = append(m.Members, Member{})
			if err := m.Members[len(m.Members)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DEs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DEs = append(m.DEs, DEGenesis{})
			if err := m.DEs[len(m.DEs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
				return fmt.Errorf("proto: wrong wireType = %d for field MaxGroupSize", wireType)
			}
			m.MaxGroupSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxGroupSize |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxDESize", wireType)
			}
			m.MaxDESize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxDESize |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreationPeriod", wireType)
			}
			m.CreationPeriod = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreationPeriod |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigningPeriod", wireType)
			}
			m.SigningPeriod = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SigningPeriod |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxSigningAttempt", wireType)
			}
			m.MaxSigningAttempt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxSigningAttempt |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxMemoLength", wireType)
			}
			m.MaxMemoLength = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxMemoLength |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxMessageLength", wireType)
			}
			m.MaxMessageLength = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxMessageLength |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *DEGenesis) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: DEGenesis: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DEGenesis: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DE", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DE.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
