package cmd

import (
	"fmt"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/server"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Latentia Agent server",
	Long: `Start the Latentia Agent server which provides:
- REST API for slow query analysis
- Web interface for reviewing optimization suggestions
- Background processing of slow queries`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Latentia Agent Starting...")
	
	fmt.Println("📝 Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	fmt.Println("🔌 Connecting to database...")
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	fmt.Println("✅ Database connected successfully")
	
	fmt.Println("⚙️  Setting up server...")
	srv := server.NewServer(db)
	
	fmt.Printf("🌐 Starting server on %s...\n", cfg.Server.Addr)
	if err := srv.Start(cfg.Server.Addr); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	
	return nil
}