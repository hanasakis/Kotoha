package observability

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/pkg/ollama"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type ChatRequest struct {
	Message string           `json:"message" binding:"required"`
	History []ollama.Message `json:"history"`
}

// @Summary      Chat with AI shopping assistant
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        body body ChatRequest true "Chat message and history"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /agent/chat [post]
func (h *Handler) Chat(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	output, err := h.svc.Chat(c.Request.Context(), userID, req.History, agent.ChatInput{Message: req.Message})
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "agent.error")
		return
	}

	resp.Success(c, gin.H{"reply": output.Reply})
}

// @Summary      Get business metrics
// @Tags         observability
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /metrics [get]
func (h *Handler) Metrics(c *gin.Context) {
	metrics := h.svc.GetMetrics()
	resp.Success(c, gin.H{"metrics": metrics})
}

func (h *Handler) TrackSearch(score float64) { h.svc.TrackSearch(score) }

func (h *Handler) TrackCartAdd() { h.svc.TrackCartAdd() }

func (h *Handler) TrackOrderCreated() { h.svc.TrackOrderCreated() }

// @Summary      Save an evaluation run
// @Tags         observability
// @Accept       json
// @Produce      json
// @Param        body body EvalRun true "Evaluation run data"
// @Security     BearerAuth
// @Success      201 {object} EvalRun
// @Failure      400 {object} map[string]interface{}
// @Router       /eval/runs [post]
func (h *Handler) SaveEvalRun(c *gin.Context) {
	var run EvalRun
	if err := c.ShouldBindJSON(&run); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	if err := h.svc.SaveEvalRun(&run); err != nil {
		resp.InternalError(c)
		return
	}
	// Push to Langfuse as score
	h.svc.PushEvalScore(run.Category, "accuracy", run.Accuracy,
		"passed: "+strconv.Itoa(run.Passed)+"/"+strconv.Itoa(run.TotalTests))
	resp.Created(c, run)
}

// @Summary      List evaluation runs
// @Tags         observability
// @Produce      json
// @Param        category query string false "Filter by category"
// @Param        limit    query int    false "Max results" default(20)
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /eval/runs [get]
func (h *Handler) ListEvalRuns(c *gin.Context) {
	category := c.DefaultQuery("category", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	runs, err := h.svc.ListEvalRuns(category, limit)
	if err != nil {
		resp.InternalError(c)
		return
	}
	if runs == nil {
		runs = []EvalRun{}
	}
	resp.Success(c, gin.H{"eval_runs": runs})
}
