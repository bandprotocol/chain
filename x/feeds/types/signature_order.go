package types

import tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"

// signature order types
const (
	SignatureOrderTypeFeeds = "feeds"
)

// Implements Content Interface
var _ tsstypes.Content = &FeedsSignatureOrder{}

// NewFeedSignatureOrder returns a new FeedSignatureOrder object
func NewFeedSignatureOrder(signalIDs []string, encoder Encoder) *FeedsSignatureOrder {
	return &FeedsSignatureOrder{signalIDs, encoder}
}

// OrderRoute returns the order router key
func (f *FeedsSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "feeds"
func (f *FeedsSignatureOrder) OrderType() string {
	return SignatureOrderTypeFeeds
}

// IsInternal returns false for FeedsSignatureOrder (allow user to submit this content type).
func (f *FeedsSignatureOrder) IsInternal() bool { return false }

// ValidateBasic validates the request's title and description of the request signature
func (f *FeedsSignatureOrder) ValidateBasic() error {
	if len(f.SignalIDs) == 0 {
		return ErrInvalidSignalIDs
	}

	// Map to track signal IDs for duplicate check
	signalIDs := make(map[string]struct{})

	for _, id := range f.SignalIDs {
		// Check for duplicate signal IDs
		if _, exists := signalIDs[id]; exists {
			return ErrDuplicateSignalID.Wrapf("duplicate signal ID found: %s", id)
		}
	}

	if _, ok := Encoder_name[int32(f.Encoder)]; !ok {
		return ErrInvalidEncoder.Wrapf("invalid encoder: %s", f.Encoder)
	}

	if f.Encoder == ENCODER_UNSPECIFIED {
		return ErrInvalidEncoder.Wrapf("encoder type must be specified")
	}

	return nil
}
