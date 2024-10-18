package types

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgRequestData)(nil)
	_ sdk.Msg = (*MsgReportData)(nil)
	_ sdk.Msg = (*MsgCreateDataSource)(nil)
	_ sdk.Msg = (*MsgEditDataSource)(nil)
	_ sdk.Msg = (*MsgCreateOracleScript)(nil)
	_ sdk.Msg = (*MsgEditOracleScript)(nil)
	_ sdk.Msg = (*MsgActivate)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)

	_ sdk.HasValidateBasic = (*MsgRequestData)(nil)
	_ sdk.HasValidateBasic = (*MsgReportData)(nil)
	_ sdk.HasValidateBasic = (*MsgCreateDataSource)(nil)
	_ sdk.HasValidateBasic = (*MsgEditDataSource)(nil)
	_ sdk.HasValidateBasic = (*MsgCreateOracleScript)(nil)
	_ sdk.HasValidateBasic = (*MsgEditOracleScript)(nil)
	_ sdk.HasValidateBasic = (*MsgActivate)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)
)

// NewMsgRequestData creates a new MsgRequestData instance.
func NewMsgRequestData(
	oracleScriptID OracleScriptID,
	calldata []byte,
	askCount, minCount uint64,
	clientID string,
	feeLimit sdk.Coins,
	prepareGas, executeGas uint64,
	sender sdk.AccAddress,
) *MsgRequestData {
	return &MsgRequestData{
		OracleScriptID: oracleScriptID,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		ClientID:       clientID,
		FeeLimit:       feeLimit,
		Sender:         sender.String(),
		PrepareGas:     prepareGas,
		ExecuteGas:     executeGas,
	}
}

// ValidateBasic checks whether the given MsgRequestData instance (sdk.Msg interface).
func (m MsgRequestData) ValidateBasic() error {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", m.Sender)
	}
	if m.MinCount <= 0 {
		return ErrInvalidMinCount.Wrapf("got: %d", m.MinCount)
	}
	if m.AskCount < m.MinCount {
		return ErrInvalidAskCount.Wrapf("got: %d, min count: %d", m.AskCount, m.MinCount)
	}
	if len(m.ClientID) > MaxClientIDLength {
		return WrapMaxError(ErrTooLongClientID, len(m.ClientID), MaxClientIDLength)
	}
	if m.PrepareGas <= 0 {
		return ErrInvalidOwasmGas.Wrapf("invalid prepare gas: %d", m.PrepareGas)
	}
	if m.ExecuteGas <= 0 {
		return ErrInvalidOwasmGas.Wrapf("invalid execute gas: %d", m.ExecuteGas)
	}
	if m.PrepareGas+m.ExecuteGas > MaximumOwasmGas {
		return ErrInvalidOwasmGas.Wrapf(
			"sum of prepare gas and execute gas (%d) exceed %d",
			m.PrepareGas+m.ExecuteGas,
			MaximumOwasmGas,
		)
	}
	if !m.FeeLimit.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.FeeLimit.String())
	}
	return nil
}

// NewMsgReportData creates a new MsgReportData instance
func NewMsgReportData(requestID RequestID, rawReports []RawReport, validator sdk.ValAddress) *MsgReportData {
	return &MsgReportData{
		RequestID:  requestID,
		RawReports: rawReports,
		Validator:  validator.String(),
	}
}

// ValidateBasic checks whether the given MsgReportData instance (sdk.Msg interface).
func (m MsgReportData) ValidateBasic() error {
	valAddr, err := sdk.ValAddressFromBech32(m.Validator)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(valAddr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("validator: %s", m.Validator)
	}
	if len(m.RawReports) == 0 {
		return ErrEmptyReport
	}
	uniqueMap := make(map[ExternalID]bool)
	for _, r := range m.RawReports {
		if _, found := uniqueMap[r.ExternalID]; found {
			return ErrDuplicateExternalID.Wrapf("external id: %d", r.ExternalID)
		}
		uniqueMap[r.ExternalID] = true
	}
	return nil
}

// NewMsgCreateDataSource creates a new MsgCreateDataSource instance
func NewMsgCreateDataSource(
	name, description string, executable []byte, fee sdk.Coins, treasury, owner, sender sdk.AccAddress,
) *MsgCreateDataSource {
	return &MsgCreateDataSource{
		Name:        name,
		Description: description,
		Executable:  executable,
		Fee:         fee,
		Treasury:    treasury.String(),
		Owner:       owner.String(),
		Sender:      sender.String(),
	}
}

// ValidateBasic checks whether the given MsgCreateDataSource instance (sdk.Msg interface).
func (m MsgCreateDataSource) ValidateBasic() error {
	treasury, err := sdk.AccAddressFromBech32(m.Treasury)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(treasury); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("treasury: %s", m.Treasury)
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("owner: %s", m.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", m.Sender)
	}
	if len(m.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(m.Name), MaxNameLength)
	}
	if len(m.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(m.Description), MaxDescriptionLength)
	}
	if !m.Fee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Fee.String())
	}
	if len(m.Executable) == 0 {
		return ErrEmptyExecutable
	}
	if len(m.Executable) > MaxExecutableSize {
		return WrapMaxError(ErrTooLargeExecutable, len(m.Executable), MaxExecutableSize)
	}
	if bytes.Equal(m.Executable, DoNotModifyBytes) {
		return ErrCreateWithDoNotModify
	}
	return nil
}

// NewMsgEditDataSource creates a new MsgEditDataSource instance
func NewMsgEditDataSource(
	dataSourceID DataSourceID,
	name string,
	description string,
	executable []byte,
	fee sdk.Coins,
	treasury, owner, sender sdk.AccAddress,
) *MsgEditDataSource {
	return &MsgEditDataSource{
		DataSourceID: dataSourceID,
		Name:         name,
		Description:  description,
		Executable:   executable,
		Fee:          fee,
		Treasury:     treasury.String(),
		Owner:        owner.String(),
		Sender:       sender.String(),
	}
}

// ValidateBasic checks whether the given MsgEditDataSource instance (sdk.Msg interface).
func (m MsgEditDataSource) ValidateBasic() error {
	treasury, err := sdk.AccAddressFromBech32(m.Treasury)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(treasury); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("treasury: %s", m.Treasury)
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("owner: %s", m.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", m.Sender)
	}
	if len(m.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(m.Name), MaxNameLength)
	}
	if len(m.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(m.Description), MaxDescriptionLength)
	}
	if !m.Fee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Fee.String())
	}
	if len(m.Executable) == 0 {
		return ErrEmptyExecutable
	}
	if len(m.Executable) > MaxExecutableSize {
		return WrapMaxError(ErrTooLargeExecutable, len(m.Executable), MaxExecutableSize)
	}
	return nil
}

// NewMsgCreateOracleScript creates a new MsgCreateOracleScript instance
func NewMsgCreateOracleScript(
	name, description, schema, sourceCodeURL string, code []byte, owner, sender sdk.AccAddress,
) *MsgCreateOracleScript {
	return &MsgCreateOracleScript{
		Name:          name,
		Description:   description,
		Schema:        schema,
		SourceCodeURL: sourceCodeURL,
		Code:          code,
		Owner:         owner.String(),
		Sender:        sender.String(),
	}
}

// ValidateBasic checks whether the given MsgCreateOracleScript instance (sdk.Msg interface).
func (m MsgCreateOracleScript) ValidateBasic() error {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("owner: %s", m.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", m.Sender)
	}
	if len(m.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(m.Name), MaxNameLength)
	}
	if len(m.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(m.Description), MaxDescriptionLength)
	}
	if len(m.Schema) > MaxSchemaLength {
		return WrapMaxError(ErrTooLongSchema, len(m.Schema), MaxSchemaLength)
	}
	if len(m.SourceCodeURL) > MaxURLLength {
		return WrapMaxError(ErrTooLongURL, len(m.SourceCodeURL), MaxURLLength)
	}
	if len(m.Code) == 0 {
		return ErrEmptyWasmCode
	}
	if len(m.Code) > MaxWasmCodeSize {
		return WrapMaxError(ErrTooLargeWasmCode, len(m.Code), MaxWasmCodeSize)
	}
	if bytes.Equal(m.Code, DoNotModifyBytes) {
		return ErrCreateWithDoNotModify
	}
	return nil
}

// NewMsgEditOracleScript creates a new MsgEditOracleScript instance
func NewMsgEditOracleScript(
	oracleScriptID OracleScriptID,
	name, description, schema, sourceCodeURL string,
	code []byte,
	owner, sender sdk.AccAddress,
) *MsgEditOracleScript {
	return &MsgEditOracleScript{
		OracleScriptID: oracleScriptID,
		Name:           name,
		Description:    description,
		Schema:         schema,
		SourceCodeURL:  sourceCodeURL,
		Code:           code,
		Owner:          owner.String(),
		Sender:         sender.String(),
	}
}

// ValidateBasic checks whether the given MsgEditOracleScript instance (sdk.Msg interface).
func (m MsgEditOracleScript) ValidateBasic() error {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("owner: %s", m.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", m.Sender)
	}
	if len(m.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(m.Name), MaxNameLength)
	}
	if len(m.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(m.Description), MaxDescriptionLength)
	}
	if len(m.Schema) > MaxSchemaLength {
		return WrapMaxError(ErrTooLongSchema, len(m.Schema), MaxSchemaLength)
	}
	if len(m.SourceCodeURL) > MaxURLLength {
		return WrapMaxError(ErrTooLongURL, len(m.SourceCodeURL), MaxURLLength)
	}
	if len(m.Code) == 0 {
		return ErrEmptyWasmCode
	}
	if len(m.Code) > MaxWasmCodeSize {
		return WrapMaxError(ErrTooLargeWasmCode, len(m.Code), MaxWasmCodeSize)
	}
	return nil
}

// NewMsgActivate creates a new MsgActivate instance
func NewMsgActivate(validator sdk.ValAddress) *MsgActivate {
	return &MsgActivate{
		Validator: validator.String(),
	}
}

// ValidateBasic checks whether the given MsgActivate instance (sdk.Msg interface).
func (m MsgActivate) ValidateBasic() error {
	val, err := sdk.ValAddressFromBech32(m.Validator)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("validator: %s", m.Validator)
	}
	return nil
}

// NewMsgActivate creates a new MsgActivate instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
