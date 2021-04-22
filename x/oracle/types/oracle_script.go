package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewOracleScript(
	Owner sdk.AccAddress,
	Name string,
	Description string,
	Filename string,
	Schema string,
	SourceCodeURL string,
) OracleScript {
	return OracleScript{
		Owner:         Owner.String(),
		Name:          Name,
		Description:   Description,
		Filename:      Filename,
		Schema:        Schema,
		SourceCodeURL: SourceCodeURL,
	}
}
