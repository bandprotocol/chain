package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName = "tss"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the tss module
	QuerierRoute = ModuleName
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
	MemberStoreKeyPrefix = []byte{0x03}

	// Round1Commitment is the key that keeps the member commitment on round 1.
	Round1CommitmentStoreKeyPrefix = []byte{0x04}

	// Round2ShareStoreKeyPrefix is the key that keeps the Round2Share of the member.
	Round2ShareStoreKeyPrefix = []byte{0x05}

	// Round1CommitmentsCountStoreKeyPrefix is the key that keeps the member commitments count on round 1.
	Round1CommitmentsCountStoreKeyPrefix = []byte{0x06}

	// Round2ShareCountStoreKeyPrefix is the key that keeps the round 2 share count.
	Round2ShareCountStoreKeyPrefix = []byte{0x07}
)

func GroupStoreKey(groupID tss.GroupID) []byte {
	return append(GroupStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func DKGContextStoreKey(groupID tss.GroupID) []byte {
	return append(DKGContextStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func MembersStoreKey(groupID tss.GroupID) []byte {
	return append(MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func MemberOfGroupKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(MembersStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round1CommitmentStoreKey(groupID tss.GroupID) []byte {
	return append(Round1CommitmentStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1CommitmentsCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round1CommitmentsCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1CommitmentMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round1CommitmentStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round2ShareStoreKey(groupID tss.GroupID) []byte {
	return append(Round2ShareStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round2ShareMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round2ShareStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round2ShareCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round2ShareCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}
