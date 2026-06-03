package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/pkg/langfuse"
	"github.com/hanasakis/kotoha/pkg/ollama"
)

type Service struct {
	agentSvc   *agent.Service
	langfuseCli *langfuse.Client
	model      string
}

func NewService(agentSvc *agent.Service, langfuseCli *langfuse.Client, model string) *Service {
	return &Service{
		agentSvc:    agentSvc,
		langfuseCli: langfuseCli,
		model:       model,
	}
}

func (s *Service) Chat(ctx context.Context, userID uint, history []ollama.Message, input agent.ChatInput) (*agent.ChatOutput, error) {
	traceID := fmt.Sprintf("trace-%d-%d", userID, time.Now().UnixNano())
	s.langfuseCli.CreateTrace(traceID, "agent-chat", fmt.Sprintf("%d", userID), map[string]interface{}{
		"user_id": userID,
	})

	genID := fmt.Sprintf("gen-%s", traceID)
	s.langfuseCli.CreateGeneration(genID, traceID, "llm-call", s.model,
		[]interface{}{map[string]string{"role": "user", "content": input.Message}},
		"")

	result, err := s.agentSvc.Chat(ctx, userID, history, input)

	if err == nil {
		s.langfuseCli.CreateGeneration(genID, traceID, "llm-response", s.model,
			[]interface{}{map[string]string{"role": "assistant", "content": result.Reply}},
			result.Reply)
	} else {
		s.langfuseCli.CreateSpan(fmt.Sprintf("err-%s", traceID), traceID,
			"error", input.Message, err.Error())
	}

	s.langfuseCli.Flush()
	return result, err
}

type EvalMetric struct {
	Category      string  `json:"category"`
	TotalSearches int     `json:"total_searches"`
	AvgRelevance  float64 `json:"avg_relevance"`
	CartAdds      int     `json:"cart_adds"`
	OrdersCreated int     `json:"orders_created"`
}

func (s *Service) GetMetrics() []EvalMetric {
	// Stub metrics for dashboard
	return []EvalMetric{
		{Category: "search", TotalSearches: 0, AvgRelevance: 0},
		{Category: "cart", CartAdds: 0},
		{Category: "order", OrdersCreated: 0},
	}
}
