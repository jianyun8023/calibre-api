package chat

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// NewLLM 创建 LangChainGo LLM 客户端
func NewLLM(cfg LLMConfig) (llms.Model, error) {
	switch cfg.Provider {
	case "ollama":
		return ollama.New(
			ollama.WithServerURL(cfg.Ollama.ServerURL),
			ollama.WithModel(cfg.Ollama.Model),
		)
	case "openai":
		opts := []openai.Option{
			openai.WithModel(cfg.OpenAI.Model),
		}
		if cfg.OpenAI.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.OpenAI.APIKey))
		}
		if cfg.OpenAI.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.OpenAI.BaseURL))
		}
		return openai.New(opts...)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// GenerateCompletion 生成文本补全
func GenerateCompletion(ctx context.Context, llm llms.Model, prompt string) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate completion: %w", err)
	}
	return completion, nil
}

// GenerateCompletionStream 生成流式文本补全
func GenerateCompletionStream(ctx context.Context, llm llms.Model, prompt string, callback func(string) error) error {
	_, err := llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		return callback(string(chunk))
	}))
	return err
}
