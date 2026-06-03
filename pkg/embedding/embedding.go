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
	Input  string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func (c *Client) Embed(text string) ([]float64, error) {
	reqBody := EmbedRequest{Model: c.model, Input: text}
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
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("embedding.empty_result")
	}
	return result.Embeddings[0], nil
}

func (c *Client) EmbedBatch(texts []string) ([][]float64, error) {
	reqBody := EmbedRequest{Model: c.model, Input: texts[0]} // Ollama handles batched input as single string
	if len(texts) > 1 {
		// One by one for Ollama
	}
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
	return result.Embeddings, nil
}
