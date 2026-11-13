package classifier

import (
	"testing"
)

func TestClassifier_SimpleQuery(t *testing.T) {
	c := NewClassifier()
	prompt := "What is the capital of France?"
	res := c.Classify(prompt)

	if res.Category != CategorySimple {
		t.Errorf("expected CategorySimple, got %s", res.Category)
	}
	if res.RecommendedTier != TierEconomy {
		t.Errorf("expected TierEconomy, got %s", res.RecommendedTier)
	}
	if res.ComplexityScore > 4.0 {
		t.Errorf("expected low complexity score <= 4.0, got %f", res.ComplexityScore)
	}
}

func TestClassifier_CodingPrompt(t *testing.T) {
	c := NewClassifier()
	prompt := "Write a Go func with a goroutine worker pool using select and sync.WaitGroup:\n```go\nfunc worker(id int) {}\n```"
	res := c.Classify(prompt)

	if res.Category != CategoryCoding {
		t.Errorf("expected CategoryCoding, got %s", res.Category)
	}
	if res.RecommendedTier != TierCoding {
		t.Errorf("expected TierCoding, got %s", res.RecommendedTier)
	}
	if res.ComplexityScore < 4.5 {
		t.Errorf("expected high complexity score >= 4.5, got %f", res.ComplexityScore)
	}
}

func TestClassifier_ReasoningPrompt(t *testing.T) {
	c := NewClassifier()
	prompt := "Evaluate the architectural trade-offs between Paxos and Raft consensus algorithms step by step, and derive the underlying causal mechanism for split-brain prevention."
	res := c.Classify(prompt)

	if res.Category != CategoryReasoning {
		t.Errorf("expected CategoryReasoning, got %s", res.Category)
	}
	if res.RecommendedTier != TierReasoning {
		t.Errorf("expected TierReasoning, got %s", res.RecommendedTier)
	}
	if res.ComplexityScore < 5.0 {
		t.Errorf("expected high complexity score >= 5.0, got %f", res.ComplexityScore)
	}
}

func TestClassifier_EmptyPrompt(t *testing.T) {
	c := NewClassifier()
	res := c.Classify("")

	if res.Category != CategorySimple {
		t.Errorf("expected CategorySimple for empty prompt, got %s", res.Category)
	}
	if res.ComplexityScore != 1.0 {
		t.Errorf("expected complexity score 1.0, got %f", res.ComplexityScore)
	}
}
