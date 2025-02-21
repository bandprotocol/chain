package types

import (
	"encoding/json"
	fmt "fmt"
	"strings"
	"unicode/utf8"

	"github.com/axelarnetwork/utils/slices"
	"golang.org/x/text/unicode/norm"

	errorsmod "cosmossdk.io/errors"
)

// Note: This code is derived from github.com/axelarnetwork/axelar-core/x/nexus/exported

// DefaultDelimiter represents the default delimiter used for the KV store keys
const DefaultDelimiter = "_"

type AxelarMessageType int

const (
	// AxelarMessageTypeUnrecognized means coin type is unrecognized by axelar
	AxelarMessageTypeUnrecognized AxelarMessageType = iota
	// AxelarMessageTypeGeneralMessage is a pure axelar message
	AxelarMessageTypeGeneralMessage
	// AxelarMessageTypeGeneralMessageWithToken is a general axelar message with token
	AxelarMessageTypeGeneralMessageWithToken
	// AxelarMessageTypeSendToken is a direct token transfer
	AxelarMessageTypeSendToken
)

// AxelarFee is used to pay relayer for executing cross chain message
type AxelarFee struct {
	Amount          string  `json:"amount"`
	Recipient       string  `json:"recipient"`
	RefundRecipient *string `json:"refund_recipient"`
}

// NewAxelarFee creates a new AxelarFee instance.
func NewAxelarFee(
	amount string,
	recipient string,
	refundRecipient *string,
) AxelarFee {
	return AxelarFee{
		Amount:          amount,
		Recipient:       recipient,
		RefundRecipient: refundRecipient,
	}
}

// WasmBytes is a wrapper around []byte that gets JSON marshalized as an array
// of numbers instead of base64-encoded string
type WasmBytes []byte

// MarshalJSON implements json.Marshaler
func (bz WasmBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(slices.Map(bz, func(b byte) uint16 { return uint16(b) }))
}

// UnmarshalJSON implements json.Unmarshaler
func (bz *WasmBytes) UnmarshalJSON(data []byte) error {
	var arr []uint16
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	*bz = slices.Map(arr, func(u uint16) byte { return byte(u) })

	return nil
}

// ChainNameLengthMax bounds the max chain name length
const ChainNameLengthMax = 20

// ChainName ensures a correctly formatted EVM chain name
type ChainName string

// Validate returns an error, if the chain name is empty or too long
func (c ChainName) Validate() error {
	if err := ValidateString(string(c)); err != nil {
		return errorsmod.Wrap(err, "invalid chain name")
	}

	if len(c) > ChainNameLengthMax {
		return fmt.Errorf("chain name length %d is greater than %d", len(c), ChainNameLengthMax)
	}

	return nil
}

func (c ChainName) String() string {
	return string(c)
}

// ValidateString checks if the given string is:
//
// 1. non-empty
// 2. entirely composed of utf8 runes
// 3. normalized as NFKC
// 4. does not contain any forbidden Unicode code points
func ValidateString(str string, forbidden ...string) error {
	var f string
	if len(forbidden) == 0 {
		f = DefaultDelimiter
	} else {
		f = strings.Join(forbidden, "")
	}

	return validateString(str, false, f)
}

func validateString(str string, canBeEmpty bool, forbidden string) error {
	if !canBeEmpty && len(str) == 0 {
		return fmt.Errorf("string is empty")
	}

	if !utf8.ValidString(str) {
		return fmt.Errorf("not an utf8 string")
	}

	if !norm.NFKC.IsNormalString(str) {
		return fmt.Errorf("wrong normalization")
	}

	if len(forbidden) == 0 {
		return nil
	}

	forbidden = norm.NFKC.String(forbidden)
	if strings.ContainsAny(str, forbidden) {
		return fmt.Errorf("string '%s' must not contain any of '%s'", str, forbidden)
	}

	return nil
}
