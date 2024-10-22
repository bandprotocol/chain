package grant

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AddGranteeCmd returns a command to add a grantee to a granter.
func AddGranteeCmd(msgGrants []string, flagExpiration string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		exp, err := cmd.Flags().GetInt64(flagExpiration)
		if err != nil {
			return err
		}
		expTime := time.Unix(exp, 0)

		granter := clientCtx.GetFromAddress()
		msgs := []sdk.Msg{}

		for _, arg := range args {
			grantee, err := sdk.AccAddressFromBech32(arg)
			if err != nil {
				return err
			}

			gMsgs, err := combineGrantMsgs(granter, grantee, msgGrants, &expTime)
			if err != nil {
				return err
			}
			msgs = append(msgs, gMsgs...)
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
	}
}

// RemoveGranteeCmd returns a command to remove a grantee from a granter.
func RemoveGranteeCmd(msgRevokes []string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		granter := clientCtx.GetFromAddress()
		msgs := []sdk.Msg{}

		for _, arg := range args {
			grantee, err := sdk.AccAddressFromBech32(arg)
			if err != nil {
				return err
			}

			rMsgs, err := combineRevokeMsgs(granter, grantee, msgRevokes)
			if err != nil {
				return err
			}
			msgs = append(msgs, rMsgs...)
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
	}
}

// combineGrantMsgs combines multiple grant messages into a single slice of messages.
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

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// combineRevokeMsgs combines multiple revoke messages into a single slice of messages.
func combineRevokeMsgs(granter sdk.AccAddress, grantee sdk.AccAddress, msgRevokes []string) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msgRevoke := range msgRevokes {
		msg := authz.NewMsgRevoke(
			granter,
			grantee,
			msgRevoke,
		)

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
