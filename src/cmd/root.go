// src/cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "holoplan",
	Short: "Holoplan CLI - Generate and audit UI layouts from user stories",
	Long:  `Holoplan CLI uses LLMs to chunk, layout, audit, and validate UI design stories into draw.io files.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
}
