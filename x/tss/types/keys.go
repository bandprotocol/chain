package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

const (
	// module name
	ModuleName = "tss"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// RouterKey is the message route for the tss module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the tss module
	QuerierRoute = ModuleName
)

var (
	ReplaceGroupMsgPrefix byte = 0x00
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

var (
	// RollingSeedSizeInBytes is the size of rolling block hash for random seed.
	RollingSeedSizeInBytes = 32

	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// GroupCountStoreKey is the key that keeps the total group count.
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)

	// ReplacementCountStoreKey is the key that keeps the total replacement group count.
	ReplacementCountStoreKey = append(GlobalStoreKeyPrefix, []byte("ReplacementCount")...)

	// LastExpiredGroupIDStoreKey is the key for keeps last expired groupID.
	LastExpiredGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredGroupID")...)

	// SigningCountStoreKey is the key that keeps the total signing count.
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)

	// LastExpiredSigningIDStoreKey is the key for keeps last expired signingID.
	LastExpiredSigningIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredSigningID")...)

	// PendingProcessGroupsStoreKey is the key for storing pending process groups.
	PendingProcessGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessGroups")...)

	// PendingSigningsStoreKey is the key for storing pending process signings.
	PendingSigningsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessSignings")...)

	// PendingReplaceGroupsStoreKey is the key for storing pending replace groups.
	PendingReplaceGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingReplaceGroups")...)

	// GroupStoreKeyPrefix is the prefix for group store.
	GroupStoreKeyPrefix = []byte{0x01}

	// DKGContextStoreKeyPrefix is the prefix for dkg context store.
	DKGContextStoreKeyPrefix = []byte{0x02}

	// MemberStoreKeyPrefix is the prefix for member store.
	MemberStoreKeyPrefix = []byte{0x03}

	// Round1InfoStoreKeyPrefix is the key that keeps the round 1 data.
	Round1InfoStoreKeyPrefix = []byte{0x04}

	// Round1InfoCountStoreKeyPrefix is the key that keeps the round 1 data count.
	Round1InfoCountStoreKeyPrefix = []byte{0x05}

	// AccumulatedCommitStoreKeyPrefix is the key that keeps total of each commit
	AccumulatedCommitStoreKeyPrefix = []byte{0x06}

	// Round2InfoStoreKeyPrefix is the key that keeps the round 2 data of the member.
	Round2InfoStoreKeyPrefix = []byte{0x07}

	// Round2InfoCountStoreKeyPrefix is the key that keeps the round 2 data count.
	Round2InfoCountStoreKeyPrefix = []byte{0x08}

	// ComplainWithStatusStoreKeyPrefix is the key that keeps complain with status.
	ComplainsWithStatusStoreKeyPrefix = []byte{0x09}

	// ConfirmComplainCountStoreKeyPrefix is the key for keep track of the progress of round 3.
	ConfirmComplainCountStoreKeyPrefix = []byte{0x0a}

	// ConfirmStoreKeyPrefix is the key that keeps confirm.
	ConfirmStoreKeyPrefix = []byte{0x0b}

	// DEStoreKeyPrefix is the key for keeps pre-commit DE.
	DEStoreKeyPrefix = []byte{0x0c}

	// DEQueueStoreKeyPrefix is the key for keeps first and last index of the DEQueue.
	DEQueueStoreKeyPrefix = []byte{0x0d}

	// SigningStoreKeyPrefix is the key for keeps signing data.
	SigningStoreKeyPrefix = []byte{0x0e}

	// SigCountStoreKeyPrefix is the key for keeps signature count data.
	SigCountStoreKeyPrefix = []byte{0x0f}

	// PartialSignatureStoreKeyPrefix is the key for keeps partial signature.
	PartialSignatureStoreKeyPrefix = []byte{0x10}

	// StatusStoreKeyPrefix is the prefix for status store.
	StatusStoreKeyPrefix = []byte{0x11}

	// ParamsKeyPrefix is a prefix for keys that store TSS's parameters
	ParamsKeyPrefix = []byte{0x12}

	// ReplacementPrefix is the prefix for keeps replacement group data.
	ReplacementKeyPrefix = []byte{0x13}

	// ReplacementQueuePrefix is the prefix for keeps replacement group queue.
	ReplacementQueuePrefix = []byte{0x14}
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

func Round1InfoStoreKey(groupID tss.GroupID) []byte {
	return append(Round1InfoStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1InfoCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round1InfoCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round1InfoMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round1InfoStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func AccumulatedCommitStoreKey(groupID tss.GroupID) []byte {
	return append(AccumulatedCommitStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func AccumulatedCommitIndexStoreKey(groupID tss.GroupID, index uint64) []byte {
	return append(AccumulatedCommitStoreKey(groupID), sdk.Uint64ToBigEndian(index)...)
}

func Round2InfoStoreKey(groupID tss.GroupID) []byte {
	return append(Round2InfoStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

func Round2InfoMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round2InfoStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func Round2InfoCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round2InfoCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
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

func AddressAndIndexFromDEStoreKey(key []byte) (sdk.AccAddress, uint64) {
	return sdk.AccAddress(key[1 : len(key)-uint64Len]), sdk.BigEndianToUint64(key[len(key)-uint64Len:])
}

func DEQueueKeyStoreKey(address sdk.AccAddress) []byte {
	return append(DEQueueStoreKeyPrefix, address...)
}

func AddressFromDEQueueStoreKey(key []byte) sdk.AccAddress {
	return sdk.AccAddress(key[1:])
}

func SigningStoreKey(signingID tss.SigningID) []byte {
	return append(SigningStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func SigCountStoreKey(signingID tss.SigningID) []byte {
	return append(SigCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func PartialSignatureStoreKey(signingID tss.SigningID) []byte {
	return append(PartialSignatureStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

func PartialSignatureMemberStoreKey(signingID tss.SigningID, memberID tss.MemberID) []byte {
	return append(PartialSignatureStoreKey(signingID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

func MemberIDFromPartialSignatureMemberStoreKey(key []byte) tss.MemberID {
	kv.AssertKeyLength(key, 1+2*uint64Len)
	return tss.MemberID(sdk.BigEndianToUint64(key[1+uint64Len:]))
}

func SigningIDFromPendingSignStoreKey(key []byte) uint64 {
	return sdk.BigEndianToUint64(key[len(key)-uint64Len:])
}

func StatusStoreKey(address sdk.AccAddress) []byte {
	return append(StatusStoreKeyPrefix, address...)
}

func ReplacementKey(replacementID uint64) []byte {
	return append(ReplacementKeyPrefix, sdk.Uint64ToBigEndian(replacementID)...)
}

func ReplacementQueueByTimeKey(endTime time.Time) []byte {
	return append(ReplacementQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

func ReplacementQueueKey(replacementID uint64, endTime time.Time) []byte {
	return append(ReplacementQueueByTimeKey(endTime), sdk.Uint64ToBigEndian(uint64(replacementID))...)
}

func SplitReplacementQueueKey(key []byte) (replacementID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

func splitKeyWithTime(key []byte) (replacementID uint64, endTime time.Time) {
	kv.AssertKeyLength(key[1:], 8+lenTime)

	endTime, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		panic(err)
	}

	replacementID = sdk.BigEndianToUint64(key[1+lenTime:])
	return
}
