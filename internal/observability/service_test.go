package observability

import (
	"testing"
)

func TestClassifyIntentTag(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		// Chat greetings
		{"你好", "chat"},
		{"hi there", "chat"},
		{"hello!", "chat"},
		{"谢谢你的帮助", "chat"},
		{"价格多少", "chat"},
		{"帮帮我", "chat"},
		{"怎么用这个", "chat"},
		{"什么是零食", "chat"},

		// Search
		{"搜索坚果", "search"},
		{"找辣条", "search"},
		{"帮我找低卡的零食", "search"},
		{"有没有辣的", "search"},
		{"有什么好吃的", "search"},
		{"推荐零食", "search"},
		{"search snacks", "search"},

		// Cart
		{"购物车", "cart"},
		{"买一个试试", "cart"},
		{"加入购物车", "cart"},
		{"添加商品", "cart"},
		{"删除那个", "cart"},
		{"移除它", "cart"},
		{"add to cart", "cart"},
		{"remove item please", "cart"},

		// Catalog
		{"商品详情", "catalog"},
		{"查看产品", "catalog"},
		{"这个是什么product", "catalog"},
		{"有哪些分类", "catalog"},

		// General (fallthrough)
		{"random text here", "general"},
		{"????", "general"},
		{"", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := classifyIntentTag(tt.msg)
			if got != tt.want {
				t.Errorf("classifyIntentTag(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}

func TestInferFeature(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		// Cart
		{"购物车有什么", "cart"},
		{"买一个", "cart"},
		{"加入购物车", "cart"},

		// Catalog
		{"商品列表", "catalog"},
		{"分类有哪些", "catalog"},
		{"查看详情", "catalog"},

		// Search
		{"搜索零食", "search"},
		{"找辣的", "search"},
		{"推荐好吃的", "search"},

		// Chat (default)
		{"你好", "chat"},
		{"random", "chat"},
		{"", "chat"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := inferFeature(tt.msg)
			if got != tt.want {
				t.Errorf("inferFeature(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}

func TestEnv(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, "test-model", "development")
	if svc.env != "development" {
		t.Errorf("expected env to be 'development', got %q", svc.env)
	}
}

func TestGetMetrics(t *testing.T) {
	svc := &Service{}
	metrics := svc.GetMetrics()
	if len(metrics) == 0 {
		t.Fatal("expected non-empty metrics")
	}

	categories := map[string]bool{}
	for _, m := range metrics {
		categories[m.Category] = true
	}
	for _, want := range []string{"search", "cart", "order"} {
		if !categories[want] {
			t.Errorf("expected metric category %q", want)
		}
	}
}

func TestNewService(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, "test-model", "production")
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.model != "test-model" {
		t.Errorf("got model %q, want test-model", svc.model)
	}
	if svc.env != "production" {
		t.Errorf("got env %q, want production", svc.env)
	}
}
