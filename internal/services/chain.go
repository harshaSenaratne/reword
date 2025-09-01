package services

import (
    "context"
    "fmt"
	"time"
    "github.com/sirupsen/logrus"
    "github.com/harshaSenaratne/reword/internal/models"
)

type ChainService struct {
    assistant *AssistantService
    moderator *ModeratorService
    logger    *logrus.Logger
}

func NewChainService(assistant *AssistantService, moderator *ModeratorService, logger *logrus.Logger) *ChainService {
    return &ChainService{
        assistant: assistant,
        moderator: moderator,
        logger:    logger,
    }
}

// processes a comment through the complete chain
func (s *ChainService) ProcessComment(ctx context.Context, req *models.CommentRequest) (*models.ModeratedResponse, error) {
    startTime := time.Now()
    
    s.logger.WithFields(logrus.Fields{
        "comment":   req.Comment,
        "sentiment": req.Sentiment,
        "user_id":   req.UserID,
    }).Info("Processing comment")

    // Step 1: Check toxicity of incoming comment
    isToxic, toxicityReason, err := s.moderator.CheckToxicity(ctx, req.Comment)
    if err != nil {
        s.logger.WithError(err).Warn("Failed to check toxicity, continuing")
    }

    // Step 2: If comment is toxic, moderate it first
    moderatedInput := req.Comment
    wasModified := false
    if isToxic {
        moderatedInput, wasModified, err = s.moderator.ModerateComment(ctx, req.Comment)
        if err != nil {
            return nil, fmt.Errorf("failed to moderate input comment: %w", err)
        }
        s.logger.WithFields(logrus.Fields{
            "original_input":  req.Comment,
            "moderated_input": moderatedInput,
            "was_modified":    wasModified,
        }).Info("Input comment was moderated")
    }

    // Step 3: Analyze sentiment if not provided - use the moderated input
    sentiment := req.Sentiment
    if sentiment == "" {
        sentiment, err = s.assistant.AnalyzeSentiment(ctx, moderatedInput)
        if err != nil {
            s.logger.WithError(err).Warn("Failed to analyze sentiment, using default")
            sentiment = "helpful"
        }
    }

    // Step 4: Generate assistant response based on the moderated input
    assistantResponse, err := s.assistant.GenerateResponse(ctx, sentiment, moderatedInput)
    if err != nil {
        return nil, fmt.Errorf("failed to generate assistant response: %w", err)
    }

    // Build response
    response := &models.ModeratedResponse{
        OriginalComment:  req.Comment,
        AssistantReply:   assistantResponse,
        WasModified:      wasModified,
        ModerationReason: toxicityReason,
        Timestamp:        time.Now(),
    }

    // Include moderated input only if it was actually modified
    if wasModified {
        response.ModeratedInput = moderatedInput
    }

    s.logger.WithFields(logrus.Fields{
        "processing_time": time.Since(startTime),
        "was_modified":    response.WasModified,
    }).Info("Comment processed successfully")

    return response, nil
}

// processes multiple comments concurrently
func (s *ChainService) ProcessBatch(ctx context.Context, requests []*models.CommentRequest) ([]*models.ModeratedResponse, error) {
    responses := make([]*models.ModeratedResponse, len(requests))
    errChan := make(chan error, len(requests))
    
    for i, req := range requests {
        go func(index int, request *models.CommentRequest) {
            response, err := s.ProcessComment(ctx, request)
            if err != nil {
                errChan <- err
                return
            }
            responses[index] = response
            errChan <- nil
        }(i, req)
    }

    // Wait for all goroutines to complete
    var firstError error
    for i := 0; i < len(requests); i++ {
        if err := <-errChan; err != nil && firstError == nil {
            firstError = err
        }
    }

    if firstError != nil {
        return nil, firstError
    }

    return responses, nil
}