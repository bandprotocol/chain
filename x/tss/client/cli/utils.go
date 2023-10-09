package cli

import (
	"encoding/json"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/x/tss/types"
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

type Complaints struct {
	Complaints []types.Complaint `json:"complaints"`
}

// parseComplaints reads and parses a JSON file containing complaints into a slice of complain objects.
func parseComplaints(complaintsFile string) ([]types.Complaint, error) {
	var complaints Complaints

	if complaintsFile == "" {
		return complaints.Complaints, nil
	}

	contents, err := os.ReadFile(complaintsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &complaints)
	if err != nil {
		return nil, err
	}

	return complaints.Complaints, nil
}
