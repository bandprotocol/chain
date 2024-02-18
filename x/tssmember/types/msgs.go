package types

import (
	"fmt"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

var (
	_ sdk.Msg = &MsgCreateGroup{}
	_ sdk.Msg = &MsgReplaceGroup{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgCreateGroup creates a new MsgCreateGroup instance.
func NewMsgCreateGroup(members []string, threshold uint64, fee sdk.Coins, authority string) *MsgCreateGroup {
	return &MsgCreateGroup{
		Members:   members,
		Threshold: threshold,
		Fee:       fee,
		Authority: authority,
	}
}

// Type returns message type name.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	// Validate members address
	for _, member := range m.Members {
		_, err := sdk.AccAddressFromBech32(member)
		if err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("member: %s", member),
			)
		}
	}

	// Check duplicate member
	if DuplicateInArray(m.Members) {
		return errors.Wrap(fmt.Errorf("members can not duplicate"), "members")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("sender: %s", m.Authority),
		)
	}

	// Validate threshold must be less than or equal to members but more than zero
	if m.Threshold > uint64(len(m.Members)) || m.Threshold <= 0 {
		return errors.Wrap(
			fmt.Errorf("threshold must be less than or equal to the members but more than zero"),
			"threshold",
		)
	}

	// Validate fee
	if !m.Fee.IsValid() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, m.Fee.String())
	}

	return nil
}

// NewMsgReplaceGroup creates a new MsgReplaceGroup instance.
func NewMsgReplaceGroup(
	currentGroupID tss.GroupID,
	newGroupID tss.GroupID,
	execTime time.Time,
	authority string,
) *MsgReplaceGroup {
	return &MsgReplaceGroup{
		CurrentGroupID: currentGroupID,
		NewGroupID:     newGroupID,
		ExecTime:       execTime,
		Authority:      authority,
	}
}

// Type returns message type name.
func (m MsgReplaceGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgReplaceGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgReplaceGroup.
func (m MsgReplaceGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgReplaceGroup) ValidateBasic() error {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("sender: %s", m.Authority),
		)
	}

	return nil
}

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Type returns message type name.
func (m MsgUpdateParams) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
