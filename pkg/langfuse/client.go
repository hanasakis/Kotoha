package langfuse

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	klog "github.com/hanasakis/kotoha/pkg/log"
)

type Client struct {
	baseURL string
	auth    string
	httpCli *http.Client
	mu      sync.Mutex
	events  []event
}

type event struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Body      interface{} `json:"body"`
}

func New(host, publicKey, secretKey string) *Client {
	host = strings.TrimRight(host, "/")
	return &Client{
		baseURL: host + "/api/public/ingestion",
		auth:    base64.StdEncoding.EncodeToString([]byte(publicKey + ":" + secretKey)),
		httpCli: &http.Client{Timeout: 15 * time.Second},
	}
}

type TraceBody struct {
	Name      string                 `json:"name"`
	UserID    string                 `json:"userId,omitempty"`
	SessionID string                 `json:"sessionId,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Input     interface{}            `json:"input,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ObservationBody struct {
	ID            string                 `json:"id,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
	Name          string                 `json:"name"`
	ParentObsID   string                 `json:"parentObservationId,omitempty"`
	StartTime     string                 `json:"startTime,omitempty"`
	EndTime       string                 `json:"endTime,omitempty"`
	Input         interface{}            `json:"input,omitempty"`
	Output        interface{}            `json:"output,omitempty"`
	Level         string                 `json:"level,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type GenerationBody struct {
	ID            string                 `json:"id,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
	Name          string                 `json:"name"`
	ParentObsID   string                 `json:"parentObservationId,omitempty"`
	StartTime     string                 `json:"startTime,omitempty"`
	EndTime       string                 `json:"endTime,omitempty"`
	Model         string                 `json:"model,omitempty"`
	ModelParams   map[string]interface{} `json:"modelParameters,omitempty"`
	Input         []interface{}          `json:"input,omitempty"`
	Output        interface{}            `json:"output,omitempty"`
	Usage         *Usage                 `json:"usage,omitempty"`
	Level         string                 `json:"level,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type Usage struct {
	Input  int `json:"input,omitempty"`
	Output int `json:"output,omitempty"`
	Total  int `json:"total,omitempty"`
}

type ScoreBody struct {
	Name     string  `json:"name"`
	TraceID  string  `json:"traceId"`
	Value    float64 `json:"value"`
	Comment  string  `json:"comment,omitempty"`
	ObsID    string  `json:"observationId,omitempty"`
	DataType string  `json:"dataType,omitempty"`
}

func newUUID() string { return uuid.New().String() }

// CreateTrace creates a trace event and returns its ID for use in child observations.
func (c *Client) CreateTrace(t *TraceBody) string {
	id := newUUID()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "trace-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Body:      t,
	})
	return id
}

// CreateSpan creates a span observation and returns its ID for use as parent.
func (c *Client) CreateSpan(s *ObservationBody) string {
	id := newUUID()
	if s.ID == "" {
		s.ID = id
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "span-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Body:      s,
	})
	return id
}

// CreateGeneration creates a generation observation and returns its ID.
func (c *Client) CreateGeneration(g *GenerationBody) string {
	id := newUUID()
	if g.ID == "" {
		g.ID = id
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "generation-create",
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Body:      g,
	})
	return id
}

// CreateScore creates a score event.
func (c *Client) CreateScore(s *ScoreBody) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event{
		Type:      "score-create",
		ID:        newUUID(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Body:      s,
	})
}

func (c *Client) Flush() {
	c.mu.Lock()
	events := c.events
	c.events = nil
	c.mu.Unlock()

	if len(events) == 0 {
		return
	}

	body, err := json.Marshal(map[string]interface{}{"batch": events})
	if err != nil {
		klog.Warnf("[langfuse] marshal error: %v", err)
		return
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		klog.Warnf("[langfuse] request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.auth)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		klog.Warnf("[langfuse] flush error: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		klog.Warnf("[langfuse] flush failed: HTTP %d (%d events) body=%s", resp.StatusCode, len(events), string(respBody))
	}
	// Success is silent — events are flowing
}

func (c *Client) FlushAsync() {
	go c.Flush()
}
