package cost

import (
	"math"
	"strings"
	"sync"
)

// ModelPrice defines pricing rates per 1,000,000 tokens
type ModelPrice struct {
	ModelName       string  `json:"model_name"`
	Provider        string  `json:"provider"`
	InputPricePerM  float64 `json:"input_price_per_m"`  // in USD
	OutputPricePerM float64 `json:"output_price_per_m"` // in USD
}

// CostComparison contains actual vs baseline cost calculation
type CostComparison struct {
	Model            string  `json:"model"`
	BaselineModel    string  `json:"baseline_model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	ActualCost       float64 `json:"actual_cost"`
	BaselineCost     float64 `json:"baseline_cost"`
	Savings          float64 `json:"savings"`
	SavingsPercent   float64 `json:"savings_percent"`
	IsSimulated      bool    `json:"is_simulated"`
	PricingLabel     string  `json:"pricing_label"`
}

// Calculator tracks and evaluates costs across providers
type Calculator struct {
	mu            sync.RWMutex
	prices        map[string]ModelPrice
	baselineModel string
}

// NewCalculator initializes the pricing registry with live standard rates
func NewCalculator(defaultBaseline string) *Calculator {
	if defaultBaseline == "" {
		defaultBaseline = "gpt-4o"
	}

	calc := &Calculator{
		prices:        make(map[string]ModelPrice),
		baselineModel: defaultBaseline,
	}

	// Standard real pricing table (USD per 1M tokens)
	calc.RegisterPrice(ModelPrice{ModelName: "gpt-4o", Provider: "openai", InputPricePerM: 2.50, OutputPricePerM: 10.00})
	calc.RegisterPrice(ModelPrice{ModelName: "gpt-4o-mini", Provider: "openai", InputPricePerM: 0.15, OutputPricePerM: 0.60})
	calc.RegisterPrice(ModelPrice{ModelName: "o1-mini", Provider: "openai", InputPricePerM: 3.00, OutputPricePerM: 12.00})
	calc.RegisterPrice(ModelPrice{ModelName: "claude-3-5-sonnet", Provider: "anthropic", InputPricePerM: 3.00, OutputPricePerM: 15.00})
	calc.RegisterPrice(ModelPrice{ModelName: "claude-3-5-haiku", Provider: "anthropic", InputPricePerM: 0.80, OutputPricePerM: 4.00})
	calc.RegisterPrice(ModelPrice{ModelName: "gemini-1.5-flash", Provider: "gemini", InputPricePerM: 0.075, OutputPricePerM: 0.30})
	calc.RegisterPrice(ModelPrice{ModelName: "gemini-1.5-pro", Provider: "gemini", InputPricePerM: 1.25, OutputPricePerM: 5.00})
	calc.RegisterPrice(ModelPrice{ModelName: "gemini-2.0-flash", Provider: "gemini", InputPricePerM: 0.10, OutputPricePerM: 0.40})

	return calc
}

// RegisterPrice adds or updates a model pricing entry
func (c *Calculator) RegisterPrice(p ModelPrice) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[strings.ToLower(p.ModelName)] = p
}

// SetBaselineModel updates the baseline reference model
func (c *Calculator) SetBaselineModel(modelName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baselineModel = modelName
}

// GetBaselineModel returns current baseline model
func (c *Calculator) GetBaselineModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baselineModel
}

// CalculateCost computes total USD expenditure for given tokens
func (c *Calculator) CalculateCost(modelName string, promptTokens, completionTokens int) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cleanName := strings.ToLower(modelName)
	price, exists := c.prices[cleanName]
	if !exists {
		// Reasonable fallback for unconfigured/mock models: $1.00 / 1M in, $3.00 / 1M out
		price = ModelPrice{InputPricePerM: 1.00, OutputPricePerM: 3.00}
	}

	cost := (float64(promptTokens)*price.InputPricePerM + float64(completionTokens)*price.OutputPricePerM) / 1_000_000.0
	return math.Round(cost*1000000) / 1000000 // 6 decimal precision
}

// CompareCost evaluates actual model cost vs configured baseline model
func (c *Calculator) CompareCost(modelName, overrideBaseline string, promptTokens, completionTokens int, isSimulated bool) *CostComparison {
	actualCost := c.CalculateCost(modelName, promptTokens, completionTokens)

	baseline := overrideBaseline
	if baseline == "" {
		baseline = c.GetBaselineModel()
	}
	baselineCost := c.CalculateCost(baseline, promptTokens, completionTokens)

	savings := baselineCost - actualCost
	if savings < 0 {
		savings = 0.0
	}

	savingsPct := 0.0
	if baselineCost > 0 {
		savingsPct = (savings / baselineCost) * 100.0
	}

	label := "Real Provider Pricing"
	if isSimulated {
		label = "Simulated Benchmark"
	}

	return &CostComparison{
		Model:            modelName,
		BaselineModel:    baseline,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		ActualCost:       actualCost,
		BaselineCost:     baselineCost,
		Savings:          math.Round(savings*1000000) / 1000000,
		SavingsPercent:   math.Round(savingsPct*10) / 10,
		IsSimulated:      isSimulated,
		PricingLabel:     label,
	}
}
