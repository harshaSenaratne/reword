package models

import (
    "time"
)

// CommentRequest ( What the client sends)
type CommentRequest struct {
    Comment   string `json:"comment" binding:"required"`
    Sentiment string `json:"sentiment,omitempty"`
    UserID    string `json:"user_id,omitempty"`
}

// ModeratedResponse - What client recieves
type ModeratedResponse struct {
    OriginalComment   string    `json:"original_comment"`
    ModeratedInput    string    `json:"moderated_input,omitempty"` 
    AssistantReply    string    `json:"assistant_reply"`
    WasModified       bool      `json:"was_modified"`                 
    ModerationReason  string    `json:"moderation_reason,omitempty"`
    Timestamp         time.Time `json:"timestamp"`
}

// ModerationContext - Additional context for moderation
type ModerationContext struct {
    UserID          string
    ConversationID  string
    PreviousContext []Message
}

// Message - Single message in conversation
type Message struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}

// HealthResponse - Server health check
type HealthResponse struct {
    Status    string    `json:"status"`
    Version   string    `json:"version"`
    Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse - Standardized error format
type ErrorResponse struct {
    Error   string    `json:"error"`
    Message string    `json:"message"`
    Code    int       `json:"code"`
    TraceID string    `json:"trace_id,omitempty"`
}