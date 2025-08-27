package main

import (
	"fmt"
	"os"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/server"
)

func main() {
	fmt.Println("Latentia Agent Starting...")
	
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent <command>")
		fmt.Println("Commands:")
		fmt.Println("  run    - Start the server")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "run":
		fmt.Println("Loading configuration...")
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Connecting to database...")
		db, err := database.NewConnection(&cfg.DB)
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		
		fmt.Printf("Database connected successfully\n")
		
		fmt.Println("Setting up server...")
		srv := server.NewServer(db)
		
		fmt.Printf("Starting server on %s...\n", cfg.Server.Addr)
		if err := srv.Start(cfg.Server.Addr); err != nil {
			fmt.Printf("Server failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}