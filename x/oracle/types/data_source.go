package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDataSource(
	owner sdk.AccAddress, name, description, filename string, treasury sdk.AccAddress, fee sdk.Coins,
) DataSource {
	return DataSource{
		Owner:       owner.String(),
		Name:        name,
		Description: description,
		Filename:    filename,
		Treasury:    treasury.String(),
		Fee:         fee,
	}
}
