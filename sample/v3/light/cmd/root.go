package cmd

import (
	"fmt"
	"log/slog"
	"os"

	env "github.com/gofurry/fiberx/v3/light/config"
	"github.com/gofurry/fiberx/v3/light/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = rootCmd.Help()
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           common.COMMON_PROJECT_NAME,
		Short:         "Awesome Fiber template service",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().String("config", "", "path to server config file")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(
		newServeCmd(),
		newVersionCmd(),
	)
	return rootCmd
}

func initConfig() error {
	env.ConfigureServerConfig(common.COMMON_PROJECT_NAME, "server.yaml", viper.GetString("config"))
	return env.InitServerConfig(common.COMMON_PROJECT_NAME)
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runService()
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show service version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := env.GetServerConfig()
			_, appName := appIdentity()
			slog.Info(appName + " " + cfg.Server.AppVersion)
		},
	}
}
