package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Latentia Agent - AI-Powered SQL Optimization",
	Long: `Latentia Agent is an intelligent system that monitors your TiDB database 
for slow queries and provides AI-powered optimization suggestions.

The agent can run as a server to provide a web interface, or be used via 
CLI commands to generate test data and analyze slow queries.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}