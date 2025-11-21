package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

type OllamaProvider struct {
	config semantic.OllamaConfig
}

func NewOllamaProvider(config semantic.OllamaConfig) *OllamaProvider {
	return &OllamaProvider{config: config}
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (p *OllamaProvider) Embed(texts []string) ([][]float32, error) {
	var embeddings [][]float32

	for _, text := range texts {
		reqBody := OllamaRequest{
			Model:  p.config.Model,
			Prompt: text,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(p.config.APIURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("ollama api error: %s", resp.Status)
		}

		var result OllamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		float32Embedding := make([]float32, len(result.Embedding))
		for i, v := range result.Embedding {
			float32Embedding[i] = float32(v)
		}
		embeddings = append(embeddings, float32Embedding)
	}

	return embeddings, nil
}
