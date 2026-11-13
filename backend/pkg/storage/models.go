package storage

import "time"

// Provider represents an AI provider configuration
type Provider struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	ProviderType        string    `json:"provider_type"` // openai, gemini, anthropic, mock
	IsActive            bool      `json:"is_active"`
	Healthy             bool      `json:"healthy"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	CircuitBreakerOpen  bool      `json:"circuit_breaker_open"`
	CreatedAt           time.Time `json:"created_at"`
}

// Model represents an AI model with pricing and characteristics
type Model struct {
	ID              string  `json:"id"`
	ProviderID      string  `json:"provider_id"`
	ModelName       string  `json:"model_name"`
	DisplayName     string  `json:"display_name"`
	TaskTier        string  `json:"task_tier"` // economy, balanced, reasoning, coding
	InputPricePerM  float64 `json:"input_price_per_m"`  // USD per 1M tokens
	OutputPricePerM float64 `json:"output_price_per_m"` // USD per 1M tokens
	AvgLatencyMs    int     `json:"avg_latency_ms"`
	ContextWindow   int     `json:"context_window"`
	IsActive        bool    `json:"is_active"`
}

// RoutingDecision captures why and how a request was routed
type RoutingDecision struct {
	ID               string    `json:"id"`
	RequestID        string    `json:"request_id"`
	PromptSnippet    string    `json:"prompt_snippet"`
	PromptTokensEst  int       `json:"prompt_tokens_est"`
	TaskCategory     string    `json:"task_category"` // simple, coding, reasoning, creative, summarization
	ComplexityScore  float64   `json:"complexity_score"` // 0.0 to 10.0
	SelectedProvider string    `json:"selected_provider"`
	SelectedModel    string    `json:"selected_model"`
	FallbackChain    string    `json:"fallback_chain"` // JSON array of fallbacks
	RoutingReason    string    `json:"routing_reason"`
	RoutingFactors   string    `json:"routing_factors"` // JSON details
	ManualOverride   bool      `json:"manual_override"`
	CreatedAt        time.Time `json:"created_at"`
}

// RequestLog captures execution telemetry and cost comparison
type RequestLog struct {
	ID               string    `json:"id"`
	RequestID        string    `json:"request_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	Status           string    `json:"status"` // success, failed, fallback_recovered
	LatencyMs        int       `json:"latency_ms"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	ActualCost       float64   `json:"actual_cost"`
	BaselineCost     float64   `json:"baseline_cost"`
	Savings          float64   `json:"savings"`
	SavingsPct       float64   `json:"savings_pct"`
	IsSimulated      bool      `json:"is_simulated"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// FailoverEvent tracks provider failure and fallback recovery
type FailoverEvent struct {
	ID                string    `json:"id"`
	RequestID         string    `json:"request_id"`
	FailedProvider    string    `json:"failed_provider"`
	FailedModel       string    `json:"failed_model"`
	FallbackProvider  string    `json:"fallback_provider"`
	FallbackModel     string    `json:"fallback_model"`
	ErrorReason       string    `json:"error_reason"`
	RecoveryLatencyMs int       `json:"recovery_latency_ms"`
	Status            string    `json:"status"` // recovered, exhausted
	CreatedAt         time.Time `json:"created_at"`
}

// RoutingRule specifies global and category-specific routing policy
type RoutingRule struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Strategy            string    `json:"strategy"` // cost_optimized, performance_first, balanced, strict_coding
	ComplexityThreshold float64   `json:"complexity_threshold"`
	BaselineModel       string    `json:"baseline_model"`
	CostWeight          float64   `json:"cost_weight"`
	LatencyWeight       float64   `json:"latency_weight"`
	ReliabilityWeight   float64   `json:"reliability_weight"`
	IsActive            bool      `json:"is_active"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// DashboardMetrics aggregates key metrics for the UI dashboard
type DashboardMetrics struct {
	TotalRequests       int     `json:"total_requests"`
	TotalExpenditure    float64 `json:"total_expenditure"`
	BaselineCost        float64 `json:"baseline_cost"`
	EstimatedSavings    float64 `json:"estimated_savings"`
	SavingsPercent      float64 `json:"savings_percent"`
	AvgResponseTimeMs   float64 `json:"avg_response_time_ms"`
	SuccessRate         float64 `json:"success_rate"`
	SuccessfulRequests  int     `json:"successful_requests"`
	FailedRequestsCount int     `json:"failed_requests_count"`
	ActiveProviders     int     `json:"active_providers"`
	ActiveModels        int     `json:"active_models"`
}

// ProviderPerformance tracks latency and reliability per provider
type ProviderPerformance struct {
	Provider     string  `json:"provider"`
	TotalCalls   int     `json:"total_calls"`
	SuccessCalls int     `json:"success_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	TotalTokens  int     `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
}

// ModelUsageSummary tracks volume and cost per model
type ModelUsageSummary struct {
	Model        string  `json:"model"`
	Provider     string  `json:"provider"`
	Tier         string  `json:"tier"`
	RequestCount int     `json:"request_count"`
	TotalTokens  int     `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

// CostOverTimePoint represents expenditure trends
type CostOverTimePoint struct {
	Timestamp    string  `json:"timestamp"`
	ActualCost   float64 `json:"actual_cost"`
	BaselineCost float64 `json:"baseline_cost"`
	Savings      float64 `json:"savings"`
	RequestCount int     `json:"request_count"`
}
