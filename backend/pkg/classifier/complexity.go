package classifier

import (
	"math"
	"regexp"
	"strings"
)

// TaskCategory defines the detected intent of the user prompt
type TaskCategory string

const (
	CategorySimple    TaskCategory = "simple"
	CategoryCoding    TaskCategory = "coding"
	CategoryReasoning TaskCategory = "reasoning"
	CategoryCreative  TaskCategory = "creative"
)

// ModelTier corresponds to the model performance/pricing tier
type ModelTier string

const (
	TierEconomy   ModelTier = "economy"
	TierBalanced  ModelTier = "balanced"
	TierReasoning ModelTier = "reasoning"
	TierCoding    ModelTier = "coding"
)

// ClassificationResult contains analyzed features and scores
type ClassificationResult struct {
	Category        TaskCategory           `json:"category"`
	ComplexityScore float64                `json:"complexity_score"` // 0.0 to 10.0
	EstimatedTokens int                    `json:"estimated_tokens"`
	RecommendedTier ModelTier              `json:"recommended_tier"`
	Features        map[string]interface{} `json:"features"`
	Reasoning       string                 `json:"reasoning"`
}

// Classifier inspects prompts to classify intent and complexity
type Classifier struct {
	codeRegex      *regexp.Regexp
	mathRegex      *regexp.Regexp
	reasoningRegex *regexp.Regexp
	simpleRegex    *regexp.Regexp
}

// NewClassifier creates a new prompt complexity classifier
func NewClassifier() *Classifier {
	return &Classifier{
		codeRegex:      regexp.MustCompile(`(?i)(def\s+|func\s+|function\s+|class\s+|import\s+|SELECT\s+.*FROM|const\s+|let\s+|var\s+|return\s+|console\.log|println|#include|<html|\.py|\.go|\.js|\.ts|\.java|\.cpp|dockerfile|git\s+|curl\s+|api\s+endpoint|python\b|golang\b|javascript\b|typescript\b|rust\b|quicksort|algorithm|sorting|refactor|debug\b|regex\b)`),
		mathRegex:      regexp.MustCompile(`(?i)(\b\d+[\+\-\*\/\^]\d+\b|\\int|\\sum|matrix|eigenvalue|theorem|derivative|integral|differential|proof|polynomial|geometric|calculus|probability\s+distribution)`),
		reasoningRegex: regexp.MustCompile(`(?i)(step\s+by\s+step|evaluate|implications?|consequence|trade-offs?|pros\s+and\s+cons|analyze|architect|deduce|counterexample|paradox|hypothesis|rigorous|causal|underlying\s+mechanism|compare\s+and\s+contrast)`),
		simpleRegex:    regexp.MustCompile(`(?i)^(what\s+is|who\s+is|where\s+is|when\s+did|capital\s+of|translate|define|synonym|antonym|how\s+many\s+days|spell|convert\s+\d+|tell\s+me\s+a\s+joke|hello|hi\b)`),
	}
}

// Classify analyzes the input prompt text and returns complexity metrics
func (c *Classifier) Classify(prompt string) *ClassificationResult {
	trimmed := strings.TrimSpace(prompt)
	if len(trimmed) == 0 {
		return &ClassificationResult{
			Category:        CategorySimple,
			ComplexityScore: 1.0,
			EstimatedTokens: 5,
			RecommendedTier: TierEconomy,
			Features:        map[string]interface{}{"length": 0},
			Reasoning:       "Empty or trivial prompt default to economy tier.",
		}
	}

	// 1. Token Estimation (approx 1 token per 3.8 characters for English/Code)
	words := strings.Fields(trimmed)
	charCount := len(trimmed)
	wordCount := len(words)
	estTokens := int(math.Ceil(float64(charCount) / 3.8))
	if estTokens < wordCount {
		estTokens = int(float64(wordCount) * 1.3)
	}

	// 2. Feature detection
	hasCodeBlock := strings.Contains(trimmed, "```")
	codeKeywordMatches := len(c.codeRegex.FindAllString(trimmed, -1))
	mathMatches := len(c.mathRegex.FindAllString(trimmed, -1))
	reasoningMatches := len(c.reasoningRegex.FindAllString(trimmed, -1))
	isSimpleIntro := c.simpleRegex.MatchString(trimmed)

	features := map[string]interface{}{
		"char_count":            charCount,
		"word_count":            wordCount,
		"estimated_tokens":      estTokens,
		"has_code_block":        hasCodeBlock,
		"code_keywords_count":   codeKeywordMatches,
		"math_terms_count":      mathMatches,
		"reasoning_terms_count": reasoningMatches,
		"is_simple_query":       isSimpleIntro,
	}

	// 3. Compute continuous complexity score (0.0 to 10.0)
	var score float64 = 1.0

	// Length factor (logarithmic scale up to 1000 tokens)
	lengthScore := math.Min(3.0, math.Log2(float64(estTokens)+1.0)*0.45)
	score += lengthScore

	// Code weight
	if hasCodeBlock {
		score += 3.5
	} else if codeKeywordMatches > 0 {
		score += math.Min(3.5, float64(codeKeywordMatches)*1.2)
	}

	// Reasoning weight
	if reasoningMatches > 0 {
		score += math.Min(3.5, float64(reasoningMatches)*1.1)
	}

	// Math weight
	if mathMatches > 0 {
		score += math.Min(2.5, float64(mathMatches)*1.0)
	}

	// Simple penalty
	if isSimpleIntro && estTokens < 50 && codeKeywordMatches == 0 && reasoningMatches == 0 {
		score = math.Max(1.0, score-2.0)
	}

	// Clamp between 1.0 and 10.0
	score = math.Max(1.0, math.Min(10.0, score))
	score = math.Round(score*10) / 10 // round to 1 decimal

	// 4. Determine Category & Tier
	var category TaskCategory
	var tier ModelTier
	var rationale string

	if hasCodeBlock || (codeKeywordMatches >= 2 && codeKeywordMatches > reasoningMatches) {
		category = CategoryCoding
		tier = TierCoding
		rationale = "Prompt contains explicit code syntax, programming patterns, or structural implementation queries."
	} else if reasoningMatches >= 2 || mathMatches >= 2 || (score >= 6.5 && estTokens > 80) {
		category = CategoryReasoning
		tier = TierReasoning
		rationale = "Complex multi-step reasoning, mathematical formulation, or high-depth analytical query."
	} else if codeKeywordMatches >= 1 {
		category = CategoryCoding
		tier = TierCoding
		rationale = "Technical or algorithm query requiring coding models."
	} else if isSimpleIntro || (score < 3.8 && estTokens < 80) {
		category = CategorySimple
		tier = TierEconomy
		rationale = "Concise informational query, lookup, or general greeting suitable for high-speed economical models."
	} else {
		category = CategoryCreative
		tier = TierBalanced
		rationale = "Open-ended, conversational, or general text generation requiring balanced fluency and latency."
	}

	return &ClassificationResult{
		Category:        category,
		ComplexityScore: score,
		EstimatedTokens: estTokens,
		RecommendedTier: tier,
		Features:        features,
		Reasoning:       rationale,
	}
}
