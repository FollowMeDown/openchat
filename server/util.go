package server

import (
	"os"
	"path/filepath"
	"time"

	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/openchatproject/openchat/version"
)


// server context

func NewDefaultContext() *server.Context {
	return server.NewContext(
		cfg.DefaultConfig(),
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	)
}

//___________________________________________________________________________________

// PersistentPreRunEFn returns a PersistentPreRunE function for cobra
// that initailizes the passed in context with a properly configured
// logger and config object.
func PersistentPreRunEFn(context *server.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == version.VersionCmd.Name() {
			return nil
		}
		config, err := interceptLoadConfig()
		if err != nil {
			return err
		}
		err = validateConfig(config)
		if err != nil {
			return err
		}
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		context.Config = config
		context.Logger = logger
		return nil
	}
}

// If a new config is created, change some of the default tendermint settings
func interceptLoadConfig() (conf *cfg.Config, err error) {
	tmpConf := cfg.DefaultConfig()
	err = viper.Unmarshal(tmpConf)
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}
	rootDir := tmpConf.RootDir
	configFilePath := filepath.Join(rootDir, "config/config.toml")
	// Intercept only if the file doesn't already exist

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// the following parse config is needed to create directories
		conf, _ = tcmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
		conf.ProfListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.TxIndex.IndexAllTags = true
		conf.Consensus.TimeoutCommit = 5 * time.Second
		cfg.WriteConfigFile(configFilePath, conf)
		// Fall through, just so that its parsed into memory.
	}

	if conf == nil {
		conf, err = tcmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
	}

	// create a default chatd config file if it does not exist
	chatdConfigFilePath := filepath.Join(rootDir, "config/chatd.toml")
	if _, err := os.Stat(chatdConfigFilePath); os.IsNotExist(err) {
		appConf, _ := config.ParseConfig()
		config.WriteConfigFile(chatdConfigFilePath, appConf)
	}

	viper.SetConfigName("chatd")
	err = viper.MergeInConfig()

	return
}

// validate the config with the sdk's requirements.
func validateConfig(conf *cfg.Config) error {
	if !conf.Consensus.CreateEmptyBlocks {
		return errors.New("config option CreateEmptyBlocks = false is currently unsupported")
	}
	return nil
}

// add server commands
func AddCommands(
	ctx *server.Context, cdc *codec.Codec,
	rootCmd *cobra.Command,
	appCreator server.AppCreator, appExport server.AppExporter) {

	rootCmd.PersistentFlags().String("log_level", ctx.Config.LogLevel, "Log level")

	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		server.ShowNodeIDCmd(ctx),
		server.ShowValidatorCmd(ctx),
		server.ShowAddressCmd(ctx),
		server.VersionCmd(ctx),
	)

	rootCmd.AddCommand(
		server.StartCmd(ctx, appCreator),
		server.UnsafeResetAllCmd(ctx),
		client.LineBreak,
		tendermintCmd,
		server.ExportCmd(ctx, cdc, appExport),
		client.LineBreak,
		version.VersionCmd,
	)
}
