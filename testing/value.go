/*
	This file contains the variables, constants, and default values
	used in the testing package and commonly defined in tests.
*/
package ibctesting

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"

	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	FirstClientID     = "07-tendermint-0"
	FirstChannelID    = "channel-0"
	FirstConnectionID = "connection-0"

	// Default params constants used to create a TM client
	TrustingPeriod     time.Duration = time.Hour * 24 * 7 * 2
	UnbondingPeriod    time.Duration = time.Hour * 24 * 7 * 3
	MaxClockDrift      time.Duration = time.Second * 10
	DefaultDelayPeriod uint64        = 0

	DefaultChannelVersion = oracletypes.Version
	InvalidID             = "IDisInvalid"

	// Application Ports
	OraclePort = oracletypes.ModuleName
	MockPort   = mock.ModuleName

	// used for testing proposals
	Title       = "title"
	Description = "description"
)

var (
	DefaultOpenInitVersion *connectiontypes.Version

	// Default params variables used to create a TM client
	DefaultTrustLevel ibctmtypes.Fraction = ibctmtypes.DefaultTrustLevel
	TestCoin                              = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))

	UpgradePath = []string{"upgrade", "upgradedIBCState"}

	ConnectionVersion = connectiontypes.ExportedVersionsToProto(connectiontypes.GetCompatibleVersions())[0]

	// MockAcknowledgement      = mock.MockAcknowledgement.Acknowledgement()
	// MockPacketData           = []byte("mock packet data")
	// MockFailPacketData       = []byte("mock failed packet data")
	// MockCanaryCapabilityName = []byte("mock async packet data")

	prefix = commitmenttypes.NewMerklePrefix([]byte("ibc"))
)
