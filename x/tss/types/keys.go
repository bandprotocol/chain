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

	// Round1DataStoreKeyPrefix is the key that keeps the member commitment on round 1.
	Round1DataStoreKeyPrefix = []byte{0x04}

	// Round1DataCountStoreKeyPrefix is the key that keeps the member commitments count on round 1.
	Round1DataCountStoreKeyPrefix = []byte{0x05}

	// DKGMaliciousIndexesStoreKeyPrefix is a list of indexes of malicious members.
	DKGMaliciousIndexesStoreKeyPrefix = []byte{0x06}

	// Round1CommitmentsCountStoreKeyPrefix is the key that keeps the member commitments count on round 1.
	Round1CommitmentsCountStoreKeyPrefix = []byte{0x07}

	// Round2ShareCountStoreKeyPrefix is the key that keeps the round 2 share count.
	Round2ShareCountStoreKeyPrefix = []byte{0x08}

	// ConfirmationsStoreKeyPrefix is a list of hash PubKey, schnorr signature on the PubKey and context.
	ConfirmationsStoreKeyPrefix = []byte{0x09}

	// Round2DataStoreKeyPrefix is the key that keeps the round2Data of the member.
	Round2DataStoreKeyPrefix = []byte{0x10}

	// Round2DataCountStoreKeyPrefix is the key that keeps the round2Data count.
	Round2DataCountStoreKeyPrefix = []byte{0x11}

	// Round3NoteStoreKeyPrefix is list for keep track of the progress of round3.
	Round3NoteStoreKeyPrefix = []byte{0x12}
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

func Round1DataStoreKey(groupID tss.GroupID) []byte {
	return append(Round1DataStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1DataCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round1DataCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1DataMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round1DataStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round2DataStoreKey(groupID tss.GroupID) []byte {
	return append(Round2DataStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round2DataMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round2DataStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round2DataCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round2DataCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func DKGMaliciousIndexesStoreKey(groupID tss.GroupID) []byte {
	return append(DKGContextStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func ConfirmationsStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmationsStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round3NoteStoreKey(groupID tss.GroupID) []byte {
	return append(Round3NoteStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}
