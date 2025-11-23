package chat

import (
	"time"
)

// Conversation 对话
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message 消息
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"` // user, assistant, system
	Content        string    `json:"content"`
	Thinking       string    `json:"thinking,omitempty"`
	Metadata       string    `json:"metadata,omitempty"` // JSON
	CreatedAt      time.Time `json:"created_at"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider string       `mapstructure:"provider"`
	Ollama   OllamaConfig `mapstructure:"ollama"`
	OpenAI   OpenAIConfig `mapstructure:"openai"`
}

// OllamaConfig Ollama 配置
type OllamaConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Model     string `mapstructure:"model"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// ChatConfig 聊天配置
type ChatConfig struct {
	DBPath string `mapstructure:"db_path"`
}
