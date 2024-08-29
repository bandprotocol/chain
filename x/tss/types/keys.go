package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
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
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// GroupCountStoreKey is the key that keeps the total group count.
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)

	// LastExpiredGroupIDStoreKey is the key for keeping last expired groupID.
	LastExpiredGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredGroupID")...)

	// SigningCountStoreKey is the key that keeps the total signing count.
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)

	// PendingProcessGroupsStoreKey is the key for storing pending process groups.
	PendingProcessGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessGroups")...)

	// PendingSigningsStoreKey is the key for storing pending process signings.
	PendingSigningsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessSignings")...)

	// SigningExpirationsStoreKey is the key for keeping signing expiration.
	SigningExpirationsStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningExpirations")...)

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

	// ComplainsWithStatusStoreKeyPrefix is the key that keeps complain with status.
	ComplainsWithStatusStoreKeyPrefix = []byte{0x09}

	// ConfirmComplainCountStoreKeyPrefix is the key for keep track of the progress of round 3.
	ConfirmComplainCountStoreKeyPrefix = []byte{0x0a}

	// ConfirmStoreKeyPrefix is the key that keeps confirm.
	ConfirmStoreKeyPrefix = []byte{0x0b}

	// DEStoreKeyPrefix is the key for keeping pre-commit DEs.
	DEStoreKeyPrefix = []byte{0x0c}

	// DEQueueStoreKeyPrefix is the prefix key for keeping the DE's queue information
	// of the specific address.
	DEQueueStoreKeyPrefix = []byte{0x0d}

	// SigningStoreKeyPrefix is the key for keeping signing data.
	SigningStoreKeyPrefix = []byte{0x0e}

	// PartialSignatureCountStoreKeyPrefix is the key for keeping signature count data.
	PartialSignatureCountStoreKeyPrefix = []byte{0x0f}

	// PartialSignatureStoreKeyPrefix is the key for keeping partial signature.
	PartialSignatureStoreKeyPrefix = []byte{0x10}

	// SigningAttemptStoreKeyPrefix is the key for keeping signing attempts.
	SigningAttemptStoreKeyPrefix = []byte{0x11}

	// ParamsKeyPrefix is a prefix for keys that store tss's parameters
	ParamsKeyPrefix = []byte{0x20}
)

// GroupStoreKey returns the key for storing group information.
func GroupStoreKey(groupID tss.GroupID) []byte {
	return append(GroupStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// DKGContextStoreKey returns the key for storing dkg context information.
func DKGContextStoreKey(groupID tss.GroupID) []byte {
	return append(DKGContextStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// MembersStoreKey returns the prefix of the MemberStoreKey for specific groupID.
func MembersStoreKey(groupID tss.GroupID) []byte {
	return append(MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// MemberStoreKey returns the key for storing member information.
func MemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(MembersStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// Round1InfoStoreKey returns the prefix for Round1InfoMemberStoreKey.
func Round1InfoStoreKey(groupID tss.GroupID) []byte {
	return append(Round1InfoStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// Round1InfoCountStoreKey returns the key for storing round-1 information count.
func Round1InfoCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round1InfoCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// Round1InfoMemberStoreKey returns the key for storing round-1 information of a given member.
func Round1InfoMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round1InfoStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// AccumulatedCommitStoreKey returns the prefix for AccumulatedCommitIndexStoreKey.
func AccumulatedCommitStoreKey(groupID tss.GroupID) []byte {
	return append(AccumulatedCommitStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// AccumulatedCommitIndexStoreKey returns the key for storing accumulated commit of a group.
func AccumulatedCommitIndexStoreKey(groupID tss.GroupID, index uint64) []byte {
	return append(AccumulatedCommitStoreKey(groupID), sdk.Uint64ToBigEndian(index)...)
}

// Round2InfoStoreKey returns the prefix for Round2InfoMemberStoreKey.
func Round2InfoStoreKey(groupID tss.GroupID) []byte {
	return append(Round2InfoStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// Round2InfoMemberStoreKey returns the key for storing round-2 information of a given member.
func Round2InfoMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(Round2InfoStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// Round2InfoCountStoreKey returns the key for storing round-2 information count.
func Round2InfoCountStoreKey(groupID tss.GroupID) []byte {
	return append(Round2InfoCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// ConfirmStoreKey returns the prefix for ConfirmMemberStoreKey.
func ConfirmStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// ConfirmMemberStoreKey returns the key for storing confirm information of a given member.
func ConfirmMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(ConfirmStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// ComplainsWithStatusStoreKey returns the prefix for ComplainsWithStatusMemberStoreKey.
func ComplainsWithStatusStoreKey(groupID tss.GroupID) []byte {
	return append(ComplainsWithStatusStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// ComplainsWithStatusMemberStoreKey returns the key for storing complain with status of a given member.
func ComplainsWithStatusMemberStoreKey(groupID tss.GroupID, memberID tss.MemberID) []byte {
	return append(ComplainsWithStatusStoreKey(groupID), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// ConfirmComplainCountStoreKey returns the key for storing confirm complain count.
func ConfirmComplainCountStoreKey(groupID tss.GroupID) []byte {
	return append(ConfirmComplainCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// DEsStoreKey returns the prefix of the key for user's DE.
func DEsStoreKey(address sdk.AccAddress) []byte {
	return append(DEStoreKeyPrefix, sdkaddress.MustLengthPrefix(address)...)
}

// DEStoreKey returns the key for storing whether DE exists or not.
func DEStoreKey(address sdk.AccAddress, index uint64) []byte {
	return append(DEsStoreKey(address), sdk.Uint64ToBigEndian(index)...)
}

// DEQueueStoreKey returns the key for storing the queue information (head and tail index)
// of DE of specific address.
func DEQueueStoreKey(address sdk.AccAddress) []byte {
	return append(DEQueueStoreKeyPrefix, sdkaddress.MustLengthPrefix(address)...)
}

// SigningStoreKey returns the key for storing signing information.
func SigningStoreKey(signingID tss.SigningID) []byte {
	return append(SigningStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

// PartialSignatureCountsStoreKey returns the prefix key for PartialSignatureCount store key.
func PartialSignatureCountsStoreKey(signingID tss.SigningID) []byte {
	return append(PartialSignatureCountStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

// PartialSignatureCountStoreKey returns the key for storing signature count information.
func PartialSignatureCountStoreKey(signingID tss.SigningID, attempt uint64) []byte {
	return append(PartialSignatureCountsStoreKey(signingID), sdk.Uint64ToBigEndian(attempt)...)
}

// PartialSignaturesBySigningIDStoreKey returns the prefix for PartialSignaturesStoreKey.
func PartialSignaturesBySigningIDStoreKey(signingID tss.SigningID) []byte {
	return append(PartialSignatureStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

// PartialSignaturesStoreKey returns the prefix for PartialSignatureStoreKey.
func PartialSignaturesStoreKey(signingID tss.SigningID, attempt uint64) []byte {
	return append(PartialSignaturesBySigningIDStoreKey(signingID), sdk.Uint64ToBigEndian(attempt)...)
}

// PartialSignatureStoreKey returns the key for storing partial signature information of a given member.
func PartialSignatureStoreKey(signingID tss.SigningID, attempt uint64, memberID tss.MemberID) []byte {
	return append(PartialSignaturesStoreKey(signingID, attempt), sdk.Uint64ToBigEndian(uint64(memberID))...)
}

// SigningAttemptsStoreKey returns the prefix key for SigningAttemptStoreKey.
func SigningAttemptsStoreKey(signingID tss.SigningID) []byte {
	return append(SigningAttemptStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(signingID))...)
}

// SigningAttemptStoreKey returns the key for storing signingAttempt information.
func SigningAttemptStoreKey(signingID tss.SigningID, attempt uint64) []byte {
	return append(SigningAttemptsStoreKey(signingID), sdk.Uint64ToBigEndian(attempt)...)
}

// MemberIDFromPartialSignatureStoreKey returns the memberID that is retrieved from the key.
func MemberIDFromPartialSignatureStoreKey(key []byte) tss.MemberID {
	kv.AssertKeyLength(key, 1+3*uint64Len)
	return tss.MemberID(sdk.BigEndianToUint64(key[1+2*uint64Len:]))
}

// ExtractAddressFromDEQueueStoreKey returns address that is retrieved from the key.
func ExtractAddressFromDEQueueStoreKey(key []byte) sdk.AccAddress {
	// key is of format prefix || addrLen (1byte) || addrBytes
	address := sdk.AccAddress(key[2:])
	return address
}
