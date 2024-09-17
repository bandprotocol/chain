package testutil

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

const (
	maxDE = 10
)

type GroupContext struct {
	GroupID               tss.GroupID
	Accounts              []bandtesting.Account
	Round1Infos           []tss.Round1Info
	EncryptedSecretShares []tss.EncSecretShares
	OwnPubKeySigs         []tss.Signature
	OwnPrivKeys           []tss.Scalar
	Secrets               []tss.Scalar
	DEs                   [][]DEWithPrivateNonce
}

func NewGroupContext(
	ctx sdk.Context,
	k *tsskeeper.Keeper,
	n uint64,
	threshold uint64,
) (*GroupContext, error) {
	accounts := GenerateAccounts(n)
	members := make([]sdk.AccAddress, n)
	for i := 0; i < len(accounts); i++ {
		members[i] = accounts[i].Address
	}

	// Create Group
	groupID, err := k.CreateGroup(ctx, members, threshold, "test")
	if err != nil {
		return nil, err
	}

	des := make([][]DEWithPrivateNonce, n)
	for i := 0; i < len(des); i++ {
		des[i] = make([]DEWithPrivateNonce, 0, maxDE)
	}

	secrets := make([]tss.Scalar, n)
	for i := 0; i < len(secrets); i++ {
		secrets[i], err = tss.RandomScalar()
		if err != nil {
			return nil, err
		}
	}

	return &GroupContext{
		GroupID:  groupID,
		Accounts: accounts,
		DEs:      des,
		Secrets:  secrets,
	}, nil
}

func (g *GroupContext) GenerateDE(ctx sdk.Context, k *tsskeeper.Keeper) error {
	msgServer := tsskeeper.NewMsgServerImpl(k)

	for i := 0; i < len(g.Accounts); i++ {
		newDEs := make([]DEWithPrivateNonce, 0, maxDE)
		for len(g.DEs[i])+len(newDEs) < maxDE {
			newDE := GenerateDE(g.Secrets[i])
			newDEs = append(newDEs, newDE)
		}

		pubDEs := make([]types.DE, len(newDEs))
		for j, newDE := range newDEs {
			pubDEs[j] = newDE.PubDE
		}

		msgSubmitDE := types.MsgSubmitDEs{DEs: pubDEs, Sender: g.Accounts[i].Address.String()}
		if _, err := msgServer.SubmitDEs(ctx, &msgSubmitDE); err != nil {
			return err
		}

		g.DEs[i] = append(g.DEs[i], newDEs...)
	}

	return nil
}

func (g *GroupContext) PopDE(id int, pubD tss.Point, pubE tss.Point) (DEWithPrivateNonce, error) {
	for i, de := range g.DEs[id] {
		pubDE := de.PubDE

		if pubDE.PubD.String() == pubD.String() && pubDE.PubE.String() == pubE.String() {
			selectedDE := g.DEs[id][i]
			g.DEs[id][i] = g.DEs[id][len(g.DEs[id])-1]
			g.DEs[id] = g.DEs[id][:len(g.DEs[id])-1]

			return selectedDE, nil
		}
	}

	return DEWithPrivateNonce{}, fmt.Errorf("cannot find selected DE from member %d", id)
}

func (g *GroupContext) SubmitSignature(
	ctx sdk.Context,
	k *tsskeeper.Keeper,
	msgServer types.MsgServer,
	signingID tss.SigningID,
) error {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return err
	}

	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	if err != nil {
		return err
	}
	assignedMembers := types.AssignedMembers(sa.AssignedMembers)

	for _, am := range assignedMembers {
		accID := int(am.MemberID) - 1
		address := g.Accounts[accID].Address

		de, err := g.PopDE(accID, am.PubD, am.PubE)
		if err != nil {
			return err
		}

		sig, err := GenerateSignature(signing, assignedMembers, am.MemberID, de, g.OwnPrivKeys[accID])
		if err != nil {
			return err
		}

		msg := types.NewMsgSubmitSignature(signingID, am.MemberID, sig, address.String())
		if _, err := msgServer.SubmitSignature(ctx, msg); err != nil {
			return err
		}
	}

	return nil
}

func (g *GroupContext) SubmitRound1(ctx sdk.Context, k *tsskeeper.Keeper) error {
	msgServer := tsskeeper.NewMsgServerImpl(k)
	dkgContext, err := k.GetDKGContext(ctx, g.GroupID)
	if err != nil {
		return err
	}

	group, err := k.GetGroup(ctx, g.GroupID)
	if err != nil {
		return err
	}

	n := group.Size_
	threshold := group.Threshold

	round1Infos := make([]tss.Round1Info, n)
	for i := uint64(0); i < n; i++ {
		mid := tss.MemberID(i + 1)
		r1, err := tss.GenerateRound1Info(mid, threshold, dkgContext)
		if err != nil {
			return err
		}

		round1Infos[i] = *r1
		r1InfoMsg := types.NewRound1Info(
			mid,
			r1.CoefficientCommits,
			r1.OneTimePubKey,
			r1.A0Signature,
			r1.OneTimeSignature,
		)

		msg := types.NewMsgSubmitDKGRound1(g.GroupID, r1InfoMsg, g.Accounts[i].Address.String())
		if _, err = msgServer.SubmitDKGRound1(ctx, msg); err != nil {
			return err
		}
	}

	k.HandleProcessGroup(ctx, g.GroupID)
	group = k.MustGetGroup(ctx, g.GroupID)
	if group.Status != types.GROUP_STATUS_ROUND_2 {
		return fmt.Errorf("unexpected group status: %s", group.Status.String())
	}

	g.Round1Infos = round1Infos
	return nil
}

func (g *GroupContext) SubmitRound2(ctx sdk.Context, k *tsskeeper.Keeper) error {
	msgServer := tsskeeper.NewMsgServerImpl(k)
	group, err := k.GetGroup(ctx, g.GroupID)
	if err != nil {
		return err
	}

	n := group.Size_
	r1Infos := g.Round1Infos

	// Get one-time public keys
	oneTimePubKeys := make(tss.Points, n)
	for i := uint64(0); i < n; i++ {
		oneTimePubKeys[i] = r1Infos[i].OneTimePubKey
	}

	encSecrets := make([]tss.EncSecretShares, n)
	for i := uint64(0); i < n; i++ {
		mid := tss.MemberID(i + 1)

		// Compute encrypted secret shares
		encSecretShares, err := tss.ComputeEncryptedSecretShares(
			mid,
			r1Infos[i].OneTimePrivKey,
			oneTimePubKeys,
			r1Infos[i].Coefficients,
			tss.DefaultNonce16Generator{},
		)
		if err != nil {
			return err
		}

		encSecrets[i] = encSecretShares
		r2Info := types.NewRound2Info(mid, encSecretShares)
		msg := types.NewMsgSubmitDKGRound2(g.GroupID, r2Info, g.Accounts[i].Address.String())

		if _, err = msgServer.SubmitDKGRound2(ctx, msg); err != nil {
			return err
		}
	}

	k.HandleProcessGroup(ctx, g.GroupID)
	group = k.MustGetGroup(ctx, g.GroupID)

	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return fmt.Errorf("unexpected group status: %s", group.Status.String())
	}

	g.EncryptedSecretShares = encSecrets
	return nil
}

func (g *GroupContext) SubmitRound3(ctx sdk.Context, k *tsskeeper.Keeper) error {
	msgServer := tsskeeper.NewMsgServerImpl(k)
	group, err := k.GetGroup(ctx, g.GroupID)
	if err != nil {
		return err
	}
	dkgContext, err := k.GetDKGContext(ctx, g.GroupID)
	if err != nil {
		return err
	}

	n := group.Size_
	r1Infos := g.Round1Infos
	encSecrets := g.EncryptedSecretShares

	// Get one-time public keys
	oneTimePubKeys := make(tss.Points, n)
	for i := uint64(0); i < n; i++ {
		oneTimePubKeys[i] = r1Infos[i].OneTimePubKey
	}

	ownPubKeySigs := make([]tss.Signature, n)
	ownPrivKeys := make([]tss.Scalar, n)
	for i := uint64(0); i < n; i++ {
		midI := tss.MemberID(i + 1)

		secretShares, err := computeSecretShares(r1Infos, encSecrets, midI)
		if err != nil {
			return err
		}

		ownPrivKey, err := tss.ComputeOwnPrivateKey(secretShares...)
		if err != nil {
			return err
		}

		ownPubKeySig, err := tss.SignOwnPubKey(
			midI,
			dkgContext,
			ownPrivKey.Point(),
			ownPrivKey,
		)
		if err != nil {
			return err
		}

		ownPubKeySigs[i] = ownPubKeySig
		ownPrivKeys[i] = ownPrivKey
		_, err = msgServer.Confirm(
			ctx,
			types.NewMsgConfirm(g.GroupID, midI, ownPubKeySig, g.Accounts[i].Address.String()),
		)
		if err != nil {
			return err
		}
	}

	k.HandleProcessGroup(ctx, g.GroupID)
	group = k.MustGetGroup(ctx, g.GroupID)

	if group.Status != types.GROUP_STATUS_ACTIVE {
		return fmt.Errorf("unexpected group status: %s", group.Status.String())
	}

	g.OwnPubKeySigs = ownPubKeySigs
	g.OwnPrivKeys = ownPrivKeys

	return nil
}

func CompleteGroupCreation(
	ctx sdk.Context,
	k *tsskeeper.Keeper,
	n uint64,
	threshold uint64,
) (*GroupContext, error) {
	groupCtx, err := NewGroupContext(ctx, k, n, threshold)
	if err != nil {
		return nil, err
	}

	if err = groupCtx.SubmitRound1(ctx, k); err != nil {
		return nil, err
	}

	if err = groupCtx.SubmitRound2(ctx, k); err != nil {
		return nil, err
	}

	if err = groupCtx.SubmitRound3(ctx, k); err != nil {
		return nil, err
	}

	if err = groupCtx.GenerateDE(ctx, k); err != nil {
		return nil, err
	}

	return groupCtx, nil
}

func computeSecretShares(
	r1Infos []tss.Round1Info,
	encSecrets []tss.EncSecretShares,
	mid tss.MemberID,
) (tss.Scalars, error) {
	n := uint64(len(r1Infos))

	secretShares := make(tss.Scalars, n)
	for i := uint64(0); i < n; i++ {
		if i == uint64(mid)-1 {
			secretShare, err := tss.ComputeSecretShare(r1Infos[i].Coefficients, mid)
			if err != nil {
				return tss.Scalars{}, err
			}

			secretShares[i] = secretShare
			continue
		}

		keySym, err := tss.ComputeSecretSym(r1Infos[mid-1].OneTimePrivKey, r1Infos[i].OneTimePubKey)
		if err != nil {
			return tss.Scalars{}, err
		}

		shifted := 0
		if uint64(mid)-1 > i {
			shifted = 1
		}

		secretShare, err := tss.DecryptSecretShare(encSecrets[i][int(mid)-1-shifted], keySym)
		if err != nil {
			return tss.Scalars{}, err
		}

		if err = tss.VerifySecretShare(mid, secretShare, r1Infos[i].CoefficientCommits); err != nil {
			return tss.Scalars{}, err
		}
		secretShares[i] = secretShare
	}

	return secretShares, nil
}
