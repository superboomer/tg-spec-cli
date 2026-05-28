package commands

import (
	"fmt"
	"strings"

	"github.com/superboomer/tg-spec-cli/internal/app"
	"github.com/superboomer/tg-spec-cli/internal/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	outputPath string
	logLevel   string
	url        string
	typeFlag   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the OpenAPI specification",
	Run: func(cmd *cobra.Command, _ []string) {
		log, err := logger.New(logLevel)
		if err != nil {
			fmt.Printf("failed to create logger: %v\n", err)
			return
		}
		defer func(log *zap.Logger) {
			// Syncing a console (stderr) on some platforms returns a harmless
			// error ("inappropriate ioctl for device" on Linux, "bad file
			// descriptor" on macOS); ignore those and only report real failures.
			if err := log.Sync(); err != nil &&
				!strings.Contains(err.Error(), "inappropriate ioctl for device") &&
				!strings.Contains(err.Error(), "bad file descriptor") {
				fmt.Printf("failed to sync logger: %v\n", err)
			}
		}(log)

		// Set defaults for type
		if typeFlag == "gateway" {
			if !cmd.Flags().Changed("url") {
				url = "https://core.telegram.org/gateway/api"
			}
		}

		a := app.NewWithType(log, url, outputPath, typeFlag)
		if err := a.Run(); err != nil {
			log.Fatal("failed to run app", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Output path for the OpenAPI specification. You can specify a directory or a full file path. If the path contains '%v', it will be replaced with the API version (e.g., './specs/bot-api-%v.json'). If a directory does not exist, it will be created automatically.")
	generateCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error, fatal)")
	generateCmd.Flags().StringVarP(&url, "url", "u", "https://core.telegram.org/bots/api", "URL of the Telegram Bot API documentation")
	generateCmd.Flags().StringVarP(&typeFlag, "type", "t", "botapi", "API type: 'botapi' (default) or 'gateway'. For 'gateway', uses https://core.telegram.org/gateway/api and different OpenAPI info/auth.")
}
