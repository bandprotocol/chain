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
			return []sdk.Msg{}, err
		}

		err = msg.ValidateBasic()
		if err != nil {
			return []sdk.Msg{}, err
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func combineRevokeMsgs(granter sdk.AccAddress, grantee sdk.AccAddress, msgRevoke []string) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msg := range msgRevoke {
		msg := authz.NewMsgRevoke(
			granter,
			grantee,
			msg,
		)

		err := msg.ValidateBasic()
		if err != nil {
			return []sdk.Msg{}, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
