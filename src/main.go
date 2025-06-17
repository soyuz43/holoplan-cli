// src/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
				fmt.Print("Please provide a filepath for the user stories.yaml: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("[x] Failed to read input:", err)
					os.Exit(1)
				}
				storiesPath = strings.TrimSpace(input)
			}

			if err := runner.RunPipeline(storiesPath); err != nil {
				fmt.Println("[x] Pipeline failed:", err)
				os.Exit(1)
			}

			fmt.Println("[✓] Pipeline completed successfully")
		},
	}

	runCmd.Flags().StringVarP(&storiesPath, "stories", "s", "", "Path to user stories YAML file")
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("[x] Command execution failed:", err)
		os.Exit(1)
	}
}
