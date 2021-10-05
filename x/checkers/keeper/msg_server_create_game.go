package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rules "github.com/xavierlepretre/checkers/x/checkers/rules"
	"github.com/xavierlepretre/checkers/x/checkers/types"
)

func (k msgServer) CreateGame(goCtx context.Context, msg *types.MsgCreateGame) (*types.MsgCreateGameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	nextGame, found := k.Keeper.GetNextGame(ctx)
	if !found {
		panic("NextGame not found")
	}
	newIndex := strconv.FormatUint(nextGame.IdValue, 10)
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.ErrInvalidCreator
	}
	red, err := sdk.AccAddressFromBech32(msg.Red)
	if err != nil {
		return nil, types.ErrInvalidRed
	}
	black, err := sdk.AccAddressFromBech32(msg.Black)
	if err != nil {
		return nil, types.ErrInvalidBlack
	}
	newGame := types.FullGame{
		Creator:   creator,
		Index:     newIndex,
		Game:      *rules.New(),
		Red:       red,
		Black:     black,
		MoveCount: 0,
		Deadline:  ctx.BlockTime().Add(types.MaxTurnDurationInSeconds),
		Winner:    rules.NO_PLAYER.Color,
		Wager:     sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(msg.Wager))),
	}
	storedGame := newGame.ToStoredGame()
	k.Keeper.SendToFifoTail(ctx, &storedGame, &nextGame)
	k.Keeper.SetStoredGame(ctx, storedGame)

	nextGame.IdValue++
	k.Keeper.SetNextGame(ctx, nextGame)

	// What to emit
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.StoredGameEventKey),
			sdk.NewAttribute(types.StoredGameEventCreator, msg.Creator),
			sdk.NewAttribute(types.StoredGameEventIndex, newIndex),
			sdk.NewAttribute(types.StoredGameEventRed, msg.Red),
			sdk.NewAttribute(types.StoredGameEventBlack, msg.Black),
			sdk.NewAttribute(types.StoredGameEventWager, strconv.FormatUint(msg.Wager, 10)),
		),
	)

	return &types.MsgCreateGameResponse{
		IdValue: newIndex,
	}, nil
}
