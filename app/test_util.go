package app

import (
	"io"

	"github.com/tendermint/tendermint/libs/log"

	dbm "github.com/tendermint/tendermint/libs/db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// used for debugging by cmd/chatdebug
// NOTE to not use this function with non-test code
func NewChatAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) (cApp *ChatApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	cApp = NewChatApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return cApp, cApp.keyMain, cApp.keyStaking, cApp.stakingKeeper
}
