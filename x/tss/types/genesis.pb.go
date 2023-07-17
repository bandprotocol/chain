// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tss/v1beta1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/durationpb"
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
	// group_count defines the number of groups.
	GroupCount uint64 `protobuf:"varint,2,opt,name=group_count,json=groupCount,proto3" json:"group_count,omitempty"`
	// signing_count defines the number of signers.
	SigningCount uint64 `protobuf:"varint,3,opt,name=signing_count,json=signingCount,proto3" json:"signing_count,omitempty"`
	// groups is an array containing information about each group.
	Groups []Group `protobuf:"bytes,4,rep,name=groups,proto3" json:"groups"`
	// de_queues_genesis is an array containing the de queues of all the address.
	DEQueuesGenesis []DEQueueGenesis `protobuf:"bytes,5,rep,name=de_queues_genesis,json=deQueuesGenesis,proto3" json:"de_queues_genesis"`
	// des_genesis is an array containing the des of all the address.
	DEsGenesis []DEGenesis `protobuf:"bytes,6,rep,name=des_genesis,json=desGenesis,proto3" json:"des_genesis"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5f1c2be1950e47, []int{0}
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

func (m *GenesisState) GetGroupCount() uint64 {
	if m != nil {
		return m.GroupCount
	}
	return 0
}

func (m *GenesisState) GetSigningCount() uint64 {
	if m != nil {
		return m.SigningCount
	}
	return 0
}

func (m *GenesisState) GetGroups() []Group {
	if m != nil {
		return m.Groups
	}
	return nil
}

func (m *GenesisState) GetDEQueuesGenesis() []DEQueueGenesis {
	if m != nil {
		return m.DEQueuesGenesis
	}
	return nil
}

func (m *GenesisState) GetDEsGenesis() []DEGenesis {
	if m != nil {
		return m.DEsGenesis
	}
	return nil
}

// Params defines the set of module parameters.
type Params struct {
	// max_group_size is the maximum of the member capacity of the group.
	MaxGroupSize uint64 `protobuf:"varint,1,opt,name=max_group_size,json=maxGroupSize,proto3" json:"max_group_size,omitempty"`
	// max_d_e_size is the maximum of the de capacity of the member.
	MaxDESize uint64 `protobuf:"varint,2,opt,name=max_d_e_size,json=maxDESize,proto3" json:"max_d_e_size,omitempty"`
	// group_sig_creating_period is the number of blocks allowed to creating group signature.
	GroupSigCreatingPeriod int64 `protobuf:"varint,3,opt,name=group_sig_creating_period,json=groupSigCreatingPeriod,proto3" json:"group_sig_creating_period,omitempty"`
	// signing_period is the number of blocks allowed to signing.
	SigningPeriod int64 `protobuf:"varint,4,opt,name=signing_period,json=signingPeriod,proto3" json:"signing_period,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5f1c2be1950e47, []int{1}
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

func (m *Params) GetGroupSigCreatingPeriod() int64 {
	if m != nil {
		return m.GroupSigCreatingPeriod
	}
	return 0
}

func (m *Params) GetSigningPeriod() int64 {
	if m != nil {
		return m.SigningPeriod
	}
	return 0
}

// DEQueueGenesis defines an account address and de queue used in the tss module's genesis state.
type DEQueueGenesis struct {
	// address is the address of the de holder.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// de_queue defines the different de queue this balance holds.
	DEQueue *DEQueue `protobuf:"bytes,2,opt,name=de_queue,json=deQueue,proto3" json:"de_queue,omitempty"`
}

func (m *DEQueueGenesis) Reset()         { *m = DEQueueGenesis{} }
func (m *DEQueueGenesis) String() string { return proto.CompactTextString(m) }
func (*DEQueueGenesis) ProtoMessage()    {}
func (*DEQueueGenesis) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5f1c2be1950e47, []int{2}
}
func (m *DEQueueGenesis) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DEQueueGenesis) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DEQueueGenesis.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DEQueueGenesis) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DEQueueGenesis.Merge(m, src)
}
func (m *DEQueueGenesis) XXX_Size() int {
	return m.Size()
}
func (m *DEQueueGenesis) XXX_DiscardUnknown() {
	xxx_messageInfo_DEQueueGenesis.DiscardUnknown(m)
}

var xxx_messageInfo_DEQueueGenesis proto.InternalMessageInfo

func (m *DEQueueGenesis) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *DEQueueGenesis) GetDEQueue() *DEQueue {
	if m != nil {
		return m.DEQueue
	}
	return nil
}

// DEGenesis defines an account address and de pair used in the tss module's genesis state.
type DEGenesis struct {
	// address is the address of the de holder.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// index is the index for store de of the address
	Index uint64 `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
	// de defines the different de this balance holds.
	DE *DE `protobuf:"bytes,3,opt,name=de,proto3" json:"de,omitempty"`
}

func (m *DEGenesis) Reset()         { *m = DEGenesis{} }
func (m *DEGenesis) String() string { return proto.CompactTextString(m) }
func (*DEGenesis) ProtoMessage()    {}
func (*DEGenesis) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb5f1c2be1950e47, []int{3}
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

func (m *DEGenesis) GetIndex() uint64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *DEGenesis) GetDE() *DE {
	if m != nil {
		return m.DE
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "tss.v1beta1.GenesisState")
	proto.RegisterType((*Params)(nil), "tss.v1beta1.Params")
	proto.RegisterType((*DEQueueGenesis)(nil), "tss.v1beta1.DEQueueGenesis")
	proto.RegisterType((*DEGenesis)(nil), "tss.v1beta1.DEGenesis")
}

func init() { proto.RegisterFile("tss/v1beta1/genesis.proto", fileDescriptor_eb5f1c2be1950e47) }

var fileDescriptor_eb5f1c2be1950e47 = []byte{
	// 537 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0x41, 0x8f, 0xd2, 0x40,
	0x18, 0xa5, 0xc0, 0x16, 0xf9, 0x8a, 0x10, 0x47, 0x5c, 0xd9, 0x35, 0x69, 0x09, 0x6a, 0xe4, 0x60,
	0xa8, 0xe0, 0xc9, 0xc4, 0x13, 0x0b, 0xee, 0xc1, 0xcb, 0xda, 0xbd, 0x79, 0xa9, 0x03, 0x33, 0x0e,
	0x4d, 0x96, 0x0e, 0x76, 0xa6, 0x1b, 0xdc, 0x5f, 0xe1, 0x5f, 0xf1, 0x5f, 0xec, 0x71, 0x8f, 0x9e,
	0x88, 0x29, 0x17, 0x7f, 0x86, 0xe9, 0xcc, 0x94, 0xc0, 0xc6, 0x78, 0xe3, 0x7b, 0xef, 0xcd, 0x7b,
	0xcc, 0xfb, 0x3a, 0x70, 0x22, 0x85, 0xf0, 0xaf, 0x87, 0x33, 0x2a, 0xf1, 0xd0, 0x67, 0x34, 0xa6,
	0x22, 0x12, 0x83, 0x55, 0xc2, 0x25, 0x47, 0x8e, 0x14, 0x62, 0x60, 0xa8, 0xd3, 0x36, 0xe3, 0x8c,
	0x2b, 0xdc, 0xcf, 0x7f, 0x69, 0xc9, 0xa9, 0xcb, 0x38, 0x67, 0x57, 0xd4, 0x57, 0xd3, 0x2c, 0xfd,
	0xea, 0x93, 0x34, 0xc1, 0x32, 0xe2, 0xb1, 0xe1, 0x9f, 0xec, 0xbb, 0xe7, 0x76, 0x0a, 0xee, 0xfd,
	0x29, 0x43, 0xe3, 0x5c, 0x67, 0x5d, 0x4a, 0x2c, 0x29, 0x1a, 0x82, 0xbd, 0xc2, 0x09, 0x5e, 0x8a,
	0x8e, 0xd5, 0xb5, 0xfa, 0xce, 0xe8, 0xf1, 0x60, 0x2f, 0x7b, 0x70, 0xa1, 0xa8, 0x71, 0xf5, 0x76,
	0xe3, 0x95, 0x02, 0x23, 0x44, 0x1e, 0x38, 0x2c, 0xe1, 0xe9, 0x2a, 0x9c, 0xf3, 0x34, 0x96, 0x9d,
	0x72, 0xd7, 0xea, 0x57, 0x03, 0x50, 0xd0, 0x59, 0x8e, 0xa0, 0xe7, 0xf0, 0x50, 0x44, 0x2c, 0x8e,
	0x62, 0x66, 0x24, 0x15, 0x25, 0x69, 0x18, 0x50, 0x8b, 0xde, 0x80, 0xad, 0x8e, 0x88, 0x4e, 0xb5,
	0x5b, 0xe9, 0x3b, 0x23, 0x74, 0x10, 0x7c, 0x9e, 0x53, 0x45, 0xae, 0xd6, 0xa1, 0x2f, 0xf0, 0x88,
	0xd0, 0xf0, 0x5b, 0x4a, 0x53, 0x2a, 0x42, 0x53, 0x58, 0xe7, 0x48, 0x1d, 0x7e, 0x76, 0x70, 0x78,
	0x32, 0xfd, 0x94, 0x8b, 0xcc, 0x3d, 0xc7, 0x4f, 0x73, 0x97, 0x6c, 0xe3, 0xb5, 0x0c, 0x2e, 0x0c,
	0x11, 0xb4, 0x08, 0x3d, 0x00, 0xd0, 0x47, 0x70, 0xc8, 0x9e, 0xb7, 0xad, 0xbc, 0x8f, 0xef, 0x79,
	0x17, 0xb6, 0xc8, 0xd8, 0xc2, 0x64, 0xba, 0x73, 0x04, 0xb2, 0x33, 0xeb, 0xfd, 0xb4, 0xc0, 0xd6,
	0xfd, 0xa1, 0x17, 0xd0, 0x5c, 0xe2, 0x75, 0xa8, 0x5b, 0x13, 0xd1, 0x0d, 0x55, 0x65, 0x57, 0x83,
	0xc6, 0x12, 0xaf, 0xd5, 0x4d, 0x2f, 0xa3, 0x1b, 0x8a, 0x3c, 0xc8, 0xe7, 0x90, 0x84, 0x54, 0x6b,
	0x74, 0xb1, 0xf5, 0x25, 0x5e, 0x4f, 0xa6, 0x4a, 0xf0, 0x0e, 0x4e, 0x0a, 0x0b, 0x16, 0xce, 0x13,
	0x8a, 0x65, 0x5e, 0xf1, 0x8a, 0x26, 0x11, 0x27, 0xaa, 0xe3, 0x4a, 0x70, 0xcc, 0xb4, 0x1d, 0x3b,
	0x33, 0xf4, 0x85, 0x62, 0xd1, 0x4b, 0x68, 0x16, 0x2b, 0x31, 0xfa, 0xaa, 0xd2, 0x17, 0x8b, 0xd2,
	0xb2, 0xde, 0x02, 0x9a, 0x87, 0xe5, 0xa1, 0x0e, 0xd4, 0x30, 0x21, 0x09, 0x15, 0xfa, 0x03, 0xa9,
	0x07, 0xc5, 0x88, 0xde, 0xc3, 0x83, 0x62, 0x1d, 0xea, 0xaf, 0x3a, 0xa3, 0xf6, 0xbf, 0xb6, 0x30,
	0x76, 0xb2, 0x8d, 0x57, 0x33, 0x43, 0x50, 0x33, 0x95, 0xf7, 0x08, 0xd4, 0x77, 0x55, 0xfe, 0x27,
	0xa4, 0x0d, 0x47, 0x51, 0x4c, 0xe8, 0xda, 0x94, 0xa1, 0x07, 0xf4, 0x0a, 0xca, 0x84, 0xaa, 0x1b,
	0x3b, 0xa3, 0xd6, 0xbd, 0xd0, 0xb1, 0x9d, 0x6d, 0xbc, 0xf2, 0x64, 0x1a, 0x94, 0x09, 0x1d, 0x7f,
	0xb8, 0xcd, 0x5c, 0xeb, 0x2e, 0x73, 0xad, 0xdf, 0x99, 0x6b, 0xfd, 0xd8, 0xba, 0xa5, 0xbb, 0xad,
	0x5b, 0xfa, 0xb5, 0x75, 0x4b, 0x9f, 0x5f, 0xb3, 0x48, 0x2e, 0xd2, 0xd9, 0x60, 0xce, 0x97, 0xfe,
	0x0c, 0xc7, 0x44, 0x3d, 0x8f, 0x39, 0xbf, 0xf2, 0xe7, 0x0b, 0x1c, 0xc5, 0xfe, 0xf5, 0xc8, 0x5f,
	0xe7, 0xcf, 0xc6, 0x97, 0xdf, 0x57, 0x54, 0xcc, 0x6c, 0x45, 0xbf, 0xfd, 0x1b, 0x00, 0x00, 0xff,
	0xff, 0xeb, 0x9d, 0xa8, 0xe1, 0xb4, 0x03, 0x00, 0x00,
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
	if len(m.DEsGenesis) > 0 {
		for iNdEx := len(m.DEsGenesis) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DEsGenesis[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.DEQueuesGenesis) > 0 {
		for iNdEx := len(m.DEQueuesGenesis) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DEQueuesGenesis[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
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
			dAtA[i] = 0x22
		}
	}
	if m.SigningCount != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.SigningCount))
		i--
		dAtA[i] = 0x18
	}
	if m.GroupCount != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.GroupCount))
		i--
		dAtA[i] = 0x10
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
	if m.SigningPeriod != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.SigningPeriod))
		i--
		dAtA[i] = 0x20
	}
	if m.GroupSigCreatingPeriod != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.GroupSigCreatingPeriod))
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

func (m *DEQueueGenesis) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DEQueueGenesis) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DEQueueGenesis) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DEQueue != nil {
		{
			size, err := m.DEQueue.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
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
	if m.DE != nil {
		{
			size, err := m.DE.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if m.Index != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x10
	}
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
	if m.GroupCount != 0 {
		n += 1 + sovGenesis(uint64(m.GroupCount))
	}
	if m.SigningCount != 0 {
		n += 1 + sovGenesis(uint64(m.SigningCount))
	}
	if len(m.Groups) > 0 {
		for _, e := range m.Groups {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DEQueuesGenesis) > 0 {
		for _, e := range m.DEQueuesGenesis {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DEsGenesis) > 0 {
		for _, e := range m.DEsGenesis {
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
	if m.GroupSigCreatingPeriod != 0 {
		n += 1 + sovGenesis(uint64(m.GroupSigCreatingPeriod))
	}
	if m.SigningPeriod != 0 {
		n += 1 + sovGenesis(uint64(m.SigningPeriod))
	}
	return n
}

func (m *DEQueueGenesis) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.DEQueue != nil {
		l = m.DEQueue.Size()
		n += 1 + l + sovGenesis(uint64(l))
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
	if m.Index != 0 {
		n += 1 + sovGenesis(uint64(m.Index))
	}
	if m.DE != nil {
		l = m.DE.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupCount", wireType)
			}
			m.GroupCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GroupCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigningCount", wireType)
			}
			m.SigningCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SigningCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
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
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DEQueuesGenesis", wireType)
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
			m.DEQueuesGenesis = append(m.DEQueuesGenesis, DEQueueGenesis{})
			if err := m.DEQueuesGenesis[len(m.DEQueuesGenesis)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DEsGenesis", wireType)
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
			m.DEsGenesis = append(m.DEsGenesis, DEGenesis{})
			if err := m.DEsGenesis[len(m.DEsGenesis)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: wrong wireType = %d for field GroupSigCreatingPeriod", wireType)
			}
			m.GroupSigCreatingPeriod = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GroupSigCreatingPeriod |= int64(b&0x7F) << shift
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
				m.SigningPeriod |= int64(b&0x7F) << shift
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
func (m *DEQueueGenesis) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: DEQueueGenesis: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DEQueueGenesis: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field DEQueue", wireType)
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
			if m.DEQueue == nil {
				m.DEQueue = &DEQueue{}
			}
			if err := m.DEQueue.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
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
			if m.DE == nil {
				m.DE = &DE{}
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
