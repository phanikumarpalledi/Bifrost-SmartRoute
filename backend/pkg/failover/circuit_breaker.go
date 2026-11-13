package failover

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents breaker state
type CircuitState string

const (
	StateClosed   CircuitState = "closed"    // Normal operation
	StateOpen     CircuitState = "open"      // Failing, rejecting requests
	StateHalfOpen CircuitState = "half_open" // Testing recovery
)

// ProviderStatus holds real-time health metrics
type ProviderStatus struct {
	Provider            string       `json:"provider"`
	State               CircuitState `json:"state"`
	ConsecutiveFailures int          `json:"consecutive_failures"`
	LastFailureTime     time.Time    `json:"last_failure_time"`
	TotalRequests       int          `json:"total_requests"`
	TotalFailures       int          `json:"total_failures"`
	SuccessRate         float64      `json:"success_rate"`
}

// Manager orchestrates circuit breaker states and fallback selection
type Manager struct {
	mu               sync.RWMutex
	failureThreshold int
	recoveryTimeout  time.Duration
	statuses         map[string]*ProviderStatus
}

// NewManager creates a new failover manager
func NewManager(failureThreshold int, recoveryTimeout time.Duration) *Manager {
	if failureThreshold <= 0 {
		failureThreshold = 3
	}
	if recoveryTimeout <= 0 {
		recoveryTimeout = 30 * time.Second
	}

	return &Manager{
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
		statuses:         make(map[string]*ProviderStatus),
	}
}

// EnsureProvider registers a provider if not already tracked
func (m *Manager) EnsureProvider(provider string) *ProviderStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	if status, exists := m.statuses[provider]; exists {
		return status
	}
	status := &ProviderStatus{
		Provider: provider,
		State:    StateClosed,
	}
	m.statuses[provider] = status
	return status
}

// RecordSuccess marks a successful call for a provider
func (m *Manager) RecordSuccess(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.statuses[provider]
	if !exists {
		status = &ProviderStatus{Provider: provider, State: StateClosed}
		m.statuses[provider] = status
	}

	status.TotalRequests++
	status.ConsecutiveFailures = 0
	if status.State == StateHalfOpen || status.State == StateOpen {
		status.State = StateClosed
	}
	if status.TotalRequests > 0 {
		status.SuccessRate = float64(status.TotalRequests-status.TotalFailures) / float64(status.TotalRequests) * 100.0
	}
}

// RecordFailure notes a provider failure and trips the breaker if threshold reached
func (m *Manager) RecordFailure(provider string, errReason string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.statuses[provider]
	if !exists {
		status = &ProviderStatus{Provider: provider, State: StateClosed}
		m.statuses[provider] = status
	}

	status.TotalRequests++
	status.TotalFailures++
	status.ConsecutiveFailures++
	status.LastFailureTime = time.Now()

	tripped := false
	if status.ConsecutiveFailures >= m.failureThreshold {
		status.State = StateOpen
		tripped = true
	}

	if status.TotalRequests > 0 {
		status.SuccessRate = float64(status.TotalRequests-status.TotalFailures) / float64(status.TotalRequests) * 100.0
	}

	return tripped
}

// IsProviderHealthy checks if provider can receive traffic
func (m *Manager) IsProviderHealthy(provider string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.statuses[provider]
	if !exists {
		return true // assume healthy until proven otherwise
	}

	if status.State == StateClosed {
		return true
	}

	// Check if recovery timeout has elapsed to allow half-open probe
	if status.State == StateOpen {
		if time.Since(status.LastFailureTime) > m.recoveryTimeout {
			status.State = StateHalfOpen
			return true
		}
		return false
	}

	return true // Half-open allows single probe
}

// CandidateOption defines an eligible model for fallback
type CandidateOption struct {
	Provider string
	Model    string
	Tier     string
	CostPerM float64
}

// SelectFallback selects an eligible alternative provider/model when primary fails
func (m *Manager) SelectFallback(failedProvider, failedModel string, candidates []CandidateOption) (*CandidateOption, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, cand := range candidates {
		// Do not select the failed provider/model
		if cand.Provider == failedProvider && cand.Model == failedModel {
			continue
		}

		// Check if candidate provider circuit breaker is open
		status, exists := m.statuses[cand.Provider]
		if exists && status.State == StateOpen {
			if time.Since(status.LastFailureTime) <= m.recoveryTimeout {
				continue // Skip unhealthy provider
			}
		}

		// Valid fallback candidate found
		return &cand, nil
	}

	return nil, fmt.Errorf("all backup providers and fallback models are currently unavailable")
}

// GetStatuses returns current health status snapshot
func (m *Manager) GetStatuses() map[string]ProviderStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ProviderStatus, len(m.statuses))
	for k, v := range m.statuses {
		result[k] = *v
	}
	return result
}
