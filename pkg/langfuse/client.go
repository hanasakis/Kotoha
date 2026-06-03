package langfuse

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	baseURL  string
	auth     string
	httpCli  *http.Client
	mu       sync.Mutex
	events   []event
}

type event struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Body      interface{} `json:"body"`
}

func New(host, publicKey, secretKey string) *Client {
	auth := base64.StdEncoding.EncodeToString([]byte(publicKey + ":" + secretKey))
	return &Client{
		baseURL: host + "/api/public/ingestion",
		auth:    auth,
		httpCli: &http.Client{Timeout: 10 * time.Second},
	}
}

type TraceBody struct {
	Name    string            `json:"name"`
	UserID  string            `json:"userId"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SpanBody struct {
	Name     string                 `json:"name"`
	TraceID  string                 `json:"traceId"`
	Input    string                 `json:"input,omitempty"`
	Output   string                 `json:"output,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type GenerationBody struct {
	Name          string                 `json:"name"`
	TraceID       string                 `json:"traceId"`
	Model         string                 `json:"model"`
	ModelParams   map[string]interface{} `json:"modelParameters,omitempty"`
	Input         []interface{}          `json:"input,omitempty"`
	Output        string                 `json:"output,omitempty"`
	Usage         map[string]int         `json:"usage,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func (c *Client) CreateTrace(id, name, userID string, metadata map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "trace-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Body: TraceBody{
			Name:    name,
			UserID:  userID,
			Metadata: metadata,
		},
	})
}

func (c *Client) CreateSpan(id, traceID, name, input, output string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "span-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Body: SpanBody{
			Name:    name,
			TraceID: traceID,
			Input:   input,
			Output:  output,
		},
	})
}

func (c *Client) CreateGeneration(id, traceID, name, model string, input []interface{}, output string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "generation-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Body: GenerationBody{
			Name:    name,
			TraceID: traceID,
			Model:   model,
			Input:   input,
			Output:  output,
		},
	})
}

func (c *Client) Flush() error {
	c.mu.Lock()
	events := c.events
	c.events = nil
	c.mu.Unlock()

	if len(events) == 0 {
		return nil
	}

	body, _ := json.Marshal(map[string]interface{}{"batch": events})
	req, _ := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.auth)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("langfuse.flush_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("langfuse.status_error: %d", resp.StatusCode)
	}
	return nil
}
