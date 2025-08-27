package cmd

import (
	"fmt"
	"strings"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/ingest"
	"github.com/spf13/cobra"
)

var (
	ingestMinTime float64
	ingestLimit   int
)

var ingestCmd = &cobra.Command{
	Use:   "ingest-slow",
	Short: "Ingest slow queries from INFORMATION_SCHEMA.SLOW_QUERY",
	Long: `Ingest slow queries from TiDB's INFORMATION_SCHEMA.SLOW_QUERY table
into the app_slow_queries table for analysis.

This command is designed for on-premise TiDB installations where
INFORMATION_SCHEMA.SLOW_QUERY is accessible. TiDB Serverless users
should use the generate-slow command with --record flag instead.`,
	RunE: ingestSlowQueries,
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	
	ingestCmd.Flags().Float64Var(&ingestMinTime, "min-time", 0.1, "Minimum query time in seconds to ingest")
	ingestCmd.Flags().IntVar(&ingestLimit, "limit", 100, "Maximum number of slow queries to ingest")
}

func ingestSlowQueries(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔄 Ingesting slow queries from INFORMATION_SCHEMA.SLOW_QUERY...\n")
	fmt.Printf("   Min time: %.1fs, Limit: %d\n", ingestMinTime, ingestLimit)
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	ingester := ingest.NewSlowQueryIngester(db)
	
	err = ingester.IngestFromInformationSchema(ingestMinTime, ingestLimit)
	if err != nil {
		return fmt.Errorf("failed to ingest slow queries: %w", err)
	}
	
	// Show summary of ingested queries
	queries, err := ingester.GetSlowQueries("pending", 10)
	if err != nil {
		return fmt.Errorf("failed to get ingested queries: %w", err)
	}
	
	fmt.Printf("\n📋 Successfully ingested slow queries!\n")
	fmt.Printf("🔍 Recent slow queries (showing last %d):\n", len(queries))
	
	for i, q := range queries {
		fmt.Printf("   %d. [%s] %.3fs - %s\n", i+1, q.Source, q.QueryTime, truncateSQL(q.SampleSQL, 60))
	}
	
	if len(queries) > 0 {
		fmt.Printf("\n💡 Use 'agent run' to start the optimization engine\n")
	} else {
		fmt.Printf("\n💡 No slow queries found. Try generating some with 'agent generate-slow'\n")
	}
	
	return nil
}

func truncateSQL(sql string, maxLen int) string {
	// Clean up whitespace
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	for strings.Contains(sql, "  ") {
		sql = strings.ReplaceAll(sql, "  ", " ")
	}
	
	if len(sql) <= maxLen {
		return sql
	}
	
	return sql[:maxLen] + "..."
}