package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hanasakis/kotoha/pkg/ollama"
)

const systemPrompt = `你是 Kotoha（果多哈）AI 零食导购助手。你帮助用户发现、了解和购买零食。

你的特点：
- 热情、专业、了解各种零食的口味和场景
- 优先用中文回复（除非用户用英文提问）
- 推荐产品时说明口味、适用场景和价格

可用工具：
- search_products: 搜索产品（关键词）
- get_product: 查看产品详情（product_id）
- list_categories: 查看所有分类
- add_to_cart: 加入购物车（sku_id, quantity）
- get_cart: 查看购物车
- remove_from_cart: 移除购物车商品（sku_id）

使用工具时，只输出以下JSON格式，不要输出其他内容：
{"tool":"工具名","args":{参数}}

用户想直接聊天时，正常回复中文即可。

价格单位是"分"（分人民币），显示时除以100转换为元。例如1990分 = ¥19.90。

保持回复简洁友好，像一位懂零食的朋友。`

type Service struct {
	ollamaCli *ollama.Client
	executor  *ToolExecutor
	model     string
	temperature float64
	maxTokens   int
}

func NewService(ollamaCli *ollama.Client, executor *ToolExecutor, model string, temperature float64, maxTokens int) *Service {
	return &Service{
		ollamaCli:   ollamaCli,
		executor:    executor,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
	}
}

type ChatInput struct {
	Message string `json:"message" binding:"required"`
}

type ChatOutput struct {
	Reply    string `json:"reply"`
	ToolUsed string `json:"tool_used,omitempty"`
}

func (s *Service) Chat(ctx context.Context, userID uint, history []ollama.Message, input ChatInput) (*ChatOutput, error) {
	messages := []ollama.Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, ollama.Message{Role: "user", Content: input.Message})

	for turn := 0; turn < 5; turn++ {
		resp, err := s.ollamaCli.Chat(s.model, messages, s.temperature, s.maxTokens)
		if err != nil {
			return nil, fmt.Errorf("agent.llm_error: %w", err)
		}

		content := strings.TrimSpace(resp.Message.Content)

		// Check for tool call
		if strings.HasPrefix(content, "{") && strings.Contains(content, "\"tool\"") {
			var toolCall struct {
				Tool string          `json:"tool"`
				Args json.RawMessage `json:"args"`
			}
			if err := json.Unmarshal([]byte(content), &toolCall); err == nil && toolCall.Tool != "" {
				result := s.executor.Execute(ctx, userID, ToolCall{
					Name:      toolCall.Tool,
					Arguments: toolCall.Args,
				})

				resultStr := result.Data
				if !result.Success {
					resultStr = result.Error
				}

				// Add assistant tool call + result to messages
				messages = append(messages, ollama.Message{Role: "assistant", Content: content})
				messages = append(messages, ollama.Message{
					Role:    "user",
					Content: fmt.Sprintf("工具 %s 返回结果: %s", toolCall.Tool, resultStr),
				})
				continue
			}
		}

		return &ChatOutput{Reply: content}, nil
	}

	return &ChatOutput{Reply: "抱歉，我暂时无法完成这个请求。请换种方式再试一次。"}, nil
}
