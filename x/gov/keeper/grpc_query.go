package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ govtypes.QueryServer = Keeper{}

// TallyResult queries the tally of a proposal vote
func (k Keeper) TallyResult(c context.Context, req *govtypes.QueryTallyResultRequest) (*govtypes.QueryTallyResultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	proposal, ok := k.GetProposal(ctx, req.ProposalId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
	}

	var tallyResult govtypes.TallyResult

	switch {
	case proposal.Status == govtypes.StatusDepositPeriod:
		tallyResult = govtypes.EmptyTallyResult()

	case proposal.Status == govtypes.StatusPassed || proposal.Status == govtypes.StatusRejected:
		tallyResult = proposal.FinalTallyResult

	default:
		// proposal is in voting period
		_, _, tallyResult = k.Tally(ctx, proposal)
	}

	return &govtypes.QueryTallyResultResponse{Tally: tallyResult}, nil
}
