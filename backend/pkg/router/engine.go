package router

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/classifier"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

// Decision contains the comprehensive result of intelligent routing
type Decision struct {
	ID               string                 `json:"id"`
	RequestID        string                 `json:"request_id"`
	Prompt           string                 `json:"prompt"`
	TaskCategory     string                 `json:"task_category"`
	ComplexityScore  float64                `json:"complexity_score"`
	EstimatedTokens  int                    `json:"estimated_tokens"`
	SelectedProvider string                 `json:"selected_provider"`
	SelectedModel    string                 `json:"selected_model"`
	FallbackChain    []storage.Model        `json:"fallback_chain"`
	EstimatedCost    float64                `json:"estimated_cost"`
	ExpectedLatency  int                    `json:"expected_latency_ms"`
	BaselineModel    string                 `json:"baseline_model"`
	BaselineCost     float64                `json:"baseline_cost"`
	EstimatedSavings float64                `json:"estimated_savings"`
	SavingsPercent   float64                `json:"savings_percent"`
	Explanation      string                 `json:"explanation"`
	Factors          map[string]interface{} `json:"factors"`
	ManualOverride   bool                   `json:"manual_override"`
	CreatedAt        time.Time              `json:"created_at"`
}

// ScoredCandidate stores evaluation score for a model
type ScoredCandidate struct {
	Model        storage.Model
	UtilityScore float64
	CostScore    float64
	LatencyScore float64
	Reliability  float64
	TierBonus    float64
}

// Engine implements multi-objective AI model routing
type Engine struct {
	mu         sync.RWMutex
	classifier *classifier.Classifier
	costCalc   *cost.Calculator
	failover   *failover.Manager
	db         *storage.DB
}

// NewEngine creates a new routing engine instance
func NewEngine(db *storage.DB, failoverMgr *failover.Manager, costCalc *cost.Calculator) *Engine {
	return &Engine{
		classifier: classifier.NewClassifier(),
		costCalc:   costCalc,
		failover:   failoverMgr,
		db:         db,
	}
}

// Route evaluates prompt complexity, model economics, and health to pick the best model
func (e *Engine) Route(prompt string, overrideModel string) (*Decision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	reqID := uuid.New().String()

	// 1. Classify prompt complexity and intent
	classRes := e.classifier.Classify(prompt)

	// 2. Fetch active routing rule policy
	rule, err := e.db.GetActiveRule()
	if err != nil || rule == nil {
		rule = &storage.RoutingRule{
			Strategy:            "cost_optimized",
			ComplexityThreshold: 4.0,
			BaselineModel:       "gpt-4o",
			CostWeight:          0.55,
			LatencyWeight:       0.25,
			ReliabilityWeight:   0.20,
		}
	}

	// 3. Fetch all active candidate models from DB
	models, err := e.db.GetActiveModels()
	if err != nil || len(models) == 0 {
		return nil, fmt.Errorf("no active models configured in gateway: %w", err)
	}

	// 4. Handle Manual Override if requested
	if strings.TrimSpace(overrideModel) != "" {
		for _, m := range models {
			if strings.EqualFold(m.ModelName, overrideModel) {
				estCost := e.costCalc.CalculateCost(m.ModelName, classRes.EstimatedTokens, int(float64(classRes.EstimatedTokens)*1.5))
				baseCost := e.costCalc.CalculateCost(rule.BaselineModel, classRes.EstimatedTokens, int(float64(classRes.EstimatedTokens)*1.5))
				savings := baseCost - estCost
				if savings < 0 {
					savings = 0
				}
				pct := 0.0
				if baseCost > 0 {
					pct = (savings / baseCost) * 100.0
				}

				return &Decision{
					ID:               uuid.New().String(),
					RequestID:        reqID,
					Prompt:           prompt,
					TaskCategory:     string(classRes.Category),
					ComplexityScore:  classRes.ComplexityScore,
					EstimatedTokens:  classRes.EstimatedTokens,
					SelectedProvider: m.ProviderID,
					SelectedModel:    m.ModelName,
					FallbackChain:    []storage.Model{},
					EstimatedCost:    estCost,
					ExpectedLatency:  m.AvgLatencyMs,
					BaselineModel:    rule.BaselineModel,
					BaselineCost:     baseCost,
					EstimatedSavings: savings,
					SavingsPercent:   pct,
					Explanation:      fmt.Sprintf("Manual override selected by user: %s on provider %s. Intelligent routing bypassed.", m.DisplayName, m.ProviderID),
					Factors:          map[string]interface{}{"manual_override": true},
					ManualOverride:   true,
					CreatedAt:        time.Now(),
				}, nil
			}
		}
	}

	// 5. Score candidates using multi-objective optimization
	var maxPrice float64 = 0.01
	var maxLatency float64 = 1.0

	for _, m := range models {
		combinedPrice := m.InputPricePerM + m.OutputPricePerM
		if combinedPrice > maxPrice {
			maxPrice = combinedPrice
		}
		if float64(m.AvgLatencyMs) > maxLatency {
			maxLatency = float64(m.AvgLatencyMs)
		}
	}

	var scored []ScoredCandidate

	for _, m := range models {
		// Circuit breaker check
		if !e.failover.IsProviderHealthy(m.ProviderID) {
			continue // Skip tripped providers
		}

		// Cost score: cheaper is higher score (0.0 to 1.0)
		combinedPrice := m.InputPricePerM + m.OutputPricePerM
		costScore := 1.0 - (combinedPrice / (maxPrice * 1.05))
		if costScore < 0 {
			costScore = 0
		}

		// Latency score: faster is higher score (0.0 to 1.0)
		latencyScore := 1.0 - (float64(m.AvgLatencyMs) / (maxLatency * 1.1))
		if latencyScore < 0 {
			latencyScore = 0
		}

		// Reliability
		reliability := 1.0

		// Tier matching bonus
		tierBonus := 0.0
		if string(classRes.RecommendedTier) == m.TaskTier {
			tierBonus = 0.40
		} else if classRes.RecommendedTier == classifier.TierEconomy && m.TaskTier == "balanced" {
			tierBonus = 0.15
		}

		// Strict category matching for Coding and Deep Reasoning
		if classRes.Category == classifier.CategoryCoding {
			if m.TaskTier == "coding" {
				tierBonus += 0.50
			} else if m.TaskTier == "economy" {
				tierBonus -= 0.35 // Penalty: economy models often struggle with complex code
			}
		} else if classRes.Category == classifier.CategoryReasoning {
			if m.TaskTier == "reasoning" {
				tierBonus += 0.50
			} else if m.TaskTier == "economy" {
				tierBonus -= 0.40
			}
		}

		utility := (rule.CostWeight * costScore) +
			(rule.LatencyWeight * latencyScore) +
			(rule.ReliabilityWeight * reliability) +
			tierBonus

		scored = append(scored, ScoredCandidate{
			Model:        m,
			UtilityScore: utility,
			CostScore:    costScore,
			LatencyScore: latencyScore,
			Reliability:  reliability,
			TierBonus:    tierBonus,
		})
	}

	if len(scored) == 0 {
		return nil, fmt.Errorf("no healthy candidate models available for routing")
	}

	// Sort by highest utility
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].UtilityScore > scored[j].UtilityScore
	})

	winner := scored[0].Model

	// Construct fallback chain from alternate providers
	var fallbacks []storage.Model
	seenProviders := map[string]bool{winner.ProviderID: true}

	for i := 1; i < len(scored); i++ {
		cand := scored[i].Model
		if !seenProviders[cand.ProviderID] && len(fallbacks) < 2 {
			fallbacks = append(fallbacks, cand)
			seenProviders[cand.ProviderID] = true
		}
	}
	// If only one other provider available, allow secondary model from same provider
	if len(fallbacks) == 0 && len(scored) > 1 {
		fallbacks = append(fallbacks, scored[1].Model)
	}

	// Calculate cost savings vs baseline
	estCompletionTokens := int(float64(classRes.EstimatedTokens) * 1.5)
	if estCompletionTokens < 50 {
		estCompletionTokens = 50
	}

	actCost := e.costCalc.CalculateCost(winner.ModelName, classRes.EstimatedTokens, estCompletionTokens)
	baseCost := e.costCalc.CalculateCost(rule.BaselineModel, classRes.EstimatedTokens, estCompletionTokens)
	savings := baseCost - actCost
	if savings < 0 {
		savings = 0
	}
	savingsPct := 0.0
	if baseCost > 0 {
		savingsPct = (savings / baseCost) * 100.0
	}

	// Build transparent routing explanation
	var explanation string
	switch classRes.Category {
	case classifier.CategorySimple:
		explanation = fmt.Sprintf("Query classified as '%s' (complexity %.1f/10). Selected %s on %s for low latency (~%dms) and high cost efficiency (%.1f%% savings vs %s).",
			classRes.Category, classRes.ComplexityScore, winner.DisplayName, winner.ProviderID, winner.AvgLatencyMs, savingsPct, rule.BaselineModel)
	case classifier.CategoryCoding:
		explanation = fmt.Sprintf("Code syntax detected (complexity %.1f/10). Selected %s on %s optimized for coding accuracy and debugging reasoning.",
			classRes.ComplexityScore, winner.DisplayName, winner.ProviderID)
	case classifier.CategoryReasoning:
		explanation = fmt.Sprintf("Deep analytical or multi-step reasoning query (complexity %.1f/10). Selected %s on %s for high conceptual depth.",
			classRes.ComplexityScore, winner.DisplayName, winner.ProviderID)
	default:
		explanation = fmt.Sprintf("General text query (complexity %.1f/10). Selected %s on %s balancing output quality, speed, and cost.",
			classRes.ComplexityScore, winner.DisplayName, winner.ProviderID)
	}

	factors := map[string]interface{}{
		"category":         classRes.Category,
		"complexity_score": classRes.ComplexityScore,
		"recommended_tier": classRes.RecommendedTier,
		"cost_weight":      rule.CostWeight,
		"latency_weight":   rule.LatencyWeight,
		"rule_strategy":    rule.Strategy,
		"winning_utility":  scored[0].UtilityScore,
	}

	// Persist routing decision in DB asynchronously
	decision := &Decision{
		ID:               uuid.New().String(),
		RequestID:        reqID,
		Prompt:           prompt,
		TaskCategory:     string(classRes.Category),
		ComplexityScore:  classRes.ComplexityScore,
		EstimatedTokens:  classRes.EstimatedTokens,
		SelectedProvider: winner.ProviderID,
		SelectedModel:    winner.ModelName,
		FallbackChain:    fallbacks,
		EstimatedCost:    actCost,
		ExpectedLatency:  winner.AvgLatencyMs,
		BaselineModel:    rule.BaselineModel,
		BaselineCost:     baseCost,
		EstimatedSavings: savings,
		SavingsPercent:   savingsPct,
		Explanation:      explanation,
		Factors:          factors,
		ManualOverride:   false,
		CreatedAt:        time.Now(),
	}

	fbBytes, _ := json.Marshal(fallbacks)
	factBytes, _ := json.Marshal(factors)
	promptSnippet := prompt
	if len(promptSnippet) > 120 {
		promptSnippet = promptSnippet[:120] + "..."
	}

	_ = e.db.SaveRoutingDecision(&storage.RoutingDecision{
		ID:               decision.ID,
		RequestID:        decision.RequestID,
		PromptSnippet:    promptSnippet,
		PromptTokensEst:  decision.EstimatedTokens,
		TaskCategory:     decision.TaskCategory,
		ComplexityScore:  decision.ComplexityScore,
		SelectedProvider: decision.SelectedProvider,
		SelectedModel:    decision.SelectedModel,
		FallbackChain:    string(fbBytes),
		RoutingReason:    decision.Explanation,
		RoutingFactors:   string(factBytes),
		ManualOverride:   false,
		CreatedAt:        decision.CreatedAt,
	})

	return decision, nil
}
