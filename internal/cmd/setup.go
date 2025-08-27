package cmd

import (
	"fmt"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/spf13/cobra"
)

var (
	dropFirst bool
	skipData  bool
)

var setupCmd = &cobra.Command{
	Use:   "setup-test-data",
	Short: "Set up test database schema and sample data",
	Long: `Creates test tables (customers, orders, products, order_items) 
and populates them with sample data for slow query testing.

This creates realistic e-commerce data that can be used to generate
various types of slow queries for testing the optimization engine.`,
	RunE: setupTestData,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	
	setupCmd.Flags().BoolVar(&dropFirst, "drop-first", false, "Drop existing test tables before creating")
	setupCmd.Flags().BoolVar(&skipData, "schema-only", false, "Create schema only, skip sample data")
}

func setupTestData(cmd *cobra.Command, args []string) error {
	fmt.Println("🔧 Setting up test database...")
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	// Drop tables if requested
	if dropFirst {
		fmt.Println("🗑️  Dropping existing test tables...")
		if err := db.DropTestSchema(); err != nil {
			return fmt.Errorf("failed to drop test schema: %w", err)
		}
	}
	
	// Create schema
	fmt.Println("📋 Creating test schema...")
	if err := db.SetupTestSchema(); err != nil {
		return fmt.Errorf("failed to setup test schema: %w", err)
	}
	
	if !skipData {
		fmt.Println("📊 Populating with sample data...")
		if err := populateSampleData(db); err != nil {
			return fmt.Errorf("failed to populate sample data: %w", err)
		}
	}
	
	fmt.Println("✅ Test database setup complete!")
	return nil
}

func populateSampleData(db *database.DB) error {
	fmt.Println("   👥 Creating customers...")
	if err := createCustomers(db); err != nil {
		return err
	}
	
	fmt.Println("   📦 Creating products...")
	if err := createProducts(db); err != nil {
		return err
	}
	
	fmt.Println("   🛒 Creating orders...")
	if err := createOrders(db); err != nil {
		return err
	}
	
	fmt.Println("   📋 Creating order items...")
	if err := createOrderItems(db); err != nil {
		return err
	}
	
	return nil
}

func createCustomers(db *database.DB) error {
	customers := []struct {
		email, firstName, lastName, company, city, country string
	}{
		{"john.doe@email.com", "John", "Doe", "Tech Corp", "New York", "USA"},
		{"jane.smith@gmail.com", "Jane", "Smith", "Design Studio", "London", "UK"},
		{"bob.wilson@yahoo.com", "Bob", "Wilson", "Wilson LLC", "Toronto", "Canada"},
		{"alice.brown@hotmail.com", "Alice", "Brown", "Brown Industries", "Sydney", "Australia"},
		{"charlie.davis@outlook.com", "Charlie", "Davis", "Davis & Co", "Berlin", "Germany"},
		{"diana.miller@company.com", "Diana", "Miller", "Miller Solutions", "Tokyo", "Japan"},
		{"frank.garcia@startup.io", "Frank", "Garcia", "Garcia Tech", "San Francisco", "USA"},
		{"grace.lee@enterprise.com", "Grace", "Lee", "Lee Enterprises", "Seoul", "South Korea"},
		{"henry.taylor@business.net", "Henry", "Taylor", "Taylor Group", "Paris", "France"},
		{"ivy.anderson@firm.org", "Ivy", "Anderson", "Anderson Firm", "Stockholm", "Sweden"},
	}
	
	for _, c := range customers {
		_, err := db.Exec(`
			INSERT INTO customers (email, first_name, last_name, company, city, country, created_at)
			VALUES (?, ?, ?, ?, ?, ?, DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 365) DAY))
		`, c.email, c.firstName, c.lastName, c.company, c.city, c.country)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func createProducts(db *database.DB) error {
	products := []struct {
		name, description, category string
		price                       float64
		stockQty                    int
	}{
		{"Laptop Pro 15\"", "High-performance laptop for professionals", "electronics", 1299.99, 50},
		{"Wireless Mouse", "Ergonomic wireless mouse with USB receiver", "electronics", 29.99, 200},
		{"Programming Book", "Complete guide to modern software development", "books", 49.99, 100},
		{"Cotton T-Shirt", "Premium cotton t-shirt, multiple sizes", "clothing", 19.99, 500},
		{"Running Shoes", "Professional running shoes for athletes", "sports", 89.99, 150},
		{"Coffee Mug", "Ceramic coffee mug with company logo", "home", 9.99, 300},
		{"Smartphone Case", "Protective case for latest smartphone models", "electronics", 24.99, 400},
		{"Cookbook Collection", "Collection of international recipes", "books", 34.99, 75},
		{"Winter Jacket", "Warm winter jacket, waterproof material", "clothing", 129.99, 80},
		{"Yoga Mat", "Non-slip yoga mat for home workouts", "sports", 39.99, 120},
		{"LED Desk Lamp", "Adjustable LED lamp for office use", "home", 59.99, 60},
		{"Tablet Stand", "Adjustable stand for tablets and phones", "electronics", 19.99, 180},
		{"Mystery Novel", "Bestselling mystery novel by famous author", "books", 12.99, 250},
		{"Business Shirt", "Professional dress shirt for business", "clothing", 39.99, 200},
		{"Tennis Racket", "Professional tennis racket for tournaments", "sports", 199.99, 40},
	}
	
	for _, p := range products {
		_, err := db.Exec(`
			INSERT INTO products (name, description, category, price, stock_qty, created_at)
			VALUES (?, ?, ?, ?, ?, DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 180) DAY))
		`, p.name, p.description, p.category, p.price, p.stockQty)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func createOrders(db *database.DB) error {
	statuses := []string{"pending", "paid", "shipped", "delivered", "cancelled"}
	
	// Create 50 orders with random customers and data
	for i := 0; i < 50; i++ {
		customerID := (i % 10) + 1 // Cycle through customer IDs 1-10
		status := statuses[i%len(statuses)]
		total := 50.0 + float64(i*10) // Varying order totals
		notes := fmt.Sprintf("Order #%d - Customer requested special handling", i+1000)
		
		_, err := db.Exec(`
			INSERT INTO orders (customer_id, status, total, notes, created_at)
			VALUES (?, ?, ?, ?, DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 90) DAY))
		`, customerID, status, total, notes)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func createOrderItems(db *database.DB) error {
	// Create 2-4 items per order
	for orderID := 1; orderID <= 50; orderID++ {
		itemCount := 2 + (orderID % 3) // 2-4 items per order
		
		for item := 0; item < itemCount; item++ {
			productID := (item*3 + orderID) % 15 + 1 // Distribute across products
			quantity := 1 + (item % 3)              // 1-3 quantity
			price := 10.0 + float64(productID*5)    // Varying prices
			
			_, err := db.Exec(`
				INSERT INTO order_items (order_id, product_id, quantity, price)
				VALUES (?, ?, ?, ?)
			`, orderID, productID, quantity, price)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}