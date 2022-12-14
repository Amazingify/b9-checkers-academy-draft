package keeper

import (
	"github.com/b9lab/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) RemoveFromFifo(ctx sdk.Context, game *types.StoredGame, info *types.SystemInfo) {

	// * if this game is not the last one
	if game.BeforeIndex != types.NoFifoIndex {
		beforeElement, found := k.GetStoredGame(ctx, game.BeforeIndex)
		if !found {
			panic("Element before in Fifo was not found")
		}

		beforeElement.AfterIndex = game.AfterIndex
		k.SetStoredGame(ctx, beforeElement)

		// * if there are no games after this one
		// * then our current element is the tail
		if game.AfterIndex == types.NoFifoIndex {
			info.FifoTailIndex = beforeElement.Index
		}
		// * is the game at the head ?
	} else if info.FifoHeadIndex == game.Index {
		info.FifoHeadIndex = game.AfterIndex
	}
	if game.AfterIndex != types.NoFifoIndex {
		afterElement, found := k.GetStoredGame(ctx, game.AfterIndex)
		if !found {
			panic("element after in Fifo was not found")
		}

		afterElement.BeforeIndex = game.BeforeIndex
		k.SetStoredGame(ctx, afterElement)

		if game.BeforeIndex == types.NoFifoIndex {
			info.FifoHeadIndex = afterElement.Index
		}
	} else if info.FifoHeadIndex == game.Index {
		info.FifoTailIndex = game.BeforeIndex
	}

	game.BeforeIndex = types.NoFifoIndex
	game.AfterIndex = types.NoFifoIndex
}
