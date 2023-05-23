package cli

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

func combineGrantMsgs(
	granter sdk.AccAddress,
	grantee sdk.AccAddress,
	msgGrants []string,
	expiration *time.Time,
) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msgGrant := range msgGrants {
		msg, err := authz.NewMsgGrant(
			granter,
			grantee,
			authz.NewGenericAuthorization(msgGrant),
			expiration,
		)
		if err != nil {
			return nil, err
		}

		if err = msg.ValidateBasic(); err != nil {
			return nil, err
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func combineRevokeMsgs(granter sdk.AccAddress, grantee sdk.AccAddress, msgRevokes []string) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msgRevoke := range msgRevokes {
		msg := authz.NewMsgRevoke(
			granter,
			grantee,
			msgRevoke,
		)

		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
