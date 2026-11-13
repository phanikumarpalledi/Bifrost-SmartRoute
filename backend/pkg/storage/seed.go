package storage

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// SeedInitialData populates initial providers, models, default rules, and historical logs
func (s *DB) SeedInitialData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already seeded
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&count)
	if count > 0 {
		return nil // already seeded
	}

	now := time.Now()

	// 1. Providers
	providers := []Provider{
		{ID: "openai", Name: "OpenAI", ProviderType: "openai", IsActive: true, Healthy: true, ConsecutiveFailures: 0, CircuitBreakerOpen: false, CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "gemini", Name: "Google Gemini", ProviderType: "gemini", IsActive: true, Healthy: true, ConsecutiveFailures: 0, CircuitBreakerOpen: false, CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "anthropic", Name: "Anthropic", ProviderType: "anthropic", IsActive: true, Healthy: true, ConsecutiveFailures: 0, CircuitBreakerOpen: false, CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "mock", Name: "Demo Mock Provider", ProviderType: "mock", IsActive: true, Healthy: true, ConsecutiveFailures: 0, CircuitBreakerOpen: false, CreatedAt: now.Add(-72 * time.Hour)},
	}

	for _, p := range providers {
		s.db.Exec(`INSERT INTO providers (id, name, provider_type, is_active, healthy, consecutive_failures, circuit_breaker_open, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.Name, p.ProviderType, p.IsActive, p.Healthy, p.ConsecutiveFailures, p.CircuitBreakerOpen, p.CreatedAt)
	}

	// 2. Models
	models := []Model{
		{ID: "m-gpt-4o", ProviderID: "openai", ModelName: "gpt-4o", DisplayName: "GPT-4o (Omni)", TaskTier: "reasoning", InputPricePerM: 2.50, OutputPricePerM: 10.00, AvgLatencyMs: 650, ContextWindow: 128000, IsActive: true},
		{ID: "m-gpt-4o-mini", ProviderID: "openai", ModelName: "gpt-4o-mini", DisplayName: "GPT-4o Mini", TaskTier: "economy", InputPricePerM: 0.15, OutputPricePerM: 0.60, AvgLatencyMs: 240, ContextWindow: 128000, IsActive: true},
		{ID: "m-o1-mini", ProviderID: "openai", ModelName: "o1-mini", DisplayName: "OpenAI o1 Mini", TaskTier: "reasoning", InputPricePerM: 3.00, OutputPricePerM: 12.00, AvgLatencyMs: 1400, ContextWindow: 128000, IsActive: true},
		{ID: "m-claude-3-5-sonnet", ProviderID: "anthropic", ModelName: "claude-3-5-sonnet", DisplayName: "Claude 3.5 Sonnet", TaskTier: "coding", InputPricePerM: 3.00, OutputPricePerM: 15.00, AvgLatencyMs: 780, ContextWindow: 200000, IsActive: true},
		{ID: "m-claude-3-5-haiku", ProviderID: "anthropic", ModelName: "claude-3-5-haiku", DisplayName: "Claude 3.5 Haiku", TaskTier: "economy", InputPricePerM: 0.80, OutputPricePerM: 4.00, AvgLatencyMs: 220, ContextWindow: 200000, IsActive: true},
		{ID: "m-gemini-1-5-flash", ProviderID: "gemini", ModelName: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash", TaskTier: "economy", InputPricePerM: 0.075, OutputPricePerM: 0.30, AvgLatencyMs: 180, ContextWindow: 1000000, IsActive: true},
		{ID: "m-gemini-1-5-pro", ProviderID: "gemini", ModelName: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro", TaskTier: "reasoning", InputPricePerM: 1.25, OutputPricePerM: 5.00, AvgLatencyMs: 820, ContextWindow: 2000000, IsActive: true},
		{ID: "m-gemini-2-0-flash", ProviderID: "gemini", ModelName: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash", TaskTier: "balanced", InputPricePerM: 0.10, OutputPricePerM: 0.40, AvgLatencyMs: 190, ContextWindow: 1000000, IsActive: true},
	}

	for _, m := range models {
		s.db.Exec(`INSERT INTO models (id, provider_id, model_name, display_name, task_tier, input_price_per_m, output_price_per_m, avg_latency_ms, context_window, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			m.ID, m.ProviderID, m.ModelName, m.DisplayName, m.TaskTier, m.InputPricePerM, m.OutputPricePerM, m.AvgLatencyMs, m.ContextWindow, m.IsActive)
	}

	// 3. Routing Rules
	rules := []RoutingRule{
		{ID: "rule-cost-opt", Name: "Smart Cost Optimizer (Recommended)", Strategy: "cost_optimized", ComplexityThreshold: 4.0, BaselineModel: "gpt-4o", CostWeight: 0.55, LatencyWeight: 0.25, ReliabilityWeight: 0.20, IsActive: true, UpdatedAt: now},
		{ID: "rule-balanced", Name: "Balanced Performance & Cost", Strategy: "balanced", ComplexityThreshold: 5.0, BaselineModel: "gpt-4o", CostWeight: 0.35, LatencyWeight: 0.40, ReliabilityWeight: 0.25, IsActive: false, UpdatedAt: now},
		{ID: "rule-latency-opt", Name: "Ultra Low-Latency Priority", Strategy: "performance_first", ComplexityThreshold: 6.0, BaselineModel: "gpt-4o", CostWeight: 0.10, LatencyWeight: 0.70, ReliabilityWeight: 0.20, IsActive: false, UpdatedAt: now},
	}

	for _, r := range rules {
		s.db.Exec(`INSERT INTO routing_rules (id, name, strategy, complexity_threshold, baseline_model, cost_weight, latency_weight, reliability_weight, is_active, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.ID, r.Name, r.Strategy, r.ComplexityThreshold, r.BaselineModel, r.CostWeight, r.LatencyWeight, r.ReliabilityWeight, r.IsActive, r.UpdatedAt)
	}

	// 4. Seed 60 Realistic Historical Logs over last 24 hours
	categories := []string{"simple", "coding", "reasoning", "creative"}
	promptExamples := map[string][]string{
		"simple": {
			"What is the capital of Australia?",
			"Explain the word 'serendipity' in one sentence.",
			"How many ounces are in 2.5 pounds?",
			"Translate 'Good morning, how are you?' to Japanese.",
			"What is the freezing point of water in Fahrenheit?",
			"List 3 high-potassium fruits.",
		},
		"coding": {
			"Write a Go goroutine worker pool with cancellation context.",
			"Fix this SQL query with index optimization for user joins.",
			"Create a React custom hook for debouncing search input.",
			"Implement an LRU cache with O(1) get and put in Python.",
			"Write a Dockerfile for multi-stage Go build with Alpine.",
		},
		"reasoning": {
			"Analyze the trade-offs between Paxos and Raft consensus protocols.",
			"Evaluate the macroeconomic implications of inverted yield curves on commercial banking.",
			"Formulate a formal proof that the halting problem is undecidable in Turing machines.",
			"Compare transformer self-attention complexity against linear state space models (Mamba).",
		},
		"creative": {
			"Draft an executive summary announcement for our Series B funding round.",
			"Summarize this 10-page research paper on multimodal agent systems.",
			"Write a creative product release blog post highlighting zero-latency AI routing.",
		},
	}

	r := rand.New(rand.NewSource(42))

	for i := 0; i < 60; i++ {
		reqID := uuid.New().String()
		cat := categories[r.Intn(len(categories))]
		examples := promptExamples[cat]
		prompt := examples[r.Intn(len(examples))]

		var selectedProv, selectedModel string
		var latency int
		var pTokens, cTokens int
		var complexityScore float64
		var reason string

		switch cat {
		case "simple":
			complexityScore = 1.2 + r.Float64()*1.8
			if r.Float64() < 0.6 {
				selectedProv = "gemini"
				selectedModel = "gemini-1.5-flash"
				latency = 140 + r.Intn(90)
			} else {
				selectedProv = "openai"
				selectedModel = "gpt-4o-mini"
				latency = 200 + r.Intn(100)
			}
			pTokens = 15 + r.Intn(30)
			cTokens = 30 + r.Intn(60)
			reason = "Low complexity query routed to economy tier for maximum cost savings with sub-250ms latency."

		case "coding":
			complexityScore = 6.2 + r.Float64()*2.2
			if r.Float64() < 0.7 {
				selectedProv = "anthropic"
				selectedModel = "claude-3-5-sonnet"
				latency = 650 + r.Intn(300)
			} else {
				selectedProv = "openai"
				selectedModel = "gpt-4o"
				latency = 580 + r.Intn(220)
			}
			pTokens = 60 + r.Intn(120)
			cTokens = 250 + r.Intn(400)
			reason = "Code generation request requiring syntactic accuracy and context comprehension; routed to top coding model."

		case "reasoning":
			complexityScore = 7.5 + r.Float64()*2.1
			if r.Float64() < 0.5 {
				selectedProv = "openai"
				selectedModel = "gpt-4o"
				latency = 600 + r.Intn(350)
			} else {
				selectedProv = "gemini"
				selectedModel = "gemini-1.5-pro"
				latency = 720 + r.Intn(300)
			}
			pTokens = 90 + r.Intn(180)
			cTokens = 350 + r.Intn(600)
			reason = "Deep analytical prompt with high logical dependencies; selected high-reasoning frontier model."

		case "creative":
			complexityScore = 3.5 + r.Float64()*2.5
			selectedProv = "gemini"
			selectedModel = "gemini-2.0-flash"
			latency = 190 + r.Intn(120)
			pTokens = 50 + r.Intn(100)
			cTokens = 180 + r.Intn(250)
			reason = "Summarization/creative generation routed to balanced high-throughput model."
		}

		totTokens := pTokens + cTokens

		// Pricing calculation:
		// Baseline: GPT-4o ($2.50 input, $10.00 output per 1M)
		baseCost := (float64(pTokens)*2.50 + float64(cTokens)*10.00) / 1_000_000.0

		// Actual cost lookup
		var inP, outP float64
		switch selectedModel {
		case "gemini-1.5-flash":
			inP, outP = 0.075, 0.30
		case "gpt-4o-mini":
			inP, outP = 0.15, 0.60
		case "gemini-2.0-flash":
			inP, outP = 0.10, 0.40
		case "gemini-1.5-pro":
			inP, outP = 1.25, 5.00
		case "claude-3-5-sonnet":
			inP, outP = 3.00, 15.00
		default: // gpt-4o
			inP, outP = 2.50, 10.00
		}
		actCost := (float64(pTokens)*inP + float64(cTokens)*outP) / 1_000_000.0
		savings := baseCost - actCost
		if savings < 0 {
			savings = 0
		}
		savingsPct := 0.0
		if baseCost > 0 {
			savingsPct = (savings / baseCost) * 100.0
		}

		timestamp := now.Add(-time.Duration(60-i) * 23 * time.Minute)

		status := "success"
		var isFailover bool
		if i == 14 || i == 37 || i == 52 { // 3 simulated failover records
			status = "fallback_recovered"
			isFailover = true
		}

		// Fallback JSON
		fallbacks := []map[string]string{
			{"provider": "gemini", "model": "gemini-1.5-flash"},
			{"provider": "openai", "model": "gpt-4o-mini"},
		}
		fbBytes, _ := json.Marshal(fallbacks)

		factors := map[string]interface{}{
			"category":         cat,
			"complexity_score": complexityScore,
			"cost_weight":      0.55,
			"latency_weight":   0.25,
			"health_score":     0.99,
		}
		factorsBytes, _ := json.Marshal(factors)

		s.db.Exec(`INSERT INTO routing_decisions (id, request_id, prompt_snippet, prompt_tokens_est, task_category, complexity_score, selected_provider, selected_model, fallback_chain, routing_reason, routing_factors, manual_override, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), reqID, prompt, pTokens, cat, complexityScore, selectedProv, selectedModel, string(fbBytes), reason, string(factorsBytes), false, timestamp)

		s.db.Exec(`INSERT INTO request_logs (id, request_id, provider, model, status, latency_ms, prompt_tokens, completion_tokens, total_tokens, actual_cost, baseline_cost, savings, savings_pct, is_simulated, error_message, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), reqID, selectedProv, selectedModel, status, latency, pTokens, cTokens, totTokens, actCost, baseCost, savings, savingsPct, true, "", timestamp)

		if isFailover {
			s.db.Exec(`INSERT INTO failover_events (id, request_id, failed_provider, failed_model, fallback_provider, fallback_model, error_reason, recovery_latency_ms, status, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				uuid.New().String(), reqID, "anthropic", "claude-3-5-sonnet", selectedProv, selectedModel, "HTTP 504 Gateway Timeout: provider upstream latency exceeded 3000ms SLA", latency+180, "recovered", timestamp)
		}
	}

	fmt.Println("Initial data seeding completed successfully with 60 realistic request logs.")
	return nil
}
