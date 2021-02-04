package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDataSource(
	Owner sdk.AccAddress,
	Name string,
	Description string,
	Filename string,
) DataSource {
	return DataSource{
		Owner:       Owner.String(),
		Name:        Name,
		Description: Description,
		Filename:    Filename,
	}
}
