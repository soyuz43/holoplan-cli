package cmd

import (
	"log"

	"holoplan-cli/src/runner"

	"github.com/spf13/cobra"
)

var storyPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Process YAML user stories into draw.io layout",
	Run: func(cmd *cobra.Command, args []string) {
		if storyPath == "" {
			log.Fatalf("❌ Please provide a path to a YAML file using --stories")
		}
		runner.RunPipeline(storyPath)
	},
}

func init() {
	runCmd.Flags().StringVar(&storyPath, "stories", "", "Path to user stories YAML")
	rootCmd.AddCommand(runCmd)
}
