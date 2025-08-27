package database

const AppSlowQueriesSQL = `
-- App slow queries table - compatible with both generated and INFORMATION_SCHEMA data
CREATE TABLE IF NOT EXISTS app_slow_queries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    digest VARCHAR(64) NOT NULL,
    sample_sql TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    query_time DOUBLE NOT NULL,
    db VARCHAR(64),
    index_names TEXT,
    is_internal BOOLEAN DEFAULT FALSE,
    user VARCHAR(64),
    host VARCHAR(64),
    tables JSON,
    source ENUM('generated', 'information_schema') NOT NULL,
    status ENUM('pending', 'analyzing', 'completed') DEFAULT 'pending',
    last_analyzed_at TIMESTAMP NULL,
    best_rewrite_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_digest (digest),
    INDEX idx_started_at (started_at),
    INDEX idx_query_time (query_time),
    INDEX idx_source_status (source, status),
    INDEX idx_db (db)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Documents table for RAG
CREATE TABLE IF NOT EXISTS app_documents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(500) NOT NULL,
    content LONGTEXT NOT NULL,
    category VARCHAR(100),
    url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    UNIQUE KEY uk_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Embeddings table for vector search
CREATE TABLE IF NOT EXISTS app_embeddings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    doc_id BIGINT NOT NULL,
    chunk_id INT NOT NULL,
    text TEXT NOT NULL,
    embedding VECTOR(1536) NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (doc_id) REFERENCES app_documents(id),
    VECTOR INDEX vec_idx (embedding),
    INDEX idx_doc_chunk (doc_id, chunk_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

const TestSchemaSQL = `
-- Customers table
CREATE TABLE IF NOT EXISTS customers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    company VARCHAR(200),
    city VARCHAR(100),
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_city (city),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock_qty INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_price (price),
    INDEX idx_created_at (created_at)
    -- Intentionally missing index on name for slow query testing
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    customer_id BIGINT NOT NULL,
    status ENUM('pending', 'paid', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
    total DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    shipped_at TIMESTAMP NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_customer_id (customer_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
    -- Intentionally missing index on total for slow query testing
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    INDEX idx_order_id (order_id),
    INDEX idx_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

// SetupTestSchema creates the test tables
func (db *DB) SetupTestSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS app_slow_queries (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    digest VARCHAR(64) NOT NULL,
		    sample_sql TEXT NOT NULL,
		    started_at TIMESTAMP NOT NULL,
		    query_time DOUBLE NOT NULL,
		    db VARCHAR(64),
		    index_names TEXT,
		    is_internal BOOLEAN DEFAULT FALSE,
		    user VARCHAR(64),
		    host VARCHAR(64),
		    tables JSON,
		    source ENUM('generated', 'information_schema') NOT NULL,
		    status ENUM('pending', 'analyzing', 'completed') DEFAULT 'pending',
		    last_analyzed_at TIMESTAMP NULL,
		    best_rewrite_id BIGINT NULL,
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    INDEX idx_digest (digest),
		    INDEX idx_started_at (started_at),
		    INDEX idx_query_time (query_time),
		    INDEX idx_source_status (source, status),
		    INDEX idx_db (db)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS app_documents (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    title VARCHAR(500) NOT NULL,
		    content LONGTEXT NOT NULL,
		    category VARCHAR(100),
		    url VARCHAR(512),
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    INDEX idx_category (category),
		    UNIQUE KEY uk_title (title)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS app_embeddings (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    doc_id BIGINT NOT NULL,
		    chunk_id INT NOT NULL,
		    text TEXT NOT NULL,
		    embedding VECTOR(1536) NOT NULL,
		    metadata JSON,
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    FOREIGN KEY (doc_id) REFERENCES app_documents(id),
		    VECTOR INDEX vec_idx ((VEC_COSINE_DISTANCE(embedding))),
		    INDEX idx_doc_chunk (doc_id, chunk_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS customers (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    email VARCHAR(255) NOT NULL,
		    first_name VARCHAR(100) NOT NULL,
		    last_name VARCHAR(100) NOT NULL,
		    company VARCHAR(200),
		    city VARCHAR(100),
		    country VARCHAR(100),
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    INDEX idx_email (email),
		    INDEX idx_city (city),
		    INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS products (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    name VARCHAR(255) NOT NULL,
		    description TEXT,
		    category VARCHAR(100) NOT NULL,
		    price DECIMAL(10,2) NOT NULL,
		    stock_qty INT NOT NULL DEFAULT 0,
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    INDEX idx_category (category),
		    INDEX idx_price (price),
		    INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS orders (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    customer_id BIGINT NOT NULL,
		    status ENUM('pending', 'paid', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
		    total DECIMAL(10,2) NOT NULL,
		    notes TEXT,
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    shipped_at TIMESTAMP NULL,
		    FOREIGN KEY (customer_id) REFERENCES customers(id),
		    INDEX idx_customer_id (customer_id),
		    INDEX idx_status (status),
		    INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		
		`CREATE TABLE IF NOT EXISTS order_items (
		    id BIGINT PRIMARY KEY AUTO_INCREMENT,
		    order_id BIGINT NOT NULL,
		    product_id BIGINT NOT NULL,
		    quantity INT NOT NULL,
		    price DECIMAL(10,2) NOT NULL,
		    FOREIGN KEY (order_id) REFERENCES orders(id),
		    FOREIGN KEY (product_id) REFERENCES products(id),
		    INDEX idx_order_id (order_id),
		    INDEX idx_product_id (product_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	
	return nil
}

// CleanupTestData removes all test data (but keeps schema)
func (db *DB) CleanupTestData() error {
	queries := []string{
		"DELETE FROM order_items",
		"DELETE FROM orders", 
		"DELETE FROM products",
		"DELETE FROM customers",
	}
	
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	
	return nil
}

// DropTestSchema removes all test tables
func (db *DB) DropTestSchema() error {
	queries := []string{
		"DROP TABLE IF EXISTS order_items",
		"DROP TABLE IF EXISTS orders",
		"DROP TABLE IF EXISTS products", 
		"DROP TABLE IF EXISTS customers",
	}
	
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	
	return nil
}