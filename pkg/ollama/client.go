package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	httpCli *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpCli: &http.Client{Timeout: 120 * time.Second},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

type ChatResponse struct {
	Message Message `json:"message"`
}

func (c *Client) Chat(model string, messages []Message, temperature float64, maxTokens int) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options: map[string]interface{}{
			"temperature": temperature,
			"num_predict":  maxTokens,
		},
	}

	body, _ := json.Marshal(req)
	resp, err := c.httpCli.Post(c.baseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama.chat_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama.status_error: %d", resp.StatusCode)
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama.decode_error: %w", err)
	}
	return &result, nil
}
