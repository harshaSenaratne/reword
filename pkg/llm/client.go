package llm

import (
    "context"
    "fmt"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/harshaSenaratne/reword/internal/config"
)

type Client struct {
    assistantLLM llms.Model
    moderatorLLM llms.Model
    config       *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
    // Create assistant LLM
    assistantLLM, err := openai.New(
        openai.WithToken(cfg.OpenAIAPIKey),
        openai.WithModel(cfg.AssistantModel),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create assistant LLM: %w", err)
    }

    // Create moderator LLM
    moderatorLLM, err := openai.New(
        openai.WithToken(cfg.OpenAIAPIKey),
        openai.WithModel(cfg.ModeratorModel),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create moderator LLM: %w", err)
    }

    return &Client{
        assistantLLM: assistantLLM,
        moderatorLLM: moderatorLLM,
        config:       cfg,
    }, nil
}

func (c *Client) GetAssistantLLM() llms.Model {
    return c.assistantLLM
}

func (c *Client) GetModeratorLLM() llms.Model {
    return c.moderatorLLM
}

func (c *Client) GenerateResponse(ctx context.Context, llm llms.Model, prompt string) (string, error) {
    response, err := llms.GenerateFromSinglePrompt(
        ctx,
        llm,
        prompt,
        llms.WithMaxTokens(c.config.MaxTokens),
        llms.WithTemperature(c.config.Temperature),
    )
    if err != nil {
        return "", fmt.Errorf("failed to generate response: %w", err)
    }
    
    return response, nil
}