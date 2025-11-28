package embedding

import (
	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// Provider defines the interface for embedding providers
type Provider interface {
	Embed(texts []string) ([][]float32, error)
}

// ProviderConfig holds configuration for creating a provider
type ProviderConfig struct {
	Provider    string
	Ollama      semantic.OllamaConfig
	SiliconFlow semantic.SiliconFlowConfig
	VectorDim   int
}
