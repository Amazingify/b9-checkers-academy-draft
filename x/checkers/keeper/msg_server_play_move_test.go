package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/b9lab/checkers/testutil/keeper"
	"github.com/b9lab/checkers/x/checkers"
	"github.com/b9lab/checkers/x/checkers/keeper"
	"github.com/b9lab/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func setupMsgServerWithOneGameForPlayMove(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context) {
	k, ctx := keepertest.CheckersKeeper(t)
	checkers.InitGenesis(ctx, *k, *types.DefaultGenesis())
	// * this gets an object that can send
	// * the messages in the system
	server := keeper.NewMsgServerImpl(*k)
	context := sdk.WrapSDKContext(ctx)
	server.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Black:   bob,
		Red:     carol,
	})

	return server, *k, context
}

func TestPlayMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(
		context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       2,
			ToY:       3,
		})

	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestGameSavedCorrectly(t *testing.T) {

	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	playMoveResponse, err := msgServer.PlayMove(
		context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       2,
			ToY:       3,
		})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)

	game, found := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found)
	require.EqualValues(t, game.Turn, "r")
	require.EqualValues(t, game.Board, "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*")
}

func TestGameNotFound(t *testing.T) {
	_, keeper, context := setupMsgServerCreateGame(t)
	_, found := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "200")
	require.False(t, found)
}

func TestSenderNotPlayer(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	_, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   alice,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.EqualValues(t, alice+": message creator is not a player", err.Error())
}

func TestMoveOutOfTurn(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	_, err := msgServer.PlayMove(
		context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       2,
			ToY:       3,
		})
	require.Nil(t, err)

	_, err = msgServer.PlayMove(
		context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       2,
			ToY:       3,
		})
	require.EqualValues(t, "{black}: player tried to play out of turn", err.Error())
}

func TestWrongMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	_, err := msgServer.PlayMove(
		context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       0,
			ToY:       0,
		})
	require.EqualValues(t, "Invalid move: {1 2} to {0 0}: wrong move", err.Error())
}

func Test2MovesSavedGame(t *testing.T) {

	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context,
		&types.MsgPlayMove{
			Creator:   bob,
			GameIndex: "1",
			FromX:     1,
			FromY:     2,
			ToX:       2,
			ToY:       3,
		})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})

	systemInfo, found := keeper.GetSystemInfo(sdk.UnwrapSDKContext(context))
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId: 2,
	}, systemInfo)

	game1, found := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index: "1",
		Board: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:  "b",
		Black: bob,
		Red:   carol,
	}, game1)

}

func TestPlayMoveEventEmitted(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "move-played",
		Attributes: []sdk.Attribute{
			{Key: "creator", Value: bob},
			{Key: "game-index", Value: "1"},
			{Key: "captured-x", Value: "-1"},
			{Key: "captured-y", Value: "-1"},
			{Key: "winner", Value: "*"},
		},
	}, event)
}

func TestPlayed2MoveeVENTe(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})

	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	event := events[0]

	require.EqualValues(t, "move-played", event.Type)
	require.EqualValues(t, []sdk.Attribute{
		{Key: "creator", Value: carol},
		{Key: "game-index", Value: "1"},
		{Key: "captured-x", Value: "-1"},
		{Key: "captured-y", Value: "-1"},
		{Key: "winner", Value: "*"},
	}, event.Attributes[5:])
}
