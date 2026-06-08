package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestClassifyIntent_SearchProducts(t *testing.T) {
	tests := []struct {
		msg      string
		wantTool string
		wantKey  string
		wantVal  string
	}{
		{"搜索坚果", "search_products", "query", "坚果"},
		{"找薯片", "search_products", "query", "薯片"},
		{"有没有辣的零食", "search_products", "query", "辣的零食"},
		{"推荐好吃的", "search_products", "query", "好吃的"},
		{"search chocolate", "search_products", "query", "chocolate"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			tool, args, ok := classifyIntent(tt.msg)
			if !ok {
				t.Fatalf("expected tool detection, got chat")
			}
			if tool != tt.wantTool {
				t.Errorf("got tool %q, want %q", tool, tt.wantTool)
			}
			if !strings.Contains(args, `"`+tt.wantKey+`":"`+tt.wantVal+`"`) {
				t.Errorf("args %q missing key=%q val=%q", args, tt.wantKey, tt.wantVal)
			}
		})
	}
}

func TestClassifyIntent_GetProduct(t *testing.T) {
	tests := []struct {
		msg       string
		productID int
	}{
		{"查看商品281", 281},
		{"商品详情42", 42},
		{"get product 999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			tool, args, ok := classifyIntent(tt.msg)
			if !ok {
				t.Fatalf("expected tool detection, got chat")
			}
			if tool != "get_product" {
				t.Errorf("got tool %q, want get_product", tool)
			}
			want := `"product_id":` + strings.TrimSpace(formatInt(tt.productID))
			if !strings.Contains(args, want) {
				t.Errorf("args %q missing %s", args, want)
			}
		})
	}
}

func TestClassifyIntent_ListCategories(t *testing.T) {
	for _, msg := range []string{"有哪些分类", "列出分类", "查看分类", "list categories"} {
		t.Run(msg, func(t *testing.T) {
			tool, args, ok := classifyIntent(msg)
			if !ok {
				t.Fatalf("expected tool detection, got chat")
			}
			if tool != "list_categories" {
				t.Errorf("got tool %q, want list_categories", tool)
			}
			if args != "{}" {
				t.Errorf("got args %q, want {}", args)
			}
		})
	}
}

func TestClassifyIntent_GetCart(t *testing.T) {
	for _, msg := range []string{"购物车", "我的购物车", "查看购物车", "购物车里有什么", "get cart"} {
		t.Run(msg, func(t *testing.T) {
			tool, args, ok := classifyIntent(msg)
			if !ok {
				t.Fatalf("expected tool detection, got chat")
			}
			if tool != "get_cart" {
				t.Errorf("got tool %q, want get_cart", tool)
			}
			if args != "{}" {
				t.Errorf("got args %q, want {}", args)
			}
		})
	}
}

func TestClassifyIntent_AddToCart(t *testing.T) {
	t.Run("buy with quantity", func(t *testing.T) {
		tool, args, ok := classifyIntent("买351加2个")
		if !ok {
			t.Fatal("expected tool detection, got chat")
		}
		if tool != "add_to_cart" {
			t.Errorf("got tool %q, want add_to_cart", tool)
		}
		if !strings.Contains(args, `"sku_id":351`) {
			t.Errorf("args %q missing sku_id", args)
		}
		if !strings.Contains(args, `"quantity":2`) {
			t.Errorf("args %q missing quantity", args)
		}
	})

	t.Run("add without quantity defaults to 1", func(t *testing.T) {
		tool, args, ok := classifyIntent("加入351到购物车")
		if !ok {
			t.Fatal("expected tool detection, got chat")
		}
		if tool != "add_to_cart" {
			t.Errorf("got tool %q, want add_to_cart", tool)
		}
		if !strings.Contains(args, `"sku_id":351`) {
			t.Errorf("args %q missing sku_id", args)
		}
		if !strings.Contains(args, `"quantity":1`) {
			t.Errorf("args %q missing quantity default 1", args)
		}
	})
}

func TestClassifyIntent_RemoveFromCart(t *testing.T) {
	t.Run("remove with sku_id", func(t *testing.T) {
		tool, args, ok := classifyIntent("删除351")
		if !ok {
			t.Fatal("expected tool detection, got chat")
		}
		if tool != "remove_from_cart" {
			t.Errorf("got tool %q, want remove_from_cart", tool)
		}
		if !strings.Contains(args, `"sku_id":351`) {
			t.Errorf("args %q missing sku_id", args)
		}
	})

	t.Run("remove from cart with context", func(t *testing.T) {
		tool, args, ok := classifyIntent("从购物车删除商品351")
		if !ok {
			t.Fatal("expected tool detection, got chat")
		}
		if tool != "remove_from_cart" {
			t.Errorf("got tool %q, want remove_from_cart", tool)
		}
		if !strings.Contains(args, `"sku_id":351`) {
			t.Errorf("args %q missing sku_id", args)
		}
	})

	t.Run("remove without number", func(t *testing.T) {
		tool, args, ok := classifyIntent("删除")
		if !ok {
			t.Fatal("expected tool detection, got chat")
		}
		if tool != "remove_from_cart" {
			t.Errorf("got tool %q, want remove_from_cart", tool)
		}
		if args != "{}" {
			t.Errorf("got args %q, want {}", args)
		}
	})
}

func TestClassifyIntent_ChatGreetings(t *testing.T) {
	greetings := []string{"你好", "嗨", "hi", "hello", "hey", "价格怎么算", "帮帮我", "怎么用", "什么是零食", "谢谢", "感谢", "再见", "bye"}
	for _, msg := range greetings {
		t.Run(msg, func(t *testing.T) {
			_, _, ok := classifyIntent(msg)
			if ok {
				t.Errorf("expected chat (no tool) for %q, got tool detected", msg)
			}
		})
	}
}

func TestClassifyIntent_UnknownGoesToLLM(t *testing.T) {
	for _, msg := range []string{"random text", "不知道说什么", "???"} {
		t.Run(msg, func(t *testing.T) {
			_, _, ok := classifyIntent(msg)
			if ok {
				t.Errorf("expected chat for %q, got tool detected", msg)
			}
		})
	}
}

func formatInt(n int) string {
	s, _ := json.Marshal(n)
	return string(s)
}

func TestChat_ClassifierPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// classifierPath tests require PostgreSQL + Redis (for ToolExecutor dependencies).
	// Pure classifyIntent unit tests above cover the classifier logic without infrastructure.
	// This test is a placeholder for full integration when all dependencies are wired.
}

func TestChat_GreetingPath(t *testing.T) {
	t.Skip("requires Ollama LLM running locally")
}

// Verify ToolCall JSON structure used by executor
func TestToolCall_JSON(t *testing.T) {
	call := ToolCall{Name: "search_products", Arguments: json.RawMessage(`{"query":"坚果"}`)}
	data, err := json.Marshal(call)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded ToolCall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Name != "search_products" {
		t.Errorf("got name %q, want search_products", decoded.Name)
	}
}

// Verify ToolResult JSON structure
func TestToolResult_JSON(t *testing.T) {
	success := ToolResult{Success: true, Data: "test data"}
	if !success.Success {
		t.Error("expected success")
	}
	if success.Data != "test data" {
		t.Errorf("got data %q, want 'test data'", success.Data)
	}

	failure := ToolResult{Success: false, Error: "something went wrong"}
	if failure.Success {
		t.Error("expected failure")
	}
	if failure.Error != "something went wrong" {
		t.Errorf("got error %q", failure.Error)
	}
}

// Test systemPrompt contains all tool names
func TestSystemPrompt_ToolCoverage(t *testing.T) {
	tools := []string{"search_products", "get_product", "list_categories", "add_to_cart", "get_cart", "remove_from_cart"}
	for _, tool := range tools {
		if !strings.Contains(systemPrompt, tool) {
			t.Errorf("systemPrompt missing tool %q", tool)
		}
	}
}

// Verify classifyIntent handles empty/whitespace input
func TestClassifyIntent_Empty(t *testing.T) {
	_, _, ok := classifyIntent("")
	if ok {
		t.Error("expected no tool for empty message")
	}
	_, _, ok = classifyIntent("   ")
	if ok {
		t.Error("expected no tool for whitespace-only message")
	}
}

// Integration: test Chat() with classifier path (requires DB + Redis, skip if short)
func TestChat_Integration(t *testing.T) {
	t.Skip("requires full stack: PostgreSQL, Redis, Ollama")
	_ = context.Background()
}
