package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgSubmitSignalPrices)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgUpdateReferenceSourceConfig)(nil)
	_ sdk.Msg = (*MsgVote)(nil)

	_ sdk.HasValidateBasic = (*MsgSubmitSignalPrices)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateReferenceSourceConfig)(nil)
	_ sdk.HasValidateBasic = (*MsgVote)(nil)
)

// ====================================
// MsgSubmitSignalPrices
// ====================================

// NewMsgSubmitSignalPrices creates a new MsgSubmitSignalPrices instance.
func NewMsgSubmitSignalPrices(
	validator string,
	timestamp int64,
	signalPrices []SignalPrice,
) *MsgSubmitSignalPrices {
	return &MsgSubmitSignalPrices{
		Validator:    validator,
		Timestamp:    timestamp,
		SignalPrices: signalPrices,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgSubmitSignalPrices) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(m.Validator); err != nil {
		return err
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, signalPrice := range m.SignalPrices {
		if signalPrice.Status != SignalPriceStatusAvailable && signalPrice.Price != 0 {
			return sdkerrors.ErrInvalidRequest.Wrap(
				"signal price must be initial value if price status is unsupported or unavailable",
			)
		}

		// Check for duplicate signal IDs
		if _, exists := signalIDSet[signalPrice.SignalID]; exists {
			return ErrDuplicateSignalID.Wrapf(
				"duplicate signal ID found: %s", signalPrice.SignalID,
			)
		}
		signalIDSet[signalPrice.SignalID] = struct{}{}
	}

	return nil
}

// ====================================
// MsgUpdateParams
// ====================================

// NewMsgUpdateParams creates a new MsgUpdateParams instance.
func NewMsgUpdateParams(
	authority string,
	params Params,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgUpdateReferenceSourceConfig
// ====================================

// NewMsgUpdateReferenceSourceConfig creates a new MsgUpdateReferenceSourceConfig instance.
func NewMsgUpdateReferenceSourceConfig(
	admin string,
	referenceSourceConfig ReferenceSourceConfig,
) *MsgUpdateReferenceSourceConfig {
	return &MsgUpdateReferenceSourceConfig{
		Admin:                 admin,
		ReferenceSourceConfig: referenceSourceConfig,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdateReferenceSourceConfig) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrap(err, "invalid admin address")
	}

	if err := m.ReferenceSourceConfig.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgVote
// ====================================

// NewMsgVote creates a new MsgVote instance.
func NewMsgVote(
	voter string,
	signals []Signal,
) *MsgVote {
	return &MsgVote{
		Voter:   voter,
		Signals: signals,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgVote) ValidateBasic() error {
	// Check if the voter address is valid
	if _, err := sdk.AccAddressFromBech32(m.Voter); err != nil {
		return errorsmod.Wrap(err, "invalid voter address")
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, signal := range m.Signals {
		// Validate Signal
		if err := signal.Validate(); err != nil {
			return err
		}

		// Check for duplicate signal IDs
		if _, exists := signalIDSet[signal.ID]; exists {
			return ErrDuplicateSignalID.Wrapf(
				"duplicate signal ID found: %s", signal.ID,
			)
		}
		signalIDSet[signal.ID] = struct{}{}
	}

	return nil
}
