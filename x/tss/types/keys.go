package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/bandprotocol/chain/v2/pkg/tss"
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
	// RollingSeedSizeInBytes is the size of rolling block hash for random seed.
	RollingSeedSizeInBytes = 32

	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// GroupCountStoreKey is the key that keeps the total group count.
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)

	// SigningCountStoreKey is the key that keeps the total signing count.
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)

	// RollingSeedStoreKey is the key that keeps the seed based on the first 8-bit of the most recent 32 block hashes.
	RollingSeedStoreKey = append(GlobalStoreKeyPrefix, []byte("RollingSeed")...)

	// GroupStoreKeyPrefix is the prefix for group store.
	GroupStoreKeyPrefix = []byte{0x01}

	// DKGContextStoreKeyPrefix is the prefix for dkg context store.
	DKGContextStoreKeyPrefix = []byte{0x02}

	// MemberStoreKeyPrefix is the prefix for member store.
	MemberStoreKeyPrefix = []byte{0x03}

	// Round1DataStoreKeyPrefix is the key that keeps the round 1 data.
	Round1DataStoreKeyPrefix = []byte{0x04}

	// Round1DataCountStoreKeyPrefix is the key that keeps the round 1 data count.
	Round1DataCountStoreKeyPrefix = []byte{0x05}

	// AccumulatedCommitStoreKeyPrefix is the key that keeps total of each commit
	AccumulatedCommitStoreKeyPrefix = []byte{0x06}

	// Round2DataStoreKeyPrefix is the key that keeps the round 2 data of the member.
	Round2DataStoreKeyPrefix = []byte{0x07}

	// Round2DataCountStoreKeyPrefix is the key that keeps the round 2 data count.
	Round2DataCountStoreKeyPrefix = []byte{0x08}

	// ComplainWithStatusStoreKeyPrefix is the key that keeps complain with status.
	ComplainsWithStatusStoreKeyPrefix = []byte{0x09}

	// ConfirmComplainCountStoreKeyPrefix is the key for keep track of the progress of round 3.
	ConfirmComplainCountStoreKeyPrefix = []byte{0x10}

	// ConfirmStoreKeyPrefix is the key that keeps confirm.
	ConfirmStoreKeyPrefix = []byte{0x11}

	// DEStoreKeyPrefix is the key for keeps pre-commit DE.
	DEStoreKeyPrefix = []byte{0x12}

	// DEQueueStoreKeyPrefix is the key for keeps first and last index of the DEQueue.
	DEQueueStoreKeyPrefix = []byte{0x13}

	// SigningStoreKeyPrefix is the key for keeps signing data.
	SigningStoreKeyPrefix = []byte{0x14}

	// PendingSignStorKeyPrefix is the key for keeps pending signs data.
	PendingSignsStorKeyPrefix = []byte{0x15}

	// ZCountStoreKeyKeyPrefix is the key for keeps signing count data.
	ZCountStoreKeyKeyPrefix = []byte{0x16}

	// PartialZStoreKeyPrefix is the key for keeps partial z.
	PartialZStoreKeyPrefix = []byte{0x17}
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

func AccumulatedCommitStoreKey(groupID tss.GroupID) []byte {
	return append(AccumulatedCommitStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func AccumulatedCommitIndexStoreKey(groupID tss.GroupID, index uint64) []byte {
	return append(AccumulatedCommitStoreKey(groupID), sdk.Uint64ToBigEndian(index)...)
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

func ConfirmStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func ConfirmMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(ConfirmStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func ComplainsWithStatusStoreKey(groupID tss.GroupID) []byte {
	return append(ComplainsWithStatusStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func ComplainsWithStatusMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(ComplainsWithStatusStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func ConfirmComplainCountStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmComplainCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func DEStoreKey(address sdk.AccAddress) []byte {
	return append(DEStoreKeyPrefix, address...)
}

func DEIndexStoreKey(address sdk.AccAddress, index uint64) []byte {
	return append(DEStoreKey(address), sdk.Uint64ToBigEndian(index)...)
}

func DEQueueKeyStoreKey(address sdk.AccAddress) []byte {
	return append(DEQueueStoreKeyPrefix, address...)
}

func SigningStoreKey(signingID tss.SigningID) []byte {
	return append(SigningStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func PendingSignsStorKey(address sdk.AccAddress) []byte {
	return append(PendingSignsStorKeyPrefix, address...)
}

func PendingSignStorKey(address sdk.AccAddress, signingID tss.SigningID) []byte {
	return append(PendingSignsStorKey(address), sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func ZCountStoreKey(signingID tss.SigningID) []byte {
	return append(ZCountStoreKeyKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func PartialZStoreKey(signingID tss.SigningID) []byte {
	return append(PartialZStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func PartialZIndexStoreKey(signingID tss.SigningID, memberID tss.MemberID) []byte {
	return append(PartialZStoreKey(signingID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func SigningIDFromPendingSignKey(key []byte) uint64 {
	kv.AssertKeyLength(key, 29)
	return sdk.BigEndianToUint64(key[21:])
}
