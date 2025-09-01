package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

// logs every request
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        
        c.Next()
        
        logger.WithFields(logrus.Fields{
            "method":      c.Request.Method,     
            "path":        c.Request.URL.Path,   
            "status":      c.Writer.Status(),    
            "latency":     time.Since(startTime), 
            "client_ip":   c.ClientIP(),         
            "user_agent":  c.Request.UserAgent(), 
        }).Info("Request processed")
    }
}

// handles errors consistently
func ErrorHandlingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()  
        
        // Check if any errors occurred
        if len(c.Errors) > 0 {
            err := c.Errors.Last()  
            c.JSON(c.Writer.Status(), gin.H{
                "error":   "Internal Server Error",
                "message": err.Error(),
            })
        }
    }
}

// allows browser-based clients
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Set CORS headers
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")  // Any origin
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        // Handle preflight requests
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)  // No content
            return
        }

        c.Next()
    }
}

func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
    requests := make(map[string][]time.Time)  // IP -> timestamps
    
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        now := time.Now()
        
        // Clean old requests (older than 1 minute)
        if timestamps, exists := requests[clientIP]; exists {
            var valid []time.Time
            for _, t := range timestamps {
                if now.Sub(t) < time.Minute {
                    valid = append(valid, t)  // Keep if < 1 minute old
                }
            }
            requests[clientIP] = valid
        }
        
        // Check if over limit
        if len(requests[clientIP]) >= requestsPerMinute {
            c.JSON(429, gin.H{  // 429 = Too Many Requests
                "error": "Rate limit exceeded",
            })
            c.Abort()  // Stop processing
            return
        }
        
        // Add current request
        requests[clientIP] = append(requests[clientIP], now)
        c.Next()
    }
}