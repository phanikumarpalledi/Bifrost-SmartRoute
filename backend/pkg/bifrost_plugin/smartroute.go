package bifrost_plugin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/router"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

// PluginName is the unique identifier for SmartRoute plugin
const PluginName = "smartroute"

// SmartRoutePlugin implements Bifrost's official schemas.LLMPlugin interface
type SmartRoutePlugin struct {
	mu              sync.RWMutex
	name            string
	engine          *router.Engine
	costCalc        *cost.Calculator
	failoverMgr     *failover.Manager
	db              *storage.DB
	demoMode        bool
	activeDecisions map[string]*router.Decision // keyed by request ID
}

// Ensure interface compliance at compile time
var _ schemas.LLMPlugin = (*SmartRoutePlugin)(nil)

// NewSmartRoutePlugin creates a new Bifrost SmartRoute LLMPlugin
func NewSmartRoutePlugin(db *storage.DB, engine *router.Engine, costCalc *cost.Calculator, failoverMgr *failover.Manager, demoMode bool) *SmartRoutePlugin {
	return &SmartRoutePlugin{
		name:            PluginName,
		engine:          engine,
		costCalc:        costCalc,
		failoverMgr:     failoverMgr,
		db:              db,
		demoMode:        demoMode,
		activeDecisions: make(map[string]*router.Decision),
	}
}

// GetName returns the plugin name for Bifrost registration
func (p *SmartRoutePlugin) GetName() string {
	return p.name
}

// Cleanup releases resources on gateway shutdown
func (p *SmartRoutePlugin) Cleanup() error {
	return nil
}

// SetDemoMode toggles simulated demo responses vs real provider calls
func (p *SmartRoutePlugin) SetDemoMode(demo bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.demoMode = demo
}

// IsDemoMode reports current demo mode status
func (p *SmartRoutePlugin) IsDemoMode() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.demoMode
}

// PreRequestHook is the canonical Bifrost hook running ONCE per incoming request.
// It evaluates prompt complexity, selects req.Provider, req.Model, and populates req.Fallbacks.
func (p *SmartRoutePlugin) PreRequestHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) error {
	if req == nil {
		return nil
	}

	// Extract prompt text from request
	promptText := p.extractPrompt(req)

	// Get current request fields using Bifrost's native accessor
	_, currentModel, _ := req.GetRequestFields()

	// Check for manual override or auto routing
	isAuto := currentModel == "" || strings.EqualFold(currentModel, "auto") || strings.EqualFold(currentModel, "smartroute")
	override := ""
	if !isAuto {
		override = currentModel
	}

	// Execute intelligent routing decision
	decision, err := p.engine.Route(promptText, override)
	if err != nil {
		// Log warning and let request proceed with original model
		return nil
	}

	// Mutate request target provider and model in place using Bifrost's setters
	req.SetProvider(schemas.ModelProvider(decision.SelectedProvider))
	req.SetModel(decision.SelectedModel)

	// Configure Bifrost's native Fallbacks array
	if len(decision.FallbackChain) > 0 {
		var bifrostFallbacks []schemas.Fallback
		for _, fb := range decision.FallbackChain {
			bifrostFallbacks = append(bifrostFallbacks, schemas.Fallback{
				Provider: schemas.ModelProvider(fb.ProviderID),
				Model:    fb.ModelName,
			})
		}
		req.SetFallbacks(bifrostFallbacks)
	}

	// Cache decision by request ID for PostLLMHook observability
	reqID := decision.RequestID
	if ctx != nil {
		if ctxID, ok := ctx.Value(schemas.BifrostContextKeyRequestID).(string); ok && ctxID != "" {
			reqID = ctxID
			decision.RequestID = ctxID
		}
	}

	p.mu.Lock()
	p.activeDecisions[reqID] = decision
	p.mu.Unlock()

	return nil
}

// PreLLMHook runs per provider attempt. Used for demo simulation or circuit breaking short-circuits.
func (p *SmartRoutePlugin) PreLLMHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error) {
	if req == nil {
		return req, nil, nil
	}

	p.mu.RLock()
	isDemo := p.demoMode
	p.mu.RUnlock()

	provider, model, _ := req.GetRequestFields()

	// In Demo Mode (or mock provider), construct a realistic response short-circuit without needing paid credentials
	if isDemo || provider == "mock" {
		promptText := p.extractPrompt(req)
		demoResp := p.generateDemoResponse(model, string(provider), promptText)
		return req, &schemas.LLMPluginShortCircuit{
			Response: demoResp,
		}, nil
	}

	return req, nil, nil
}

// PostLLMHook runs after provider execution finishes.
// It records token consumption, calculates exact costs and savings vs baseline, and logs telemetry.
func (p *SmartRoutePlugin) PostLLMHook(ctx *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	var reqID string
	if ctx != nil {
		if id, ok := ctx.Value(schemas.BifrostContextKeyRequestID).(string); ok {
			reqID = id
		}
	}

	p.mu.Lock()
	decision, exists := p.activeDecisions[reqID]
	if exists {
		delete(p.activeDecisions, reqID)
	}
	p.mu.Unlock()

	now := time.Now()

	// Handle Provider Error
	if bifrostErr != nil {
		prov := "unknown"
		model := "unknown"
		if decision != nil {
			prov = decision.SelectedProvider
			model = decision.SelectedModel
		}

		errMsg := "Unknown upstream provider error"
		if bifrostErr.Error != nil && bifrostErr.Error.Message != "" {
			errMsg = bifrostErr.Error.Message
		}

		p.failoverMgr.RecordFailure(prov, errMsg)

		_ = p.db.SaveRequestLog(&storage.RequestLog{
			ID:           uuid.New().String(),
			RequestID:    reqID,
			Provider:     prov,
			Model:        model,
			Status:       "failed",
			LatencyMs:    0,
			PromptTokens: 0,
			ErrorMessage: errMsg,
			CreatedAt:    now,
		})

		return resp, bifrostErr, nil
	}

	// Handle Success Response
	if resp != nil {
		prov := "unknown"
		model := "unknown"
		promptTokens := 0
		compTokens := 0
		latency := 0

		if decision != nil {
			prov = decision.SelectedProvider
			model = decision.SelectedModel
			latency = decision.ExpectedLatency
			promptTokens = decision.EstimatedTokens
			compTokens = int(float64(promptTokens) * 1.5)
		}

		// Inspect concrete response fields if ChatResponse is present
		if resp.ChatResponse != nil {
			if resp.ChatResponse.Model != "" {
				model = resp.ChatResponse.Model
			}
			if resp.ChatResponse.Usage != nil {
				promptTokens = resp.ChatResponse.Usage.PromptTokens
				compTokens = resp.ChatResponse.Usage.CompletionTokens
			}
			if resp.ChatResponse.ExtraFields.RoutingInfo.Provider != "" {
				prov = string(resp.ChatResponse.ExtraFields.RoutingInfo.Provider)
			}
		}

		p.failoverMgr.RecordSuccess(prov)

		// Calculate exact costs & savings
		baselineModel := "gpt-4o"
		if decision != nil && decision.BaselineModel != "" {
			baselineModel = decision.BaselineModel
		}

		cmp := p.costCalc.CompareCost(model, baselineModel, promptTokens, compTokens, p.IsDemoMode())

		status := "success"
		// Detect fallback recovery
		if decision != nil && decision.SelectedModel != model {
			status = "fallback_recovered"
			_ = p.db.SaveFailoverEvent(&storage.FailoverEvent{
				ID:                uuid.New().String(),
				RequestID:         reqID,
				FailedProvider:    decision.SelectedProvider,
				FailedModel:       decision.SelectedModel,
				FallbackProvider:  prov,
				FallbackModel:     model,
				ErrorReason:       "Automatic failover triggered on primary timeout/error",
				RecoveryLatencyMs: latency,
				Status:            "recovered",
				CreatedAt:         now,
			})
		}

		_ = p.db.SaveRequestLog(&storage.RequestLog{
			ID:               uuid.New().String(),
			RequestID:        reqID,
			Provider:         prov,
			Model:            model,
			Status:           status,
			LatencyMs:        latency,
			PromptTokens:     promptTokens,
			CompletionTokens: compTokens,
			TotalTokens:      promptTokens + compTokens,
			ActualCost:       cmp.ActualCost,
			BaselineCost:     cmp.BaselineCost,
			Savings:          cmp.Savings,
			SavingsPct:       cmp.SavingsPercent,
			IsSimulated:      p.IsDemoMode(),
			ErrorMessage:     "",
			CreatedAt:        now,
		})
	}

	return resp, nil, nil
}

// extractPrompt extracts text content from schemas.BifrostRequest
func (p *SmartRoutePlugin) extractPrompt(req *schemas.BifrostRequest) string {
	if req == nil {
		return ""
	}

	// If Chat Request, inspect input messages
	if req.ChatRequest != nil && len(req.ChatRequest.Input) > 0 {
		var parts []string
		for _, m := range req.ChatRequest.Input {
			if m.Content != nil && m.Content.ContentStr != nil {
				parts = append(parts, *m.Content.ContentStr)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
	}

	return "Default query"
}

// generateDemoResponse generates an intelligent simulated response for testing without paid keys
func (p *SmartRoutePlugin) generateDemoResponse(model, provider, prompt string) *schemas.BifrostResponse {
	text := fmt.Sprintf(
		"**[Bifrost SmartRoute Demo Mode]**\n\n"+
			"Your request was intelligently routed to **%s** on **%s**.\n\n"+
			"### Answer Summary\n"+
			"Thank you for your prompt: *\"%s\"*.\n\n"+
			"This response was processed through Bifrost's `PreRequestHook` and `PreLLMHook` pipeline with full token and cost telemetry tracking.",
		model, provider, prompt,
	)

	promptToks := len(prompt) / 4
	if promptToks < 15 {
		promptToks = 15
	}
	compToks := len(text) / 4

	return &schemas.BifrostResponse{
		ChatResponse: &schemas.BifrostChatResponse{
			ID:    "bifrost-demo-" + uuid.New().String()[:8],
			Model: model,
			Choices: []schemas.BifrostResponseChoice{
				{
					Index: 0,
					ChatNonStreamResponseChoice: &schemas.ChatNonStreamResponseChoice{
						Message: &schemas.ChatMessage{
							Role: schemas.ChatMessageRoleAssistant,
							Content: &schemas.ChatMessageContent{
								ContentStr: &text,
							},
						},
					},
					FinishReason: strPtr("stop"),
				},
			},
			Usage: &schemas.BifrostLLMUsage{
				PromptTokens:     promptToks,
				CompletionTokens: compToks,
				TotalTokens:      promptToks + compToks,
			},
		},
	}
}

func strPtr(s string) *string {
	return &s
}
