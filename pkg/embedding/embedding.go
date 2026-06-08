package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	model   string
	httpCli *http.Client
}

func New(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpCli: &http.Client{Timeout: 30 * time.Second},
	}
}

type EmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *Client) Embed(text string) ([]float64, error) {
	reqBody := EmbedRequest{Model: c.model, Prompt: text}
	body, _ := json.Marshal(reqBody)

	resp, err := c.httpCli.Post(c.baseURL+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding.error: %w", err)
	}
	defer resp.Body.Close()

	var result EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embedding.decode_error: %w", err)
	}
	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("embedding.empty_result")
	}
	return result.Embedding, nil
}

func (c *Client) EmbedBatch(texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, text := range texts {
		vec, err := c.Embed(text)
		if err != nil {
			return nil, err
		}
		result[i] = vec
	}
	return result, nil
}
