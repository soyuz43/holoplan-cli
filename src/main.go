// src/main.go
package main

import (
	"fmt"
	"os"

	"holoplan-cli/src/runner"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "holoplan",
		Short: "Holoplan generates UI wireframes from user stories",
	}

	var storiesPath string

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Generate wireframes from a YAML file of user stories",
		Run: func(cmd *cobra.Command, args []string) {
			if storiesPath == "" {
				fmt.Println("❌ Please provide a path to the YAML file using --stories")
				os.Exit(1)
			}
			if err := runner.RunPipeline(storiesPath); err != nil {
				fmt.Printf("🚨 Pipeline failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("🎉 All done!")
		},
	}

	runCmd.Flags().StringVarP(&storiesPath, "stories", "s", "", "Path to user stories YAML file")
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
