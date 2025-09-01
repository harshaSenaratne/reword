
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp" 
    "github.com/sirupsen/logrus"
	"github.com/harshaSenaratne/reword/internal/config"
    "github.com/harshaSenaratne/reword/internal/handlers"
    "github.com/harshaSenaratne/reword/internal/middleware"
    "github.com/harshaSenaratne/reword/internal/services"
    "github.com/harshaSenaratne/reword/pkg/llm"
)

func main() {
    // Initialize logger
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.WithError(err).Fatal("Failed to load configuration")
    }
    
    // Set log level
    level, err := logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        level = logrus.InfoLevel
    }
    logger.SetLevel(level)
    
    // Initialize LLM client
    llmClient, err := llm.NewClient(cfg)
    if err != nil {
        logger.WithError(err).Fatal("Failed to initialize LLM client")
    }
    
    // Initialize services
    assistantService := services.NewAssistantService(llmClient, logger)
    moderatorService := services.NewModeratorService(llmClient, logger)
    chainService := services.NewChainService(assistantService, moderatorService, logger)
    
    // Initialize handlers
    moderatorHandler := handlers.NewModeratorHandler(chainService, logger)
    
    // Setup Gin router
    if cfg.LogLevel != "debug" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    router := gin.New()
    
    // Apply middleware
    router.Use(middleware.LoggingMiddleware(logger))
    router.Use(middleware.ErrorHandlingMiddleware())
    router.Use(middleware.CORSMiddleware())
    router.Use(gin.Recovery())
    
    // API routes
    api := router.Group("/api/v1")
    {
        // Apply rate limiting to API routes
        api.Use(middleware.RateLimitMiddleware(cfg.RateLimitPerMin))
        
        api.POST("/moderate", moderatorHandler.ProcessComment)
        api.POST("/moderate/batch", moderatorHandler.ProcessBatch)
    }
    
    // Health check
    router.GET("/health", moderatorHandler.Health)
    
    // Metrics endpoint (if enabled)
    if cfg.EnableMetrics {
        router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }
    
    // Create HTTP server
    srv := &http.Server{
        Addr:           fmt.Sprintf(":%d", cfg.ServerPort),
        Handler:        router,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   cfg.RequestTimeout,
        MaxHeaderBytes: 1 << 20,
    }
    
    // Start server in goroutine
    go func() {
        logger.WithField("port", cfg.ServerPort).Info("Starting server")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.WithError(err).Fatal("Failed to start server")
        }
    }()
    
    // Wait for interrupt signal to gracefully shutdown the server
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        logger.WithError(err).Fatal("Server forced to shutdown")
    }
    
    logger.Info("Server shutdown complete")
}