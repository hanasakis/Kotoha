package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	klog "github.com/hanasakis/kotoha/pkg/log"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/cart"
	usersvc "github.com/hanasakis/kotoha/internal/user"
	"github.com/hanasakis/kotoha/pkg/ollama"
)

const systemPrompt = `你是Kotoha(果多哈)零食导购助手。你必须只输出JSON工具调用,不要输出任何其他内容。

工具列表:
- search_products: query(搜索词)
- get_product: product_id(数字)
- list_categories: 无参数
- add_to_cart: sku_id(数字) quantity(数字)
- get_cart: 无参数
- remove_from_cart: sku_id(数字)

示例:
用户: 搜索坚果 → 你输出: {"tool":"search_products","args":{"query":"坚果"}}
用户: 查看商品281 → 你输出: {"tool":"get_product","args":{"product_id":281}}
用户: 有哪些分类 → 你输出: {"tool":"list_categories","args":{}}
用户: 把商品351加2个到购物车 → 你输出: {"tool":"add_to_cart","args":{"sku_id":351,"quantity":2}}
用户: 我的购物车有什么 → 你输出: {"tool":"get_cart","args":{}}
用户: 从购物车删除商品351 → 你输出: {"tool":"remove_from_cart","args":{"sku_id":351}}

对话示例:
用户: 你好 → 你输出: 你好！我是Kotoha，可以帮你搜索和购买零食~
用户: 价格怎么算 → 你输出: 1990分=¥19.90，价格以分为单位显示

记住: 搜索/商品/分类/购物车相关请求必须只输出JSON，纯聊天可以简短回复中文。`

const prefExtractionPrompt = `你是Kotoha偏好分析师。从用户的自然语言描述中提取饮食偏好，只输出JSON，不要任何额外内容。

输出格式:
{"dietary_limits":"饮食限制(素食/清真/低碳/无糖等)","taste_prefs":"口味偏好(辣/甜/咸/酸等)","scene_prefs":"场景偏好(追剧/办公/健身/送礼等)","allergens":"过敏原(花生/牛奶/海鲜/坚果等)"}

没有的字段填空字符串""。只输出JSON。`

type Service struct {
	ollamaCli   *ollama.Client
	executor    *ToolExecutor
	userSvc     *usersvc.Service
	model       string
	temperature float64
	maxTokens   int
}

func NewService(ollamaCli *ollama.Client, executor *ToolExecutor, userSvc *usersvc.Service, model string, temperature float64, maxTokens int) *Service {
	return &Service{
		ollamaCli:   ollamaCli,
		executor:    executor,
		userSvc:     userSvc,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
	}
}

type ChatInput struct {
	Message string `json:"message" binding:"required"`
}

type ChatOutput struct {
	Reply        string `json:"reply"`
	ToolUsed     string `json:"tool_used,omitempty"`
	Path         string `json:"path,omitempty"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
}

// ----- Preference detection -----

// isRecommendationIntent checks if this is a "recommend snacks" request.
func isRecommendationIntent(msg string) bool {
	return regexp.MustCompile(`推荐.*零食|推荐.*吃的|给我推荐|帮我推荐|推荐一下|有什么.*推荐|推荐.*好.*的`).MatchString(msg) ||
		regexp.MustCompile(`^(?:推荐|帮我推荐|给我推荐)\s*$`).MatchString(msg)
}

// isPreferenceDescription detects if user is describing their dietary preferences.
func isPreferenceDescription(msg string) bool {
	patterns := []string{
		`(?:我|本人).*?(?:喜欢|爱吃|喜欢吃|爱吃|偏好|钟爱|口味).*?(?:辣|甜|咸|酸|苦|麻|鲜|香)`,
		`(?:不吃|忌口|忌|不能吃|不敢吃|拒绝|讨厌|不喜欢)`,
		`(?:过敏|不能碰)`,
		`(?:素食|清真|低碳|生酮|无糖|低卡|低脂|高蛋白)`,
		`(?:追剧|办公|健身|送礼|聚会|出游|夜宵|下午茶).*?(?:吃|零食)`,
		`我的.*?(?:口味|偏好|饮食|忌口|过敏)`,
		`(?:饮食|口味|偏好|场景|过敏).*?(?:是|为|：|:)`,
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, msg); matched {
			return true
		}
	}
	return false
}

// hasPreferences checks if user has any meaningful preference data.
func hasPreferences(pref *usersvc.Preference) bool {
	if pref == nil {
		return false
	}
	return pref.DietaryLimits != "" || pref.TastePrefs != "" || pref.ScenePrefs != "" || pref.Allergens != ""
}

// preferenceContext builds a search-augmenting string from user preferences.
func preferenceContext(pref *usersvc.Preference) string {
	var parts []string
	if pref.TastePrefs != "" {
		parts = append(parts, pref.TastePrefs)
	}
	if pref.DietaryLimits != "" {
		parts = append(parts, pref.DietaryLimits)
	}
	if pref.ScenePrefs != "" {
		parts = append(parts, pref.ScenePrefs)
	}
	if pref.Allergens != "" {
		parts = append(parts, "不含"+pref.Allergens)
	}
	return strings.Join(parts, " ")
}

// formatPrefSummary returns a human-readable summary of user preferences.
func formatPrefSummary(pref *usersvc.Preference) string {
	var parts []string
	if pref.TastePrefs != "" {
		parts = append(parts, "口味："+pref.TastePrefs)
	}
	if pref.DietaryLimits != "" {
		parts = append(parts, "饮食："+pref.DietaryLimits)
	}
	if pref.Allergens != "" {
		parts = append(parts, "忌："+pref.Allergens)
	}
	if pref.ScenePrefs != "" {
		parts = append(parts, "场景："+pref.ScenePrefs)
	}
	if len(parts) == 0 {
		return ""
	}
	return "你的偏好 — " + strings.Join(parts, " | ")
}

// extractPreferencesFromNL uses LLM to parse natural language into structured preferences.
func (s *Service) extractPreferencesFromNL(userMsg string) (*usersvc.Preference, error) {
	messages := []ollama.Message{
		{Role: "system", Content: prefExtractionPrompt},
		{Role: "user", Content: userMsg},
	}

	resp, err := s.ollamaCli.Chat(s.model, messages, 0.1, 512, nil)
	if err != nil {
		return nil, fmt.Errorf("pref.extraction_error: %w", err)
	}

	content := strings.TrimSpace(resp.Message.Content)
	// Strip markdown code fences if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var pref usersvc.Preference
	if err := json.Unmarshal([]byte(content), &pref); err != nil {
		klog.Warnf("[AGENT] failed to parse preference JSON: %v raw=%q", err, content)
		return nil, fmt.Errorf("pref.parse_error: %w", err)
	}
	return &pref, nil
}

// enhanceSearchQuery appends preference context to the search keyword.
func enhanceSearchQuery(keyword string, pref *usersvc.Preference) string {
	ctx := preferenceContext(pref)
	if ctx == "" {
		return keyword
	}
	// Append preference context as additional keywords
	return keyword + " " + ctx
}

// ----- classifyIntent -----

// classifyIntent pre-processes the user message to detect clear tool intents.
func classifyIntent(msg string) (string, string, bool) {
	msg = strings.TrimSpace(msg)

	// Add to cart variant
	if re := regexp.MustCompile(`(?:加|添加|放入|放)(?:到|入|进).*?(\d+)`); re.MatchString(msg) {
		m := re.FindStringSubmatch(msg)
		skuID, _ := strconv.Atoi(m[1])
		return "add_to_cart", fmt.Sprintf(`{"sku_id":%d,"quantity":1}`, skuID), true
	}

	// Remove from cart
	if re := regexp.MustCompile(`(?:删除|移除|去掉|remove).*?(\d+)`); re.MatchString(msg) {
		m := re.FindStringSubmatch(msg)
		id, _ := strconv.Atoi(m[1])
		return "remove_from_cart", fmt.Sprintf(`{"sku_id":%d}`, id), true
	}
	if matched, _ := regexp.MatchString(`^(删除|移除|去掉|remove)`, msg); matched {
		return "remove_from_cart", "{}", true
	}

	// Add to cart
	if re := regexp.MustCompile(`(?:买|加|添加|add|放入).*?(\d+)\D*(\d+)?\s*(?:个|件|份|quantity)?`); re.MatchString(msg) {
		m := re.FindStringSubmatch(msg)
		skuID, _ := strconv.Atoi(m[1])
		qty := 1
		if len(m) > 2 && m[2] != "" {
			qty, _ = strconv.Atoi(m[2])
		}
		return "add_to_cart", fmt.Sprintf(`{"sku_id":%d,"quantity":%d}`, skuID, qty), true
	}

	// Get cart
	if matched, _ := regexp.MatchString(`购物车|cart|show.*(my )?(cart|shopping)`, msg); matched {
		return "get_cart", "{}", true
	}

	// List categories
	if matched, _ := regexp.MatchString(`分类|categor`, msg); matched {
		return "list_categories", "{}", true
	}

	// Get product by ID
	if re := regexp.MustCompile(`(?:商品|产品|详情|get\s*product)\s*(\d+)`); re.MatchString(msg) {
		m := re.FindStringSubmatch(msg)
		id, _ := strconv.Atoi(m[1])
		return "get_product", fmt.Sprintf(`{"product_id":%d}`, id), true
	}

	// Greeting/chat patterns
	for _, g := range []string{"你好", "嗨", "hi", "hello", "hey", "帮帮我", "谢谢", "感谢", "再见", "bye", "价格怎么算", "怎么用", "什么是"} {
		if strings.HasPrefix(strings.ToLower(msg), g) {
			klog.Infof("[AGENT] classifyIntent: matched greeting %q for msg=%q", g, msg)
			return "", "", false
		}
	}

	// Search: precise intent prefixes
	searchRe := regexp.MustCompile(`^(?:搜索|帮我找|search|find|looking\s*for)\s*(.+)`)
	if m := searchRe.FindStringSubmatch(msg); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
		keyword := cleanKeyword(m[1])
		return "search_products", fmt.Sprintf(`{"query":"%s"}`, keyword), true
	}

	// Recommendation: "帮我推荐/给我推荐/推荐一些 + keyword"
	if re := regexp.MustCompile(`(?:帮我推荐|给我推荐|推荐一下|推荐一些|推荐几个|推荐)\s*(.+)`); re.MatchString(msg) {
		m := re.FindStringSubmatch(msg)
		keyword := cleanKeyword(m[1])
		for _, prefix := range []string{"一些", "几个", "一下", "好的", "好吃的", "好玩的"} {
			keyword = strings.TrimPrefix(keyword, prefix)
			keyword = strings.TrimSpace(keyword)
		}
		if len([]rune(keyword)) > 1 {
			return "search_products", fmt.Sprintf(`{"query":"%s"}`, keyword), true
		}
	}

	// Broader search
	broadRe := regexp.MustCompile(`(?:有没有|有哪些?|有什么|帮我看看|找找|找)\s*(.+)`)
	if m := broadRe.FindStringSubmatch(msg); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
		keyword := cleanKeyword(m[1])
		keyword = strings.ReplaceAll(keyword, "可以推荐", "")
		keyword = strings.ReplaceAll(keyword, "推荐", "")
		keyword = strings.TrimSpace(keyword)
		if len([]rune(keyword)) > 1 {
			return "search_products", fmt.Sprintf(`{"query":"%s"}`, keyword), true
		}
	}

	klog.Infof("[AGENT] classifyIntent: no match for msg=%q (len=%d runes)", msg, len([]rune(msg)))
	return "", "", false
}

func cleanKeyword(k string) string {
	k = strings.TrimSpace(k)
	k = strings.TrimRight(k, "。！？,.!?，、：:；;…~～")
	for _, p := range []string{"吗", "呢", "吧", "啊", "呀", "哦", "哟", "嘛"} {
		k = strings.TrimSuffix(k, p)
	}
	return strings.TrimSpace(k)
}

// ----- Multi-step -----

func isMultiStep(msg string) bool {
	if regexp.MustCompile(`(并|然后|接着|之后|再|也|同时|并且|and\s+then|also).*(加|买|添加|搜索|找|推荐|查看|删除|放入|购物车)`).MatchString(msg) {
		return true
	}
	hasSearch := regexp.MustCompile(`搜索|帮我找|推荐|有没有|有什么|找找|帮我推荐`).MatchString(msg)
	hasCartAction := regexp.MustCompile(`加入购物车|加到购物车|放入购物车|添加.*购物车|买.*个.*购物车`).MatchString(msg)
	return hasSearch && hasCartAction
}

func (s *Service) runMultiStep(ctx context.Context, userID uint, msg string) (*ChatOutput, error) {
	re := regexp.MustCompile(`(?:推荐|搜索|帮我找|找|有没有|有什么)\s*(.+?)\s*(?:并|然后|接着|之后|再|也|and|,|，)\s*(?:把.*)?(?:加入|加到|添加|放入|买)`)
	m := re.FindStringSubmatch(msg)
	if m == nil {
		re2 := regexp.MustCompile(`(.+?)(?:并|然后|接着|之后|再)\s*.+?(?:购物车|加购|买)`)
		m = re2.FindStringSubmatch(msg)
	}
	if m == nil {
		if tool, args, ok := classifyIntent(msg); ok {
			result := s.executor.Execute(ctx, userID, ToolCall{Name: tool, Arguments: json.RawMessage(args)})
			if result.Success {
				return &ChatOutput{Reply: formatToolReply(tool, result.Data, nil), ToolUsed: tool, Path: "classifier"}, nil
			}
			return &ChatOutput{Reply: "抱歉，" + result.Error, ToolUsed: tool, Path: "classifier"}, nil
		}
		return nil, nil
	}

	keyword := cleanKeyword(m[1])
	if len([]rune(keyword)) == 0 {
		return nil, nil
	}

	klog.Infof("[AGENT] runMultiStep: keyword=%q", keyword)

	// Augment with preferences
	query := keyword
	if s.userSvc != nil {
		if pref, _ := s.userSvc.GetPreference(userID); pref != nil && hasPreferences(pref) {
			query = enhanceSearchQuery(keyword, pref)
			klog.Infof("[AGENT] runMultiStep: augmented query=%q", query)
		}
	}

	searchResult := s.executor.Execute(ctx, userID, ToolCall{
		Name:      "search_products",
		Arguments: json.RawMessage(fmt.Sprintf(`{"query":"%s"}`, query)),
	})

	if !searchResult.Success || searchResult.Data == "" || searchResult.Data == "null" {
		return &ChatOutput{
			Reply:    fmt.Sprintf("没有找到与「%s」相关的商品，换个关键词试试吧~", keyword),
			ToolUsed: "search_products",
			Path:     "multi-step",
		}, nil
	}

	var products []catalog.Product
	if err := json.Unmarshal([]byte(searchResult.Data), &products); err != nil || len(products) == 0 {
		return &ChatOutput{
			Reply:    fmt.Sprintf("没有找到与「%s」相关的商品~", keyword),
			ToolUsed: "search_products",
			Path:     "multi-step",
		}, nil
	}

	var addedSKUs []string
	for _, p := range products {
		if len(p.SKUs) == 0 {
			continue
		}
		sku := p.SKUs[0]
		addResult := s.executor.Execute(ctx, userID, ToolCall{
			Name:      "add_to_cart",
			Arguments: json.RawMessage(fmt.Sprintf(`{"sku_id":%d,"quantity":1}`, sku.ID)),
		})
		if addResult.Success {
			addedSKUs = append(addedSKUs, fmt.Sprintf("%s（%s，¥%.2f）", p.Name, sku.Name, float64(sku.Price)/100))
		}
		if len(addedSKUs) >= 3 {
			break
		}
	}

	var reply string
	if len(addedSKUs) == 0 {
		reply = fmt.Sprintf("为你找到「%s」相关商品，但添加购物车失败，请重试~", keyword)
	} else {
		reply = fmt.Sprintf("已为你找到并加入购物车：\n")
		for i, s := range addedSKUs {
			reply += fmt.Sprintf("  %d. %s\n", i+1, s)
		}
		reply += fmt.Sprintf("\n共 %d 件商品，快去结算吧~", len(addedSKUs))
	}

	return &ChatOutput{
		Reply:    reply,
		ToolUsed: "search_products,add_to_cart",
		Path:     "multi-step",
	}, nil
}

// ----- formatToolReply -----

func formatToolReply(tool string, data string, pref *usersvc.Preference) string {
	switch tool {
	case "search_products":
		var products []catalog.Product
		if err := json.Unmarshal([]byte(data), &products); err != nil || len(products) == 0 {
			return "没有找到相关商品，换个关键词试试吧~"
		}
		var b strings.Builder
		if pref != nil && hasPreferences(pref) {
			b.WriteString(formatPrefSummary(pref) + "\n\n")
		}
		b.WriteString(fmt.Sprintf("为你找到 %d 个相关商品：\n\n", len(products)))
		for i, p := range products {
			if i >= 5 {
				b.WriteString("...还有更多商品，试试缩小搜索范围~\n")
				break
			}
			lowestPrice := 0
			if len(p.SKUs) > 0 {
				lowestPrice = p.SKUs[0].Price
				for _, sku := range p.SKUs {
					if sku.Price < lowestPrice {
						lowestPrice = sku.Price
					}
				}
			}
			b.WriteString(fmt.Sprintf("  %d. %s — ¥%.2f (%s)\n", i+1, p.Name, float64(lowestPrice)/100, p.Tags))
		}
		return strings.TrimSpace(b.String())

	case "get_product":
		var p catalog.Product
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return "未找到该商品"
		}
		var b strings.Builder
		b.WriteString(fmt.Sprintf("【%s】\n%s\n\n", p.Name, p.Description))
		for _, sku := range p.SKUs {
			b.WriteString(fmt.Sprintf("  %s — ¥%.2f (库存: %d)\n", sku.Name, float64(sku.Price)/100, sku.Stock))
		}
		return strings.TrimSpace(b.String())

	case "list_categories":
		var cats []catalog.Category
		if err := json.Unmarshal([]byte(data), &cats); err != nil || len(cats) == 0 {
			return "暂无分类"
		}
		var b strings.Builder
		b.WriteString("零食分类如下：\n")
		for i, cat := range cats {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, cat.Name))
		}
		return strings.TrimSpace(b.String())

	case "get_cart":
		var items []cart.CartItem
		if err := json.Unmarshal([]byte(data), &items); err != nil || len(items) == 0 {
			return "购物车是空的，去逛逛吧~"
		}
		var b strings.Builder
		b.WriteString("你的购物车：\n")
		total := 0
		for i, item := range items {
			subtotal := item.Price * item.Quantity
			total += subtotal
			b.WriteString(fmt.Sprintf("  %d. %s x%d — ¥%.2f\n", i+1, item.ProductName, item.Quantity, float64(subtotal)/100))
		}
		b.WriteString(fmt.Sprintf("\n合计: ¥%.2f", float64(total)/100))
		return strings.TrimSpace(b.String())

	default:
		return data
	}
}

// ----- Chat entry point -----

func (s *Service) Chat(ctx context.Context, userID uint, history []ollama.Message, input ChatInput) (*ChatOutput, error) {
	msg := input.Message

	// --- Phase 1: Multi-step detection ---
	if isMultiStep(msg) {
		klog.Infof("[AGENT] multi-step detected: msg=%q", msg)
		if output, err := s.runMultiStep(ctx, userID, msg); output != nil || err != nil {
			return output, err
		}
	}

	// --- Phase 2: Preference description detection ---
	// User is describing their dietary preferences in natural language
	if isPreferenceDescription(msg) && s.userSvc != nil {
		klog.Infof("[AGENT] preference description detected: msg=%q", msg)
		extracted, err := s.extractPreferencesFromNL(msg)
		if err == nil && extracted != nil {
			extracted.UserID = userID
			if saveErr := s.userSvc.UpsertPreference(extracted); saveErr != nil {
				klog.Warnf("[AGENT] failed to save preferences: %v", saveErr)
			} else {
				summary := formatPrefSummary(extracted)
				return &ChatOutput{
					Reply: fmt.Sprintf("已了解你的偏好！%s\n\n现在可以为你精准推荐零食了，试试对我说「推荐零食」吧~", summary),
					Path:  "preference-save",
				}, nil
			}
		}
		// If extraction failed, fall through — don't block the user
		klog.Warnf("[AGENT] preference extraction failed, falling through: %v", err)
	}

	// --- Phase 3: Recommendation intent with empty preferences ---
	if isRecommendationIntent(msg) && s.userSvc != nil {
		pref, _ := s.userSvc.GetPreference(userID)
		if !hasPreferences(pref) {
			klog.Infof("[AGENT] recommendation intent with no preferences, eliciting")
			return &ChatOutput{
				Reply: "在为你推荐之前，我想先了解你的口味偏好~\n\n请随便说说，比如：\n• 喜欢什么口味？（辣、甜、咸、酸…）\n• 有什么忌口？（不吃猪肉、清真、低碳…）\n• 对什么过敏？（花生、牛奶、海鲜…）\n• 通常在什么场景吃？（追剧、办公、健身…）\n\n直接告诉我就好，我能听懂！",
				Path:  "preference-elicit",
			}, nil
		}
	}

	// --- Phase 4: Classifier ---
	if tool, args, ok := classifyIntent(msg); ok {
		klog.Infof("[AGENT] classifyIntent: msg=%q tool=%s args=%s", msg, tool, args)

		// For search_products, augment with user preferences
		if tool == "search_products" && s.userSvc != nil {
			pref, _ := s.userSvc.GetPreference(userID)
			if hasPreferences(pref) {
				// Parse original args to extract query
				var searchArgs struct{ Query string }
				json.Unmarshal([]byte(args), &searchArgs)
				enhanced := enhanceSearchQuery(searchArgs.Query, pref)
				args = fmt.Sprintf(`{"query":"%s"}`, enhanced)
				klog.Infof("[AGENT] enhanced search query: %q -> %q", searchArgs.Query, enhanced)
			}
		}

		result := s.executor.Execute(ctx, userID, ToolCall{Name: tool, Arguments: json.RawMessage(args)})
		reply := result.Data
		if !result.Success {
			reply = "抱歉，" + result.Error
		} else {
			// Pass preferences for formatted output
			var pref *usersvc.Preference
			if s.userSvc != nil {
				pref, _ = s.userSvc.GetPreference(userID)
			}
			reply = formatToolReply(tool, result.Data, pref)
		}
		return &ChatOutput{Reply: reply, ToolUsed: tool, Path: "classifier"}, nil
	}

	// --- Phase 5: Fall back to LLM for chat ---
	messages := []ollama.Message{
		{Role: "system", Content: "你是Kotoha(果多哈)零食导购助手。用中文简短友好地回复用户，不超过两句话。你可以帮用户搜索零食、查看商品详情、管理购物车。"},
	}
	messages = append(messages, history...)
	messages = append(messages, ollama.Message{Role: "user", Content: msg})

	resp, err := s.ollamaCli.Chat(s.model, messages, s.temperature, s.maxTokens, nil)
	if err != nil {
		return nil, fmt.Errorf("agent.llm_error: %w", err)
	}

	content := strings.TrimSpace(resp.Message.Content)

	if content == "" {
		content = "你好！我是Kotoha，可以帮你搜索零食、查看商品和管理购物车~"
	}

	return &ChatOutput{
		Reply:        content,
		Path:         "llm",
		InputTokens:  resp.InputTokens(),
		OutputTokens: resp.OutputTokens(),
	}, nil
}
