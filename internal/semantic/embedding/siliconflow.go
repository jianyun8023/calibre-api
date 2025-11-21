package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

type SiliconFlowProvider struct {
	config semantic.SiliconFlowConfig
}

func NewSiliconFlowProvider(config semantic.SiliconFlowConfig) *SiliconFlowProvider {
	return &SiliconFlowProvider{config: config}
}

type SiliconFlowRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type SiliconFlowResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (p *SiliconFlowProvider) Embed(texts []string) ([][]float32, error) {
	var embeddings [][]float32

	for _, text := range texts {
		reqBody := SiliconFlowRequest{
			Model: p.config.Model,
			Input: text,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", p.config.APIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.config.APIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("siliconflow api error: %s", resp.Status)
		}

		var result SiliconFlowResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		if len(result.Data) == 0 {
			return nil, fmt.Errorf("no embedding data returned")
		}

		float32Embedding := make([]float32, len(result.Data[0].Embedding))
		for i, v := range result.Data[0].Embedding {
			float32Embedding[i] = float32(v)
		}
		embeddings = append(embeddings, float32Embedding)
	}

	return embeddings, nil
}
