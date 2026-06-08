# Kotoha 项目演示指南

## 10 分钟演示脚本

### 0:00-1:00 — 开场：问题与愿景

> "Kotoha 是一个 AI 驱动的零食电商平台。它解决了一个真实问题——用户在零食选购时面临选择困难、偏好匹配不准、购物流程割裂三大痛点。我们通过 LLM Agent + 混合搜索 + 偏好学习，让 AI 成为你的私人零食导购。"

**Slide**: 问题 → 方案架构图

### 1:00-5:00 — Live Demo（核心演示）

**Demo 1: 新用户体验流** (2min)
1. 打开 localhost:3000 → 首页（分类卡片 + 推荐商品）
2. 注册新用户 → 浏览商品 → 点击分类筛选（6 个分类不同商品）
3. 点击 AI 助手 → 发送"推荐一些零食"
4. **AI 引导填写偏好**："在为你推荐之前，我想先了解你的口味偏好~"
5. 用户自然语言描述："我喜欢吃辣的，不吃花生，对海鲜过敏"
6. **AI 自动提取并保存偏好** → 即时返回精准推荐

**Demo 2: 偏好精准推荐** (1.5min)
7. 发送"推荐零食" → AI 根据已保存偏好精准搜索
8. 展示搜索结果避开过敏原、匹配口味
9. 发送"推荐巧克力并加入购物车" → 多步骤编排：搜索 + 加购

**Demo 3: 完整购物流程** (1.5min)
10. 查看购物车 → 结算 → 创建订单
11. Stripe 测试支付 → 支付成功 → 自动同步订单状态为 paid
12. 查看订单详情 → 显示已支付

### 5:00-7:00 — 管理后台

13. 切换 admin 账号 → Header 出现"管理"入口
14. 仪表盘：统计卡片 + 评测指标概览
15. 商品管理：CRUD + 搜索索引重建
16. 可观测性：Langfuse 实时追踪（切换浏览器 Tab 到 cloud.langfuse.com）
17. 评测分析：评测运行记录 + 准确率趋势

### 7:00-9:00 — 设计决策深度解析

**为什么用正则分类器而非 LLM Function Calling？**
> "qwen3.5:9b 不支持原生 tool calling。我们尝试过文本 ReAct 循环，但模型会幻觉工具结果。最终方案：正则意图识别直接执行 → 6 工具覆盖 95% 场景 → LLM 只做兜底闲聊。这保证了延迟（<50ms 分类）和准确性（零幻觉）。"

**为什么选 Milvus 混合搜索？**
> "纯向量搜索对中文零食名（如'灯影牛肉丝'）召回不佳。我们用了 Dense (bge-large-zh-v1.5) + BM25 关键词混合搜索，权重 0.65/0.35，保证语义理解 + 关键词精确匹配。"

**为什么用规则多步骤而非 LLM 编排？**
> "LangGraph 风格的编排对于 6 个工具的场景太重。'搜索+加购'是最常见的组合模式——正则提取关键词 → 搜索 → 取 Top3 加购，整个流程在 200ms 内完成，不需要 LLM 参与。"

**为什么选 TanStack Query + Zustand？**
> "服务端状态（API 数据）用 TanStack Query 管理缓存和去重；客户端状态（认证、购物车、语言）用 Zustand。职责分离，各有优化。"

**为什么 Langfuse 而非 MLflow？**
> "MLflow 偏向实验跟踪和模型注册。Langfuse 专为 LLM 应用设计：trace/span/generation 三级结构、token 用量追踪、score 评分体系，开箱即用。"

### 9:00-10:00 — 评测指标 & 未来规划

**评测指标体系**:

| 指标 | 定义 | 数据来源 | 为什么选它 |
|------|------|----------|-----------|
| Intent Accuracy | 意图分类正确率 | Eval Runs (agent_accuracy) | 核心路径——分类错则全部错 |
| Search Relevance | 搜索 NDCG@10, MRR | Eval Runs (search_relevance) | 衡量混合搜索排序质量 |
| Recommendation Hit Rate | Hit@5 推荐命中率 | Eval Runs (recommendation) | 偏好匹配有效性 |
| Tool Success Rate | 工具执行成功率 | Langfuse Scores (tool-success) | 每个 trace 自动评分 |
| P50/P95 Latency | 端到端延迟 | Langfuse Scores (latency) | 用户体验关键指标 |
| Conversion Rate | 搜索→加购→下单转化 | Redis Metrics | 业务导向指标 |

**当前系统数据**:
- 6 分类 / 30 商品 / 30 SKU
- Agent 分类器: 6 工具, 12 种正则模式
- 混合搜索: Dense 65% + BM25 35%, Recall Top40
- Langfuse: trace-create + span-create + generation-create + score-create

**With more time, I would**:
1. **Ragas 评测集成** — 自动评估 Agent 回复的忠实度、相关性、有害性
2. **A/B 实验框架** — 多模型/多 prompt 的在线对比评估
3. **用户反馈闭环** — 支付成功页加入评分 → Langfuse Scores → 反馈到模型调优
4. **Flutter 移动端** — 跨平台客户端，复用同一 API
5. **Kubernetes 部署** — Helm chart + 水平扩展 + 灰度发布
6. **实时推荐** — 用户浏览行为 → Kafka → 实时特征 → 动态推荐
7. **多语言扩展到日本市场** — 复用 i18n 架构，添加日语 locale

---

## Live Demo Checklist

```markdown
准备阶段:
- [ ] server.exe 启动 (localhost:8080)
- [ ] npm run dev 启动 (localhost:3000)
- [ ] PostgreSQL, Redis, Milvus, Ollama 运行中
- [ ] Langfuse cloud 已登录 (另一个 Tab)
- [ ] Stripe 测试模式 (卡号 4242 4242 4242 4242)

Demo 用户:
- [ ] admin@kotoha.com / 253121 (admin)
- [ ] fsnn@vip.qq.com / 123456 (普通用户)
- [ ] fsnn253@qq.com / 123456 (普通用户)
```

## 关键 Talking Points

1. **"这不是一个玩具项目"** — 完整的认证、授权、支付、可观测性、评测体系
2. **"AI 不是噱头"** — 偏好提取 → 精准推荐 → 多步骤编排，形成闭环
3. **"可观测性是基础设施"** — 每个 API 调用、每次 Agent 交互都有 trace
4. **"评测驱动迭代"** — 指标定义清晰，数据可追溯，改进可量化
5. **"工程化思维"** — 模块化架构、测试覆盖、错误处理、安全防范
