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
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function FunctionDef  `json:"function"`
}

type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string           `json:"id,omitempty"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ChatRequest struct {
	Model     string                 `json:"model"`
	Messages  []Message              `json:"messages"`
	Stream    bool                   `json:"stream"`
	Tools     []ToolDef              `json:"tools,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

type ChatResponse struct {
	Message         Message `json:"message"`
	EvalCount       int     `json:"eval_count"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	TotalDuration   int64   `json:"total_duration"`
}

func (r *ChatResponse) InputTokens() int  { return r.PromptEvalCount }
func (r *ChatResponse) OutputTokens() int { return r.EvalCount }

func (c *Client) Chat(model string, messages []Message, temperature float64, maxTokens int, tools []ToolDef) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Tools:    tools,
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
