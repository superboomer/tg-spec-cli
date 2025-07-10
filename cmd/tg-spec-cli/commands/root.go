package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tg-bot-api-gen",
	Short: "A CLI tool to generate OpenAPI spec for Telegram Bot API",
	Long:  `A CLI tool to generate OpenAPI spec for Telegram Bot API.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
