package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewOracleScript(
	owner sdk.AccAddress,
	name string,
	description string,
	filename string,
	schema string,
	sourceCodeURL string,
) OracleScript {
	return OracleScript{
		Owner:         owner.String(),
		Name:          name,
		Description:   description,
		Filename:      filename,
		Schema:        schema,
		SourceCodeURL: sourceCodeURL,
	}
}
