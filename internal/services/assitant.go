package services

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/sirupsen/logrus"
    "github.com/harshaSenaratne/reword/pkg/llm"
)

type AssistantService struct {
    llmClient *llm.Client
    logger    *logrus.Logger
}

func NewAssistantService(llmClient *llm.Client, logger *logrus.Logger) *AssistantService {
    return &AssistantService{
        llmClient: llmClient,
        logger:    logger,
    }
}

//  generates a response based on sentiment and customer request
func (s *AssistantService) GenerateResponse(ctx context.Context, sentiment, customerRequest string) (string, error) {
    // Default sentiment if not provided
    if sentiment == "" {
        sentiment = "helpful and professional"
    }

    prompt := s.buildPrompt(sentiment, customerRequest)
    
    s.logger.WithFields(logrus.Fields{
        "sentiment": sentiment,
        "request":   customerRequest,
    }).Debug("Generating assistant response")

    response, err := s.llmClient.GenerateResponse(ctx, s.llmClient.GetAssistantLLM(), prompt)
    if err != nil {
        s.logger.WithError(err).Error("Failed to generate assistant response")
        return "", fmt.Errorf("assistant response generation failed: %w", err)
    }

    s.logger.WithField("response", response).Debug("Assistant response generated")
    
    return strings.TrimSpace(response), nil
}

func (s *AssistantService) buildPrompt(sentiment, customerRequest string) string {
    template := `You are a %s assistant that responds to user comments, using similar vocabulary as the user.
User: "%s"
Comment:`

    return fmt.Sprintf(template, sentiment, customerRequest)
}

//  analyzes the sentiment of the customer request
func (s *AssistantService) AnalyzeSentiment(ctx context.Context, customerRequest string) (string, error) {
    prompt := fmt.Sprintf(`Analyze the sentiment of the following comment and respond with only one word: 
    positive, negative, or neutral.
    
    Comment: "%s"
    
    Sentiment:`, customerRequest)

    response, err := s.llmClient.GenerateResponse(ctx, s.llmClient.GetAssistantLLM(), prompt)
    if err != nil {
        return "neutral", err
    }

    sentiment := strings.ToLower(strings.TrimSpace(response))
    if sentiment != "positive" && sentiment != "negative" && sentiment != "neutral" {
        sentiment = "neutral"
    }

    return sentiment, nil
}