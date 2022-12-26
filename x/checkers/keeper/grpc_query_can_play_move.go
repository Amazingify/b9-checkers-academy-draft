package keeper

import (
	"context"
	"fmt"

	"github.com/b9lab/checkers/x/checkers/rules"
	"github.com/b9lab/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) CanPlayMove(goCtx context.Context, req *types.QueryCanPlayMoveRequest) (*types.QueryCanPlayMoveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	storedGame, found := k.GetStoredGame(ctx, req.GameIndex)

	if !found {
		return nil, sdkerrors.Wrapf(types.ErrGameNotFound, "%s", req.GameIndex)
	}

	if storedGame.Winner != rules.PieceStrings[rules.NO_PLAYER] {
		return &types.QueryCanPlayMoveResponse{
			Possible: false,
			Reason:   types.ErrGameFinished.Error(),
		}, nil
	}

	sender := req.Player
	var player rules.Player
	if sender != rules.PieceStrings[rules.BLACK_PLAYER] && sender != rules.PieceStrings[rules.RED_PLAYER] {
		return &types.QueryCanPlayMoveResponse{
			Possible: false,
			Reason:   fmt.Sprintf("%s: %s", types.ErrCreatorNotPlayer.Error(), req.Player),
		}, nil
	} else if sender == rules.PieceStrings[rules.BLACK_PLAYER] {
		player = rules.BLACK_PLAYER
	} else if sender == rules.PieceStrings[rules.RED_PLAYER] {
		player = rules.RED_PLAYER
	}

	game, err := storedGame.ParseGame()

	if err != nil {
		return nil, err
	}

	if !game.TurnIs(player) {
		return &types.QueryCanPlayMoveResponse{
			Possible: false,
			Reason:   fmt.Sprintf("%s: %s", types.ErrNotPlayerTurn.Error(), player.Color),
		}, nil
	}

	_, moveErr := game.Move(
		rules.Pos{
			X: int(req.FromX),
			Y: int(req.FromY),
		},
		rules.Pos{
			X: int(req.ToX),
			Y: int(req.ToY),
		},
	)

	if moveErr != nil {
		return &types.QueryCanPlayMoveResponse{
			Possible: false,
			Reason:   fmt.Sprintf("%s: %s", types.ErrWrongMove.Error(), moveErr.Error()),
		}, nil
	}

	// TODO: Process the query
	_ = ctx

	return &types.QueryCanPlayMoveResponse{
		Possible: true,
		Reason:   "ok",
	}, nil
}
