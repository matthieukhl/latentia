package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matthieukhl/latentia/internal/database"
)

type Server struct {
	router *gin.Engine
	db     *database.DB
}

// NewServer creates a new server instance
func NewServer(db *database.DB) *Server {
	router := gin.Default()
	
	server := &Server{
		router: router,
		db:     db,
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		api.GET("/health", s.healthCheck)
	}
}

// healthCheck endpoint for monitoring
func (s *Server) healthCheck(c *gin.Context) {
	// Check database health
	if err := s.db.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "database connection failed",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "latentia",
		"version": "0.1.0",
	})
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}