package cli

import (
	"encoding/json"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// PointRequests defines a repeated slice of point objects.
type PointRequests struct {
	Points []types.Point
}

func parsePoints(pointsFile string) ([]types.Point, error) {
	points := PointRequests{}

	if pointsFile == "" {
		return points.Points, nil
	}

	contents, err := os.ReadFile(pointsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &points)
	if err != nil {
		return nil, err
	}

	return points.Points, nil
}

func combineGrantMsgs(granter sdk.AccAddress, grantee sdk.AccAddress, msgGrants []string, expiration *time.Time) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msgGrant := range msgGrants {
		mg, err := authz.NewMsgGrant(
			granter,
			grantee,
			authz.NewGenericAuthorization(msgGrant),
			expiration,
		)
		if err != nil {
			return []sdk.Msg{}, err
		}

		err = mg.ValidateBasic()
		if err != nil {
			return []sdk.Msg{}, err
		}

		msgs = append(msgs, mg)
	}

	return msgs, nil
}

func combineRevokeMsgs(granter sdk.AccAddress, grantee sdk.AccAddress, msgRevokes []string) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	for _, msg := range msgRevokes {
		me := authz.NewMsgRevoke(
			granter,
			grantee,
			msg,
		)

		err := me.ValidateBasic()
		if err != nil {
			return []sdk.Msg{}, err
		}

		msgs = append(msgs, &me)
	}

	return msgs, nil
}
