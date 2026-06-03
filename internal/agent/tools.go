package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/search"
)

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"` // JSON schema for parameters
}

var toolDefs = []Tool{
	{
		Name:        "search_products",
		Description: "搜索零食产品，根据关键词(name/description/tags)搜索最匹配的商品",
		Parameters:  `{"type":"object","properties":{"query":{"type":"string","description":"搜索关键词"}},"required":["query"]}`,
	},
	{
		Name:        "get_product",
		Description: "获取单个产品的详细信息，包括所有SKU、价格和库存",
		Parameters:  `{"type":"object","properties":{"product_id":{"type":"integer","description":"产品ID"}},"required":["product_id"]}`,
	},
	{
		Name:        "list_categories",
		Description: "列出所有零食分类",
		Parameters:  `{"type":"object","properties":{}}`,
	},
	{
		Name:        "add_to_cart",
		Description: "将指定SKU和数量的商品添加到购物车",
		Parameters:  `{"type":"object","properties":{"sku_id":{"type":"integer","description":"SKU ID"},"quantity":{"type":"integer","description":"数量"}},"required":["sku_id","quantity"]}`,
	},
	{
		Name:        "get_cart",
		Description: "查看当前购物车中的所有商品",
		Parameters:  `{"type":"object","properties":{}}`,
	},
	{
		Name:        "remove_from_cart",
		Description: "从购物车中移除指定SKU的商品",
		Parameters:  `{"type":"object","properties":{"sku_id":{"type":"integer","description":"SKU ID"}},"required":["sku_id"]}`,
	},
}

type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Error   string `json:"error,omitempty"`
}

type ToolExecutor struct {
	catalogRepo *catalog.Repository
	searchSvc   *search.Service
	cartSvc     *cart.Service
}

func NewToolExecutor(catalogRepo *catalog.Repository, searchSvc *search.Service, cartSvc *cart.Service) *ToolExecutor {
	return &ToolExecutor{catalogRepo: catalogRepo, searchSvc: searchSvc, cartSvc: cartSvc}
}

func (e *ToolExecutor) Execute(ctx context.Context, userID uint, call ToolCall) ToolResult {
	switch call.Name {
	case "search_products":
		return e.searchProducts(call.Arguments)
	case "get_product":
		return e.getProduct(call.Arguments)
	case "list_categories":
		return e.listCategories()
	case "add_to_cart":
		return e.addToCart(ctx, userID, call.Arguments)
	case "get_cart":
		return e.getCart(ctx, userID)
	case "remove_from_cart":
		return e.removeFromCart(ctx, userID, call.Arguments)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("unknown tool: %s", call.Name)}
	}
}

func (e *ToolExecutor) searchProducts(args json.RawMessage) ToolResult {
	var p struct{ Query string }
	if err := json.Unmarshal(args, &p); err != nil {
		return ToolResult{Success: false, Error: "invalid arguments"}
	}
	products, err := e.searchSvc.Search(p.Query, 5)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	data, _ := json.Marshal(products)
	return ToolResult{Success: true, Data: string(data)}
}

func (e *ToolExecutor) getProduct(args json.RawMessage) ToolResult {
	var p struct{ ProductID uint }
	if err := json.Unmarshal(args, &p); err != nil {
		return ToolResult{Success: false, Error: "invalid arguments"}
	}
	product, err := e.catalogRepo.GetProduct(p.ProductID)
	if err != nil {
		return ToolResult{Success: false, Error: "product not found"}
	}
	data, _ := json.Marshal(product)
	return ToolResult{Success: true, Data: string(data)}
}

func (e *ToolExecutor) listCategories() ToolResult {
	cats, err := e.catalogRepo.ListCategories()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	data, _ := json.Marshal(cats)
	return ToolResult{Success: true, Data: string(data)}
}

func (e *ToolExecutor) addToCart(ctx context.Context, userID uint, args json.RawMessage) ToolResult {
	var p struct {
		SKUID    uint `json:"sku_id"`
		Quantity int  `json:"quantity"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return ToolResult{Success: false, Error: "invalid arguments"}
	}
	if err := e.cartSvc.AddItem(ctx, userID, p.SKUID, p.Quantity); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: "已添加到购物车"}
}

func (e *ToolExecutor) getCart(ctx context.Context, userID uint) ToolResult {
	items, err := e.cartSvc.GetCart(ctx, userID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if len(items) == 0 {
		return ToolResult{Success: true, Data: "购物车是空的"}
	}
	data, _ := json.Marshal(items)
	return ToolResult{Success: true, Data: string(data)}
}

func (e *ToolExecutor) removeFromCart(ctx context.Context, userID uint, args json.RawMessage) ToolResult {
	var p struct{ SKUID uint }
	if err := json.Unmarshal(args, &p); err != nil {
		return ToolResult{Success: false, Error: "invalid arguments"}
	}
	if err := e.cartSvc.RemoveItem(ctx, userID, p.SKUID); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: "已从购物车移除"}
}
