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

	// Round1Commitments is the key that keeps the member commitments on round 1.
	Round1CommitmentsStoreKeyPrefix = []byte{0x04}

	// Round2ShareStoreKeyPrefix is the key that keeps the member encrypted secret share on round 2.
	Round2ShareStoreKeyPrefix = []byte{0x05}

	// DKGMaliciousIndexesStoreKeyPrefix is a list of indexes of malicious members.
	DKGMaliciousIndexesStoreKeyPrefix = []byte{0x06}

	// ConfirmationsStoreKeyPrefix is a list of hash PubKey, schnorr signature on the PubKey and context.
	ConfirmationsStoreKeyPrefix = []byte{0x07}

	// PendingRoundNoteStoreKeyPrefix is list for keep track of the progress of the group status PENDING.
	PendingRoundNoteStoreKeyPrefix = []byte{0x08}
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
	buf := append(MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
	buf = append(buf, sdk.Uint64ToBigEndian(uint64(memberID))...)
	return buf
}

func Round1CommitmentsStoreKey(groupID tss.GroupID) []byte {
	return append(Round1CommitmentsStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1CommitmentsMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	buf := append(Round1CommitmentsStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
	buf = append(buf, sdk.Uint64ToBigEndian(uint64(memberID))...)
	return buf
}

func Round2ShareStoreKey(groupID tss.GroupID) []byte {
	return append(Round2ShareStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round2ShareMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	buf := append(Round2ShareStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
	buf = append(buf, sdk.Uint64ToBigEndian(uint64(memberID))...)
	return buf
}

func DKGMaliciousIndexesStoreKey(groupID tss.GroupID) []byte {
	return append(DKGContextStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func ConfirmationsStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmationsStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func PendingRoundNoteStoreKey(groupID tss.GroupID) []byte {
	return append(PendingRoundNoteStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}
