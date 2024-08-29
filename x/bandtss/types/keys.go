package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

const (
	// module name
	ModuleName = "bandtss"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// RouterKey is the message route for the bandtss module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the bandtss module
	QuerierRoute = ModuleName
)

var (
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}
	// ParamsKeyPrefix is a prefix for keys that store bandtss's parameters
	ParamsKeyPrefix = []byte{0x01}
	// MemberStoreKeyPrefix is the prefix for member store.
	MemberStoreKeyPrefix = []byte{0x02}
	// SigningInfoStoreKeyPrefix is the prefix for SigningInfoStoreKey.
	SigningInfoStoreKeyPrefix = []byte{0x03}
	// SigningIDMappingStoreKeyPrefix is the prefix for SigningIDMappingStoreKey.
	SigningIDMappingStoreKeyPrefix = []byte{0x04}

	// SigningCountStoreKey is the key that keeps the total number of Signing.
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)
	// CurrentGroupIDStoreKey is the key for storing the current group ID.
	CurrentGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("CurrentGroupID")...)
	// GroupTransitionStoreKey is the key for storing the group transition information.
	GroupTransitionStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupTransition")...)
)

// MemberStoreKey returns the key for storing the member information.
func MemberStoreKey(address sdk.AccAddress, groupID tss.GroupID) []byte {
	return bytes.Join(
		[][]byte{
			MemberStoreKeyPrefix,
			sdk.Uint64ToBigEndian(uint64(groupID)),
			sdkaddress.MustLengthPrefix(address),
		}, []byte(""),
	)
}

// SigningInfoStoreKey returns the key for storing the bandtss signing info.
func SigningInfoStoreKey(id SigningID) []byte {
	return append(SigningInfoStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(id))...)
}

// SigningIDMappingStoreKey returns the key for storing the tss signing ID mapping.
func SigningIDMappingStoreKey(id tss.SigningID) []byte {
	return append(SigningIDMappingStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(id))...)
}
