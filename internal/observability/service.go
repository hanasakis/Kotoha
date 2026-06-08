package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/pkg/langfuse"
	"github.com/hanasakis/kotoha/pkg/ollama"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
)

type Service struct {
	agentSvc    *agent.Service
	langfuseCli *langfuse.Client
	model       string
	env         string
	redisCli    *goredis.Client
	evalRepo    *EvalRepository
}

func NewService(agentSvc *agent.Service, langfuseCli *langfuse.Client, redisCli *goredis.Client, evalRepo *EvalRepository, model, env string) *Service {
	return &Service{
		agentSvc:    agentSvc,
		langfuseCli: langfuseCli,
		model:       model,
		env:         env,
		redisCli:    redisCli,
		evalRepo:    evalRepo,
	}
}

func (s *Service) Chat(ctx context.Context, userID uint, history []ollama.Message, input agent.ChatInput) (*agent.ChatOutput, error) {
	startTime := time.Now()
	result, err := s.agentSvc.Chat(ctx, userID, history, input)
	duration := time.Since(startTime).Milliseconds()

	if s.langfuseCli == nil {
		return result, err
	}

	userIDStr := fmt.Sprintf("%d", userID)
	feature := inferFeature(input.Message)
	sessionID := fmt.Sprintf("session-%d-%d", userID, time.Now().Unix()/3600)
	ts := startTime.UTC().Format(time.RFC3339Nano)

	tags := []string{
		"env:" + s.env,
		"feature:" + feature,
		"intent:" + classifyIntentTag(input.Message),
	}

	traceID := s.langfuseCli.CreateTrace(&langfuse.TraceBody{
		Name:      "agent-chat",
		UserID:    userIDStr,
		SessionID: sessionID,
		Tags:      tags,
		Input:     input.Message,
		Metadata: map[string]interface{}{
			"history_len": len(history),
			"environment": s.env,
		},
	})

	if err != nil {
		s.langfuseCli.CreateSpan(&langfuse.ObservationBody{
			TraceID:       traceID,
			Name:          "agent-error",
			StartTime:     ts,
			Level:         "ERROR",
			StatusMessage: err.Error(),
			Input:         input.Message,
			Metadata:      map[string]interface{}{"duration_ms": duration},
		})
		s.langfuseCli.CreateScore(&langfuse.ScoreBody{
			Name:     "success",
			TraceID:  traceID,
			Value:    0,
			DataType: "BOOLEAN",
			Comment:  err.Error(),
		})
		s.langfuseCli.FlushAsync()
		return result, err
	}

	// Always push a success score on the trace
	s.langfuseCli.CreateScore(&langfuse.ScoreBody{
		Name:     "success",
		TraceID:  traceID,
		Value:    1,
		DataType: "BOOLEAN",
	})

	switch result.Path {
	case "classifier":
		s.createClassifierTrace(traceID, ts, input, result, duration)
	case "llm", "llm-react", "multi-step":
		s.createLLMTrace(traceID, ts, input, result, duration)
	}

	s.langfuseCli.FlushAsync()
	return result, err
}

func (s *Service) createClassifierTrace(traceID, startTime string, input agent.ChatInput, result *agent.ChatOutput, duration int64) {
	// Span 1: intent classification
	classifierSpanID := s.langfuseCli.CreateSpan(&langfuse.ObservationBody{
		TraceID:   traceID,
		Name:      "intent-classification",
		StartTime: startTime,
		Input:     input.Message,
		Output:    map[string]string{"tool": result.ToolUsed, "path": result.Path},
		Metadata: map[string]interface{}{
			"tool_called": result.ToolUsed,
		},
	})

	// Span 2: tool execution (child of classification)
	s.langfuseCli.CreateSpan(&langfuse.ObservationBody{
		TraceID:     traceID,
		ParentObsID: classifierSpanID,
		Name:        "tool-execution:" + result.ToolUsed,
		StartTime:   startTime,
		Input:       result.ToolUsed,
		Output:      result.Reply,
		Metadata: map[string]interface{}{
			"tool":        result.ToolUsed,
			"success":     !strings.HasPrefix(result.Reply, "抱歉"),
			"duration_ms": duration,
		},
	})

	// Score: tool success (1.0 = success, 0.0 = failure)
	successVal := 1.0
	if strings.HasPrefix(result.Reply, "抱歉") {
		successVal = 0.0
	}
	s.langfuseCli.CreateScore(&langfuse.ScoreBody{
		Name:     "tool-success",
		TraceID:  traceID,
		Value:    successVal,
		DataType: "BOOLEAN",
		Comment:  result.ToolUsed,
	})

	// Score: latency in seconds
	s.langfuseCli.CreateScore(&langfuse.ScoreBody{
		Name:     "latency",
		TraceID:  traceID,
		Value:    float64(duration) / 1000.0,
		DataType: "NUMERIC",
		Comment:  fmt.Sprintf("%dms", duration),
	})
}

func (s *Service) createLLMTrace(traceID, startTime string, input agent.ChatInput, result *agent.ChatOutput, duration int64) {
	usage := &langfuse.Usage{}
	if result.InputTokens > 0 || result.OutputTokens > 0 {
		usage.Input = result.InputTokens
		usage.Output = result.OutputTokens
		usage.Total = result.InputTokens + result.OutputTokens
	}

	s.langfuseCli.CreateGeneration(&langfuse.GenerationBody{
		TraceID:   traceID,
		Name:      "llm-chat",
		StartTime: startTime,
		Model:     s.model,
		ModelParams: map[string]interface{}{
			"temperature": 0.3,
			"max_tokens":  2048,
		},
		Input: []interface{}{
			map[string]string{"role": "system", "content": "你是Kotoha零食导购助手"},
			map[string]string{"role": "user", "content": input.Message},
		},
		Output: result.Reply,
		Usage:  usage,
		Metadata: map[string]interface{}{
			"environment": s.env,
		},
	})

	// Score: latency
	s.langfuseCli.CreateScore(&langfuse.ScoreBody{
		Name:     "latency",
		TraceID:  traceID,
		Value:    float64(duration) / 1000.0,
		DataType: "NUMERIC",
		Comment:  fmt.Sprintf("%dms", duration),
	})
}

func classifyIntentTag(msg string) string {
	msg = strings.TrimSpace(msg)
	lower := strings.ToLower(msg)

	chatGreetings := []string{"你好", "嗨", "hi", "hello", "hey", "谢谢", "感谢", "再见", "bye", "价格", "帮帮我", "怎么", "什么"}
	for _, g := range chatGreetings {
		if strings.HasPrefix(lower, g) {
			return "chat"
		}
	}

	searchPrefixes := []string{"搜索", "找", "帮我找", "有没有", "有什么", "推荐", "search"}
	for _, p := range searchPrefixes {
		if strings.HasPrefix(msg, p) {
			return "search"
		}
	}

	cartPrefixes := []string{"购物车", "买", "加", "添加", "删除", "移除", "去掉", "add", "remove"}
	for _, p := range cartPrefixes {
		if strings.HasPrefix(msg, p) || strings.Contains(msg, p) {
			return "cart"
		}
	}

	if strings.Contains(msg, "商品") || strings.Contains(msg, "产品") || strings.Contains(msg, "详情") || strings.Contains(msg, "product") || strings.Contains(msg, "分类") {
		return "catalog"
	}

	return "general"
}

func inferFeature(msg string) string {
	msg = strings.TrimSpace(msg)
	if strings.Contains(msg, "购物车") || strings.HasPrefix(msg, "买") || strings.HasPrefix(msg, "加") {
		return "cart"
	}
	if strings.Contains(msg, "商品") || strings.Contains(msg, "分类") || strings.Contains(msg, "详情") {
		return "catalog"
	}
	if strings.HasPrefix(msg, "搜索") || strings.HasPrefix(msg, "找") || strings.HasPrefix(msg, "推荐") {
		return "search"
	}
	return "chat"
}

type EvalMetric struct {
	Category      string  `json:"category"`
	TotalSearches int     `json:"total_searches"`
	AvgRelevance  float64 `json:"avg_relevance"`
	CartAdds      int     `json:"cart_adds"`
	OrdersCreated int     `json:"orders_created"`
}

func (s *Service) SaveEvalRun(run *EvalRun) error {
	if s.evalRepo == nil {
		return nil
	}
	return s.evalRepo.SaveRun(run)
}

func (s *Service) ListEvalRuns(category string, limit int) ([]EvalRun, error) {
	if s.evalRepo == nil {
		return nil, nil
	}
	return s.evalRepo.ListRuns(category, limit)
}

func (s *Service) PushEvalScore(category, name string, value float64, comment string) {
	if s.langfuseCli == nil {
		return
	}
	traceID := s.langfuseCli.CreateTrace(&langfuse.TraceBody{
		Name:   "eval-" + category,
		Tags:   []string{"eval", "category:" + category},
		Input:  name,
		Output: fmt.Sprintf("%.4f", value),
	})
	s.langfuseCli.CreateScore(&langfuse.ScoreBody{
		Name:    name,
		TraceID: traceID,
		Value:   value,
		Comment: comment,
	})
	s.langfuseCli.FlushAsync()
}

const (
	redisKeySearchCount    = "metrics:search:count"
	redisKeySearchScoreSum = "metrics:search:score_sum"
	redisKeySearchScoreCnt = "metrics:search:score_count"
	redisKeyCartAdds       = "metrics:cart:adds"
	redisKeyOrdersCreated  = "metrics:orders:created"
)

func (s *Service) TrackSearch(relevanceScore float64) {
	ctx := context.Background()
	s.redisCli.RDB.Incr(ctx, redisKeySearchCount)
	if relevanceScore > 0 {
		s.redisCli.RDB.IncrByFloat(ctx, redisKeySearchScoreSum, relevanceScore)
		s.redisCli.RDB.Incr(ctx, redisKeySearchScoreCnt)
	}
}

func (s *Service) TrackCartAdd() {
	s.redisCli.RDB.Incr(context.Background(), redisKeyCartAdds)
}

func (s *Service) TrackOrderCreated() {
	s.redisCli.RDB.Incr(context.Background(), redisKeyOrdersCreated)
}

func (s *Service) GetMetrics() []EvalMetric {
	totalSearches := 0
	avgRelevance := 0.0
	cartAdds := 0
	ordersCreated := 0

	if s.redisCli != nil {
		ctx := context.Background()
		totalSearches = s.getInt(ctx, redisKeySearchCount)
		scoreSum := s.getFloat(ctx, redisKeySearchScoreSum)
		scoreCnt := s.getInt(ctx, redisKeySearchScoreCnt)
		if scoreCnt > 0 {
			avgRelevance = scoreSum / float64(scoreCnt)
		}
		cartAdds = s.getInt(ctx, redisKeyCartAdds)
		ordersCreated = s.getInt(ctx, redisKeyOrdersCreated)
	}

	return []EvalMetric{
		{Category: "search", TotalSearches: totalSearches, AvgRelevance: avgRelevance},
		{Category: "cart", CartAdds: cartAdds},
		{Category: "order", OrdersCreated: ordersCreated},
	}
}

func (s *Service) getInt(ctx context.Context, key string) int {
	val, err := s.redisCli.RDB.Get(ctx, key).Int()
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) getFloat(ctx context.Context, key string) float64 {
	val, err := s.redisCli.RDB.Get(ctx, key).Float64()
	if err != nil {
		return 0.0
	}
	return val
}
