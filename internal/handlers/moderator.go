package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/harshaSenaratne/reword/internal/models"
    "github.com/harshaSenaratne/reword/internal/services"
)

type ModeratorHandler struct {
    chainService *services.ChainService
    logger       *logrus.Logger
}

func NewModeratorHandler(chainService *services.ChainService, logger *logrus.Logger) *ModeratorHandler {
    return &ModeratorHandler{
        chainService: chainService,
        logger:       logger,
    }
}

//  handles single comment processing
func (h *ModeratorHandler) ProcessComment(c *gin.Context) {
    var req models.CommentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.WithError(err).Error("Invalid request payload")
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "Invalid Request",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Add user ID from context if available
    if userID, exists := c.Get("user_id"); exists {
        req.UserID = userID.(string)
    }

    ctx := c.Request.Context()
    response, err := h.chainService.ProcessComment(ctx, &req)
    if err != nil {
        h.logger.WithError(err).Error("Failed to process comment")
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "Processing Failed",
            Message: "Failed to process comment",
            Code:    http.StatusInternalServerError,
        })
        return
    }

    c.JSON(http.StatusOK, response)
}

//  handles batch comment processing
func (h *ModeratorHandler) ProcessBatch(c *gin.Context) {
    var requests []*models.CommentRequest
    if err := c.ShouldBindJSON(&requests); err != nil {
        h.logger.WithError(err).Error("Invalid batch request payload")
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "Invalid Request",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    if len(requests) > 10 {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "Batch Too Large",
            Message: "Maximum 10 comments per batch",
            Code:    http.StatusBadRequest,
        })
        return
    }

    ctx := c.Request.Context()
    responses, err := h.chainService.ProcessBatch(ctx, requests)
    if err != nil {
        h.logger.WithError(err).Error("Failed to process batch")
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "Batch Processing Failed",
            Message: "Failed to process batch",
            Code:    http.StatusInternalServerError,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "responses": responses,
        "count":     len(responses),
    })
}

// Health handles health check
func (h *ModeratorHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, models.HealthResponse{
        Status:    "healthy",
        Version:   "1.0.0",
        Timestamp: time.Now(),
    })
}