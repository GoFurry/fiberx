package cmd

import (
	"fmt"
	"os"

	env "github.com/gofurry/fiberx/v3/extra-light/config"
	"github.com/gofurry/fiberx/v3/extra-light/internal/bootstrap"
	apphttp "github.com/gofurry/fiberx/v3/extra-light/internal/http"
	"github.com/gofurry/fiberx/v3/extra-light/pkg/common"
)

func Execute(args []string) int {
	command, configPath, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		printUsage()
		return 1
	}

	switch command {
	case "serve":
		if err := env.InitServerConfig(configPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := bootstrap.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := apphttp.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	case "version":
		appName := common.COMMON_PROJECT_NAME
		if err := env.InitServerConfig(configPath); err == nil {
			if name := env.GetServerConfig().Server.AppName; name != "" {
				appName = name
			}
		}
		fmt.Printf("%s %s\n", appName, common.APP_VERSION)
		return 0
	default:
		printUsage()
		return 1
	}
}

func parseArgs(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("missing command")
	}

	command := args[0]
	configPath := ""

	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--config", "-config":
			index++
			if index >= len(args) {
				return "", "", fmt.Errorf("missing value for --config")
			}
			configPath = args[index]
		default:
			return "", "", fmt.Errorf("unknown argument: %s", args[index])
		}
	}

	return command, configPath, nil
}

func printUsage() {
	fmt.Print(common.COMMON_PROJECT_HELP)
}
