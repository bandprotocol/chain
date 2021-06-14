package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// oracle message types
const (
	TypeMsgRequestData        = "request"
	TypeMsgReportData         = "report"
	TypeMsgCreateDataSource   = "create_data_source"
	TypeMsgEditDataSource     = "edit_data_source"
	TypeMsgCreateOracleScript = "create_oracle_script"
	TypeMsgEditOracleScript   = "edit_oracle_script"
	TypeMsgActivate           = "activate"
	TypeMsgAddReporter        = "add_reporter"
	TypeMsgRemoveReporter     = "remove_reporter"
)

var (
	_ sdk.Msg = &MsgRequestData{}
	_ sdk.Msg = &MsgReportData{}
	_ sdk.Msg = &MsgCreateDataSource{}
	_ sdk.Msg = &MsgEditDataSource{}
	_ sdk.Msg = &MsgCreateOracleScript{}
	_ sdk.Msg = &MsgEditOracleScript{}
	_ sdk.Msg = &MsgActivate{}
	_ sdk.Msg = &MsgAddReporter{}
	_ sdk.Msg = &MsgRemoveReporter{}
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

// Route returns the route of MsgRequestData - "oracle" (sdk.Msg interface).
func (msg MsgRequestData) Route() string { return RouterKey }

// Type returns the message type of MsgRequestData (sdk.Msg interface).
func (msg MsgRequestData) Type() string { return TypeMsgRequestData }

// ValidateBasic checks whether the given MsgRequestData instance (sdk.Msg interface).
func (msg MsgRequestData) ValidateBasic() error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %s", msg.Sender)
	}
	if len(msg.Calldata) > MaxDataSize {
		return WrapMaxError(ErrTooLargeCalldata, len(msg.Calldata), MaxDataSize)
	}
	if msg.MinCount <= 0 {
		return sdkerrors.Wrapf(ErrInvalidMinCount, "got: %d", msg.MinCount)
	}
	if msg.AskCount < msg.MinCount {
		return sdkerrors.Wrapf(ErrInvalidAskCount, "got: %d, min count: %d", msg.AskCount, msg.MinCount)
	}
	if len(msg.ClientID) > MaxClientIDLength {
		return WrapMaxError(ErrTooLongClientID, len(msg.ClientID), MaxClientIDLength)
	}
	if msg.PrepareGas <= 0 {
		return sdkerrors.Wrapf(ErrInvalidOwasmGas, "invalid prepare gas: %d", msg.PrepareGas)
	}
	if msg.ExecuteGas <= 0 {
		return sdkerrors.Wrapf(ErrInvalidOwasmGas, "invalid execute gas: %d", msg.ExecuteGas)
	}
	if msg.PrepareGas+msg.ExecuteGas > MaximumOwasmGas {
		return sdkerrors.Wrapf(ErrInvalidOwasmGas, "sum of prepare gas and execute gas (%d) exceed %d", msg.PrepareGas+msg.ExecuteGas, MaximumOwasmGas)
	}
	if !msg.FeeLimit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.FeeLimit.String())
	}
	return nil
}

// GetSigners returns the required signers for the given MsgRequestData (sdk.Msg interface).
func (msg MsgRequestData) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgRequestData) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgReportData creates a new MsgReportData instance
func NewMsgReportData(requestID RequestID, rawReports []RawReport, validator sdk.ValAddress, reporter sdk.AccAddress) *MsgReportData {
	return &MsgReportData{
		RequestID:  requestID,
		RawReports: rawReports,
		Validator:  validator.String(),
		Reporter:   reporter.String(),
	}
}

// Route returns the route of MsgReportData - "oracle" (sdk.Msg interface).
func (msg MsgReportData) Route() string { return RouterKey }

// Type returns the message type of MsgReportData (sdk.Msg interface).
func (msg MsgReportData) Type() string { return TypeMsgReportData }

// ValidateBasic checks whether the given MsgReportData instance (sdk.Msg interface).
func (msg MsgReportData) ValidateBasic() error {
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return err
	}
	repAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(valAddr); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "validator: %s", msg.Validator)
	}
	if err := sdk.VerifyAddressFormat(repAddr); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "reporter: %s", msg.Reporter)
	}
	if len(msg.RawReports) == 0 {
		return ErrEmptyReport
	}
	uniqueMap := make(map[ExternalID]bool)
	for _, r := range msg.RawReports {
		if _, found := uniqueMap[r.ExternalID]; found {
			return sdkerrors.Wrapf(ErrDuplicateExternalID, "external id: %d", r.ExternalID)
		}
		uniqueMap[r.ExternalID] = true
		if len(r.Data) > MaxDataSize {
			return WrapMaxError(ErrTooLargeRawReportData, len(r.Data), MaxDataSize)
		}
	}
	return nil
}

// GetSigners returns the required signers for the given MsgReportData (sdk.Msg interface).
func (msg MsgReportData) GetSigners() []sdk.AccAddress {
	reporter, _ := sdk.AccAddressFromBech32(msg.Reporter)
	return []sdk.AccAddress{reporter}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgReportData) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
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

// Route returns the route of MsgCreateDataSource - "oracle" (sdk.Msg interface).
func (msg MsgCreateDataSource) Route() string { return RouterKey }

// Type returns the message type of MsgCreateDataSource (sdk.Msg interface).
func (msg MsgCreateDataSource) Type() string { return TypeMsgCreateDataSource }

// ValidateBasic checks whether the given MsgCreateDataSource instance (sdk.Msg interface).
func (msg MsgCreateDataSource) ValidateBasic() error {
	treasury, err := sdk.AccAddressFromBech32(msg.Treasury)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(treasury); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "treasury: %s", msg.Treasury)
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "owner: %s", msg.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %s", msg.Sender)
	}
	if len(msg.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(msg.Name), MaxNameLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(msg.Description), MaxDescriptionLength)
	}
	if !msg.Fee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Fee.String())
	}
	if len(msg.Executable) == 0 {
		return ErrEmptyExecutable
	}
	if len(msg.Executable) > MaxExecutableSize {
		return WrapMaxError(ErrTooLargeExecutable, len(msg.Executable), MaxExecutableSize)
	}
	if bytes.Equal(msg.Executable, DoNotModifyBytes) {
		return ErrCreateWithDoNotModify
	}
	return nil
}

// GetSigners returns the required signers for the given MsgCreateDataSource (sdk.Msg interface).
func (msg MsgCreateDataSource) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgCreateDataSource) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgEditDataSource creates a new MsgEditDataSource instance
func NewMsgEditDataSource(
	dataSourceID DataSourceID, name string, description string, executable []byte, fee sdk.Coins, treasury, owner, sender sdk.AccAddress,
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

// Route returns the route of MsgEditDataSource - "oracle" (sdk.Msg interface).
func (msg MsgEditDataSource) Route() string { return RouterKey }

// Type returns the message type of MsgEditDataSource (sdk.Msg interface).
func (msg MsgEditDataSource) Type() string { return TypeMsgEditDataSource }

// ValidateBasic checks whether the given MsgEditDataSource instance (sdk.Msg interface).
func (msg MsgEditDataSource) ValidateBasic() error {
	treasury, err := sdk.AccAddressFromBech32(msg.Treasury)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(treasury); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "treasury: %s", msg.Treasury)
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "owner: %s", msg.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %s", msg.Sender)
	}
	if len(msg.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(msg.Name), MaxNameLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(msg.Description), MaxDescriptionLength)
	}
	if !msg.Fee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Fee.String())
	}
	if len(msg.Executable) == 0 {
		return ErrEmptyExecutable
	}
	if len(msg.Executable) > MaxExecutableSize {
		return WrapMaxError(ErrTooLargeExecutable, len(msg.Executable), MaxExecutableSize)
	}
	return nil
}

// GetSigners returns the required signers for the given MsgEditDataSource (sdk.Msg interface).
func (msg MsgEditDataSource) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgEditDataSource) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
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

// Route returns the route of MsgCreateOracleScript - "oracle" (sdk.Msg interface).
func (msg MsgCreateOracleScript) Route() string { return RouterKey }

// Type returns the message type of MsgCreateOracleScript (sdk.Msg interface).
func (msg MsgCreateOracleScript) Type() string { return TypeMsgCreateOracleScript }

// ValidateBasic checks whether the given MsgCreateOracleScript instance (sdk.Msg interface).
func (msg MsgCreateOracleScript) ValidateBasic() error {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "owner: %s", msg.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %s", msg.Sender)
	}
	if len(msg.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(msg.Name), MaxNameLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(msg.Description), MaxDescriptionLength)
	}
	if len(msg.Schema) > MaxSchemaLength {
		return WrapMaxError(ErrTooLongSchema, len(msg.Schema), MaxSchemaLength)
	}
	if len(msg.SourceCodeURL) > MaxURLLength {
		return WrapMaxError(ErrTooLongURL, len(msg.SourceCodeURL), MaxURLLength)
	}
	if len(msg.Code) == 0 {
		return ErrEmptyWasmCode
	}
	if len(msg.Code) > MaxWasmCodeSize {
		return WrapMaxError(ErrTooLargeWasmCode, len(msg.Code), MaxWasmCodeSize)
	}
	if bytes.Equal(msg.Code, DoNotModifyBytes) {
		return ErrCreateWithDoNotModify
	}
	return nil
}

// GetSigners returns the required signers for the given MsgCreateOracleScript (sdk.Msg interface).
func (msg MsgCreateOracleScript) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgCreateOracleScript) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgEditOracleScript creates a new MsgEditOracleScript instance
func NewMsgEditOracleScript(
	oracleScriptID OracleScriptID, name, description, schema, sourceCodeURL string, code []byte, owner, sender sdk.AccAddress,
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

// Route returns the route of MsgEditOracleScript - "oracle" (sdk.Msg interface).
func (msg MsgEditOracleScript) Route() string { return RouterKey }

// Type returns the message type of MsgEditOracleScript (sdk.Msg interface).
func (msg MsgEditOracleScript) Type() string { return TypeMsgEditOracleScript }

// ValidateBasic checks whether the given MsgEditOracleScript instance (sdk.Msg interface).
func (msg MsgEditOracleScript) ValidateBasic() error {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "owner: %s", msg.Owner)
	}
	if err := sdk.VerifyAddressFormat(sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %s", msg.Sender)
	}
	if len(msg.Name) > MaxNameLength {
		return WrapMaxError(ErrTooLongName, len(msg.Name), MaxNameLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return WrapMaxError(ErrTooLongDescription, len(msg.Description), MaxDescriptionLength)
	}
	if len(msg.Schema) > MaxSchemaLength {
		return WrapMaxError(ErrTooLongSchema, len(msg.Schema), MaxSchemaLength)
	}
	if len(msg.SourceCodeURL) > MaxURLLength {
		return WrapMaxError(ErrTooLongURL, len(msg.SourceCodeURL), MaxURLLength)
	}
	if len(msg.Code) == 0 {
		return ErrEmptyWasmCode
	}
	if len(msg.Code) > MaxWasmCodeSize {
		return WrapMaxError(ErrTooLargeWasmCode, len(msg.Code), MaxWasmCodeSize)
	}
	return nil
}

// GetSigners returns the required signers for the given MsgEditOracleScript (sdk.Msg interface).
func (msg MsgEditOracleScript) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgEditOracleScript) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgActivate creates a new MsgActivate instance
func NewMsgActivate(validator sdk.ValAddress) *MsgActivate {
	return &MsgActivate{
		Validator: validator.String(),
	}
}

// Route returns the route of MsgActivate - "oracle" (sdk.Msg interface).
func (msg MsgActivate) Route() string { return RouterKey }

// Type returns the message type of MsgActivate (sdk.Msg interface).
func (msg MsgActivate) Type() string { return TypeMsgActivate }

// ValidateBasic checks whether the given MsgActivate instance (sdk.Msg interface).
func (msg MsgActivate) ValidateBasic() error {
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "validator: %s", msg.Validator)
	}
	return nil
}

// GetSigners returns the required signers for the given MsgActivate (sdk.Msg interface).
func (msg MsgActivate) GetSigners() []sdk.AccAddress {
	val, _ := sdk.ValAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{sdk.AccAddress(val)}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgActivate) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgAddReporter creates a new MsgAddReporter instance
func NewMsgAddReporter(validator sdk.ValAddress, reporter sdk.AccAddress) *MsgAddReporter {
	return &MsgAddReporter{
		Validator: validator.String(),
		Reporter:  reporter.String(),
	}
}

// Route returns the route of MsgAddReporter - "oracle" (sdk.Msg interface).
func (msg MsgAddReporter) Route() string { return RouterKey }

// Type returns the message type of MsgAddReporter (sdk.Msg interface).
func (msg MsgAddReporter) Type() string { return TypeMsgAddReporter }

// ValidateBasic checks whether the given MsgAddReporter instance (sdk.Msg interface).
func (msg MsgAddReporter) ValidateBasic() error {
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return err
	}
	rep, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "validator: %s", msg.Validator)
	}
	if err := sdk.VerifyAddressFormat(rep); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "reporter: %s", msg.Reporter)
	}
	if sdk.ValAddress(rep).Equals(val) {
		return ErrSelfReferenceAsReporter
	}
	return nil
}

// GetSigners returns the required signers for the given MsgAddReporter (sdk.Msg interface).
func (msg MsgAddReporter) GetSigners() []sdk.AccAddress {
	val, _ := sdk.ValAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{sdk.AccAddress(val)}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgAddReporter) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// NewMsgRemoveReporter creates a new MsgRemoveReporter instance
func NewMsgRemoveReporter(validator sdk.ValAddress, reporter sdk.AccAddress) *MsgRemoveReporter {
	return &MsgRemoveReporter{
		Validator: validator.String(),
		Reporter:  reporter.String(),
	}
}

// Route returns the route of MsgRemoveReporter - "oracle" (sdk.Msg interface).
func (msg MsgRemoveReporter) Route() string { return RouterKey }

// Type returns the message type of MsgRemoveReporter (sdk.Msg interface).
func (msg MsgRemoveReporter) Type() string { return TypeMsgRemoveReporter }

// ValidateBasic checks whether the given MsgRemoveReporter instance (sdk.Msg interface).
func (msg MsgRemoveReporter) ValidateBasic() error {
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return err
	}
	rep, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(val); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "validator: %s", msg.Validator)
	}
	if err := sdk.VerifyAddressFormat(rep); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "reporter: %s", msg.Reporter)
	}
	if sdk.ValAddress(rep).Equals(val) {
		return ErrSelfReferenceAsReporter
	}
	return nil
}

// GetSigners returns the required signers for the given MsgRemoveReporter (sdk.Msg interface).
func (msg MsgRemoveReporter) GetSigners() []sdk.AccAddress {
	val, _ := sdk.ValAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{sdk.AccAddress(val)}
}

// GetSignBytes returns raw JSON bytes to be signed by the signers (sdk.Msg interface).
func (msg MsgRemoveReporter) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}
