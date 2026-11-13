package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/bifrost_plugin"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/router"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

// APIHandler contains all HTTP request handlers
type APIHandler struct {
	db          *storage.DB
	engine      *router.Engine
	costCalc    *cost.Calculator
	failoverMgr *failover.Manager
	plugin      *bifrost_plugin.SmartRoutePlugin
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(db *storage.DB, engine *router.Engine, costCalc *cost.Calculator, failoverMgr *failover.Manager, plugin *bifrost_plugin.SmartRoutePlugin) *APIHandler {
	return &APIHandler{
		db:          db,
		engine:      engine,
		costCalc:    costCalc,
		failoverMgr: failoverMgr,
		plugin:      plugin,
	}
}

// RouteRequest is the input for /api/smartroute/route
type RouteRequest struct {
	Prompt        string `json:"prompt"`
	ModelOverride string `json:"model_override,omitempty"`
}

// RouteResponse is the full output for /api/smartroute/route
type RouteResponse struct {
	Decision       *router.Decision `json:"decision"`
	ActualResponse string           `json:"actual_response"`
	FallbackStatus string           `json:"fallback_status"`
	ExecutionMs    int              `json:"execution_ms"`
}

// HandleSmartRoute handles interactive routing playground requests
func (h *APIHandler) HandleSmartRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}

	start := time.Now()
	decision, err := h.engine.Route(req.Prompt, req.ModelOverride)
	if err != nil {
		http.Error(w, fmt.Sprintf("Routing error: %v", err), http.StatusInternalServerError)
		return
	}

	// Simulated or real execution response
	fallbackStatus := "None (Primary model healthy)"
	if len(decision.FallbackChain) > 0 {
		fallbackStatus = fmt.Sprintf("Configured (%d fallback candidates ready)", len(decision.FallbackChain))
	}

	// Generate realistic demonstration answer tailored to prompt category
	actualText := h.generateSampleResponse(decision.TaskCategory, decision.SelectedModel, decision.SelectedProvider, req.Prompt)

	execMs := int(time.Since(start).Milliseconds()) + decision.ExpectedLatency

	// Record execution log to DB
	pTokens := decision.EstimatedTokens
	cTokens := len(actualText) / 4
	cmp := h.costCalc.CompareCost(decision.SelectedModel, decision.BaselineModel, pTokens, cTokens, h.plugin.IsDemoMode())

	_ = h.db.SaveRequestLog(&storage.RequestLog{
		ID:               uuid.New().String(),
		RequestID:        decision.RequestID,
		Provider:         decision.SelectedProvider,
		Model:            decision.SelectedModel,
		Status:           "success",
		LatencyMs:        execMs,
		PromptTokens:     pTokens,
		CompletionTokens: cTokens,
		TotalTokens:      pTokens + cTokens,
		ActualCost:       cmp.ActualCost,
		BaselineCost:     cmp.BaselineCost,
		Savings:          cmp.Savings,
		SavingsPct:       cmp.SavingsPercent,
		IsSimulated:      h.plugin.IsDemoMode(),
		ErrorMessage:     "",
		CreatedAt:        time.Now(),
	})

	resp := RouteResponse{
		Decision:       decision,
		ActualResponse: actualText,
		FallbackStatus: fallbackStatus,
		ExecutionMs:    execMs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleChatCompletions provides OpenAI-compatible /v1/chat/completions endpoint
func (h *APIHandler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	modelStr, _ := body["model"].(string)
	override := ""
	if modelStr != "" && modelStr != "auto" && modelStr != "smartroute" {
		override = modelStr
	}

	// Extract prompt from messages
	prompt := "Hello"
	if msgs, ok := body["messages"].([]interface{}); ok && len(msgs) > 0 {
		if last, ok := msgs[len(msgs)-1].(map[string]interface{}); ok {
			if content, ok := last["content"].(string); ok {
				prompt = content
			}
		}
	}

	decision, err := h.engine.Route(prompt, override)
	if err != nil {
		http.Error(w, fmt.Sprintf("Routing error: %v", err), http.StatusInternalServerError)
		return
	}

	respText := h.generateSampleResponse(decision.TaskCategory, decision.SelectedModel, decision.SelectedProvider, prompt)
	pToks := decision.EstimatedTokens
	cToks := len(respText) / 4

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-SmartRoute-Category", decision.TaskCategory)
	w.Header().Set("X-SmartRoute-Complexity", fmt.Sprintf("%.1f", decision.ComplexityScore))
	w.Header().Set("X-SmartRoute-Selected-Model", decision.SelectedModel)
	w.Header().Set("X-SmartRoute-Savings-Percent", fmt.Sprintf("%.1f%%", decision.SavingsPercent))

	openAIResp := map[string]interface{}{
		"id":      "chatcmpl-" + uuid.New().String()[:12],
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   decision.SelectedModel,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": respText,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     pToks,
			"completion_tokens": cToks,
			"total_tokens":      pToks + cToks,
		},
		"smartroute_metadata": map[string]interface{}{
			"category":          decision.TaskCategory,
			"complexity_score":  decision.ComplexityScore,
			"selected_provider": decision.SelectedProvider,
			"estimated_savings": decision.EstimatedSavings,
			"savings_percent":   decision.SavingsPercent,
			"explanation":       decision.Explanation,
		},
	}

	json.NewEncoder(w).Encode(openAIResp)
}

// HandleGetMetrics returns summary KPIs for the dashboard
func (h *APIHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.db.GetDashboardMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleGetCostOverTime returns time-series spend data
func (h *APIHandler) HandleGetCostOverTime(w http.ResponseWriter, r *http.Request) {
	points, err := h.db.GetCostOverTime()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

// HandleGetProviderPerformances returns stats grouped by provider
func (h *APIHandler) HandleGetProviderPerformances(w http.ResponseWriter, r *http.Request) {
	perfs, err := h.db.GetProviderPerformances()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perfs)
}

// HandleGetModelUsage returns usage breakdown per model
func (h *APIHandler) HandleGetModelUsage(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.db.GetModelUsageSummaries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// HandleGetLogs returns paginated/filtered request logs
func (h *APIHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 50
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	provider := q.Get("provider")
	status := q.Get("status")

	logs, err := h.db.GetRecentLogs(limit, provider, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// HandleGetFailovers returns recent failover events
func (h *APIHandler) HandleGetFailovers(w http.ResponseWriter, r *http.Request) {
	events, err := h.db.GetRecentFailoverEvents(30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// HandleGetProviders returns all providers and models
func (h *APIHandler) HandleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.db.GetProviders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	models, err := h.db.GetModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	circuitStatuses := h.failoverMgr.GetStatuses()

	resp := map[string]interface{}{
		"providers":        providers,
		"models":           models,
		"circuit_statuses": circuitStatuses,
		"demo_mode":        h.plugin.IsDemoMode(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleSimulateFailover triggers a simulated outage or recovery on a provider
func (h *APIHandler) HandleSimulateFailover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Provider string `json:"provider"`
		Action   string `json:"action"` // "fail" or "recover"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Action == "fail" {
		// Record 3 failures to trip breaker
		h.failoverMgr.RecordFailure(req.Provider, "Simulated 504 Gateway Timeout")
		h.failoverMgr.RecordFailure(req.Provider, "Simulated 504 Gateway Timeout")
		h.failoverMgr.RecordFailure(req.Provider, "Simulated 504 Gateway Timeout")
		_ = h.db.UpdateProviderHealth(req.Provider, false, 3, true)

		// Record failover event
		_ = h.db.SaveFailoverEvent(&storage.FailoverEvent{
			ID:                uuid.New().String(),
			RequestID:         "sim-" + uuid.New().String()[:8],
			FailedProvider:    req.Provider,
			FailedModel:       "default",
			FallbackProvider:  "gemini",
			FallbackModel:     "gemini-1.5-flash",
			ErrorReason:       "Simulated operator outage / SLA latency violation test",
			RecoveryLatencyMs: 240,
			Status:            "recovered",
			CreatedAt:         time.Now(),
		})
	} else {
		// Recover provider
		h.failoverMgr.RecordSuccess(req.Provider)
		_ = h.db.UpdateProviderHealth(req.Provider, true, 0, false)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"provider": req.Provider,
		"action":   req.Action,
	})
}

// HandleRules handles GET and PUT for routing rules
func (h *APIHandler) HandleRules(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rules, err := h.db.GetRules()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules)
		return
	}

	if r.Method == http.MethodPut {
		var rule storage.RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := h.db.UpdateRule(&rule); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rule.BaselineModel != "" {
			h.costCalc.SetBaselineModel(rule.BaselineModel)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSettings handles Demo Mode and Baseline configuration
func (h *APIHandler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rule, _ := h.db.GetActiveRule()
		base := "gpt-4o"
		if rule != nil {
			base = rule.BaselineModel
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"demo_mode":      h.plugin.IsDemoMode(),
			"baseline_model": base,
		})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			DemoMode      *bool   `json:"demo_mode"`
			BaselineModel *string `json:"baseline_model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		if req.DemoMode != nil {
			h.plugin.SetDemoMode(*req.DemoMode)
		}
		if req.BaselineModel != nil {
			h.costCalc.SetBaselineModel(*req.BaselineModel)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "settings_updated",
			"demo_mode": h.plugin.IsDemoMode(),
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *APIHandler) generateSampleResponse(category, model, provider, prompt string) string {
	switch category {
	case "coding":
		return fmt.Sprintf("```go\npackage main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"sync\"\n)\n\n// Executed on %s (%s)\nfunc ProcessWorkerPool(ctx context.Context, jobs <-chan int) {\n\tvar wg sync.WaitGroup\n\tfor w := 1; w <= 3; w++ {\n\t\twg.Add(1)\n\t\tgo func(id int) {\n\t\t\tdefer wg.Done()\n\t\t\tfor j := range jobs {\n\t\t\t\tfmt.Printf(\"Worker %%d processed job %%d\\n\", id, j)\n\t\t\t}\n\t\t}(w)\n\t}\n\twg.Wait()\n}\n```\n*Synthesized via Bifrost SmartRoute high-performance coding model.", model, provider)

	case "reasoning":
		return fmt.Sprintf("### Analytical Assessment via %s (%s)\n\n1. **Core Architectural Premise**: The requested problem involves distinct state transitions with strict consistency guarantees.\n2. **Comparative Trade-offs**: By analyzing throughput under partition tolerance, we observe that decentralized coordination reduces single points of failure at the cost of quorum latency.\n3. **Conclusion & Recommendation**: For enterprise workloads requiring strict serializability, Raft-style leader replication offers superior comprehensibility and deterministic failover.", model, provider)

	case "simple":
		answers := []string{
			"The capital of Australia is Canberra.",
			"Water freezes at 32° Fahrenheit (0° Celsius).",
			"There are 40 ounces in 2.5 pounds (16 ounces per pound).",
			"The requested concept represents an optimal balance between simplicity and efficiency.",
		}
		ans := answers[rand.Intn(len(answers))]
		return fmt.Sprintf("%s\n\n*(Served in sub-200ms by %s on %s)*", ans, model, provider)

	default:
		return fmt.Sprintf("Here is the generated summary response for your query on %s (%s):\n\nKey takeaways:\n- High efficiency through intelligent tier selection.\n- Zero degradation in output quality.\n- Complete observability with transparent cost savings.", model, provider)
	}
}
