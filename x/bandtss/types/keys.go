package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bandprotocol/chain/v3/pkg/tss"
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
	// global store keys
	SigningCountStoreKey    = []byte{0x00}
	CurrentGroupStoreKey    = []byte{0x01}
	GroupTransitionStoreKey = []byte{0x02}

	// store prefixes
	MemberStoreKeyPrefix           = []byte{0x10}
	SigningInfoStoreKeyPrefix      = []byte{0x11}
	SigningIDMappingStoreKeyPrefix = []byte{0x12}

	// param store key
	ParamsKey = []byte{0x90}
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
