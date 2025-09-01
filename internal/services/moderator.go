package services

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/sirupsen/logrus"
    "github.com/harshaSenaratne/reword/pkg/llm"
)

type ModeratorService struct {
    llmClient *llm.Client
    logger    *logrus.Logger
}

func NewModeratorService(llmClient *llm.Client, logger *logrus.Logger) *ModeratorService {
    return &ModeratorService{
        llmClient: llmClient,
        logger:    logger,
    }
}

// cleans up inappropriate content
func (s *ModeratorService) ModerateComment(ctx context.Context, comment string) (string, bool, error) {
    prompt := s.buildModerationPrompt(comment)
    
    s.logger.WithField("comment", comment).Debug("Moderating comment")

    response, err := s.llmClient.GenerateResponse(ctx, s.llmClient.GetModeratorLLM(), prompt)
    if err != nil {
        s.logger.WithError(err).Error("Failed to moderate comment")
        return "", false, fmt.Errorf("moderation failed: %w", err)
    }

    moderatedComment := strings.TrimSpace(response)
    wasModified := moderatedComment != comment

    s.logger.WithFields(logrus.Fields{
        "original":     comment,
        "moderated":    moderatedComment,
        "was_modified": wasModified,
    }).Debug("Comment moderated")

    return moderatedComment, wasModified, nil
}

func (s *ModeratorService) buildModerationPrompt(comment string) string {
    template := `You are the moderator of an online forum. You are strict and will not tolerate any negative, offensive, or inappropriate comments.

Your task:
1. Review the original comment
2. If it contains any rudeness, profanity, negativity, or inappropriate content, transform it to be polite and constructive while maintaining the core meaning
3. If it's already polite and appropriate, return it exactly as is

Guidelines:
- Preserve the intent and information
- Remove all profanity and offensive language
- Transform negative tone to constructive feedback
- Maintain professional language

Original comment: "%s"

Moderated comment:`

    return fmt.Sprintf(template, comment)
}

//  checks if a comment is toxic
func (s *ModeratorService) CheckToxicity(ctx context.Context, comment string) (bool, string, error) {
    prompt := fmt.Sprintf(`Analyze if the following comment contains toxicity, rudeness, or inappropriate content.
    Respond with "YES" or "NO" followed by a brief reason.
    
    Comment: "%s"
    
    Analysis:`, comment)

    response, err := s.llmClient.GenerateResponse(ctx, s.llmClient.GetModeratorLLM(), prompt)
    if err != nil {
        return false, "", err
    }

    response = strings.TrimSpace(response)
    isToxic := strings.HasPrefix(strings.ToUpper(response), "YES")
    
    reason := ""
    if parts := strings.SplitN(response, " ", 2); len(parts) > 1 {
        reason = parts[1]
    }

    return isToxic, reason, nil
}