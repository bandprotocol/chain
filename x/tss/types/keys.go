package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName = "tss"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName
)

var (
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// GroupCountStoreKey is the key that keeps the total request count.
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)

	// GroupStoreKeyPrefix is the prefix for group store.
	GroupStoreKeyPrefix = []byte{0x01}

	// DKGContextStoreKeyPrefix is the prefix for dkg context store.
	DKGContextStoreKeyPrefix = []byte{0x02}

	// MemberStoreKeyPrefix is the prefix for member store.
	MemberStoreKeyPrefix = []byte{0x3}

	// Round1Commitments is the key that keeps the member commitments on round 1
	Round1CommitmentsStoreKeyPrefix = []byte{0x04}
)

func GroupStoreKey(groupID uint64) []byte {
	return append(GroupStoreKeyPrefix, sdk.Uint64ToBigEndian(groupID)...)
}

func DKGContextStoreKey(groupID uint64) []byte {
	return append(DKGContextStoreKeyPrefix, sdk.Uint64ToBigEndian(groupID)...)
}

func MembersStoreKey(groupID uint64) []byte {
	return append(MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func MemberOfGroupKey(groupID uint64, memberID uint64) []byte {
	buf := append(MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(groupID)...)
	buf = append(buf, sdk.Uint64ToBigEndian(memberID)...)
	return buf
}

func Round1CommitmentsStoreKey(groupID uint64) []byte {
	return append(Round1CommitmentsStoreKeyPrefix, sdk.Uint64ToBigEndian(groupID)...)
}

func Round1CommitmentsMemberStoreKey(groupID uint64, memberID uint64) []byte {
	buf := append(Round1CommitmentsStoreKeyPrefix, sdk.Uint64ToBigEndian(groupID)...)
	buf = append(buf, sdk.Uint64ToBigEndian(memberID)...)
	return buf
}
