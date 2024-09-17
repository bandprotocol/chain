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
	// global store keys
	GroupCountStoreKey           = []byte{0x00}
	SigningCountStoreKey         = []byte{0x01}
	PendingProcessGroupsStoreKey = []byte{0x02} // for processing group at endblock
	PendingSigningsStoreKey      = []byte{0x03} // for signing aggregation at endblock
	LastExpiredGroupIDStoreKey   = []byte{0x04} // for processing group expiration at endblock
	SigningExpirationsStoreKey   = []byte{0x05} // for process signing expiration at endblock

	// store prefixes for group, member
	GroupStoreKeyPrefix  = []byte{0x10}
	MemberStoreKeyPrefix = []byte{0x11}

	// store prefixes for group creation
	DKGContextStoreKeyPrefix           = []byte{0x12}
	Round1InfoStoreKeyPrefix           = []byte{0x13}
	Round1InfoCountStoreKeyPrefix      = []byte{0x14}
	AccumulatedCommitStoreKeyPrefix    = []byte{0x15}
	Round2InfoStoreKeyPrefix           = []byte{0x16}
	Round2InfoCountStoreKeyPrefix      = []byte{0x17}
	ComplainsWithStatusStoreKeyPrefix  = []byte{0x18}
	ConfirmStoreKeyPrefix              = []byte{0x19}
	ConfirmComplainCountStoreKeyPrefix = []byte{0x1a}

	// store prefixes for DE
	DEStoreKeyPrefix      = []byte{0x1b}
	DEQueueStoreKeyPrefix = []byte{0x1c}

	// store prefixes for signing
	SigningStoreKeyPrefix               = []byte{0x1d}
	PartialSignatureCountStoreKeyPrefix = []byte{0x1e}
	PartialSignatureStoreKeyPrefix      = []byte{0x1f}
	SigningAttemptStoreKeyPrefix        = []byte{0x20}

	// param store key
	ParamsKey = []byte{0x90}
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
