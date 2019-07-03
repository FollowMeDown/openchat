package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	//
	"github.com/openchatproject/openchat/app"
	chatServer "github.com/openchatproject/openchat/server"
	chatgenutil "github.com/openchatproject/openchat/x/genutil"
)

// chatd custom flags
const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	config.Seal()

	//sdk.PowerReduction = chat.MicroChatUnit.TruncateInt()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "chatd",
		Short:             "OpenChat Daemon (server)",
		PersistentPreRunE: chatServer.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(chatgenutil.InitCmd(ctx, cdc))
	rootCmd.AddCommand(chatgenutil.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(chatgenutil.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(chatgenutil.AddGenesisAccountCmd(ctx, cdc))
	rootCmd.AddCommand(chatgenutil.ValidateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(chatgenutil.TestnetCmd(ctx, cdc))

	chatServer.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "OC", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewChatApp(
		logger, db, traceStore, true, invCheckPeriod,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		cApp := app.NewChatApp(logger, db, traceStore, false, uint(1))
		err := cApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return cApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	cApp := app.NewChatApp(logger, db, traceStore, true, uint(1))
	return cApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
