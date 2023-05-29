package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

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

		if err = msg.ValidateBasic(); err != nil {
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

		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

type ComplainsData struct {
	Complains []types.Complain `json:"complains"`
}

// parseComplains reads and parses a JSON file containing complaints into a slice of Complain objects.
func parseComplains(complainsFile string) ([]types.Complain, error) {
	var complainsData ComplainsData

	if complainsFile == "" {
		return complainsData.Complains, nil
	}

	contents, err := os.ReadFile(complainsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &complainsData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return complainsData.Complains, nil
}
