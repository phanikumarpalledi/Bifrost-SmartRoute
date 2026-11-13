package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

type BenchmarkConfig struct {
	URL         string
	Concurrency int
	TotalReqs   int
}

type RouteRequest struct {
	Prompt string `json:"prompt"`
}

type RouteResponse struct {
	Decision struct {
		SelectedModel   string  `json:"selected_model"`
		SelectedProvider string  `json:"selected_provider"`
		ComplexityScore float64 `json:"complexity_score"`
		TaskCategory    string  `json:"task_category"`
		EstimatedCost   float64 `json:"estimated_cost"`
		BaselineCost    float64 `json:"baseline_cost"`
		EstimatedSavings float64 `json:"estimated_savings"`
	} `json:"decision"`
	ExecutionMs int `json:"execution_ms"`
}

func main() {
	url := flag.String("url", "http://localhost:8080/api/smartroute/route", "Target SmartRoute URL")
	concurrency := flag.Int("concurrency", 10, "Concurrent workers")
	totalReqs := flag.Int("total", 50, "Total requests to benchmark")
	flag.Parse()

	prompts := []string{
		"What is the population of Tokyo?",
		"Write an efficient quicksort implementation in Go with in-place swapping",
		"Analyze the pros and cons of event-driven microservices vs monoliths for fintech",
		"Convert this JSON object to a typed TypeScript interface",
		"Solve this logic riddle: A farmer has 17 sheep, all but 9 die, how many are left?",
		"Draft a friendly confirmation email for an enterprise SaaS subscription renewal",
	}

	fmt.Println("╔═════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     BIFROST SMARTROUTE - HIGH-THROUGHPUT BENCHMARK TOOL         ║")
	fmt.Printf("║  Target: %-55s ║\n", *url)
	fmt.Printf("║  Total Requests: %-4d | Concurrency Workers: %-3d              ║\n", *totalReqs, *concurrency)
	fmt.Println("╚═════════════════════════════════════════════════════════════════╝")
	fmt.Println("Starting benchmark run...")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
		},
	}

	reqChan := make(chan int, *totalReqs)
	for i := 0; i < *totalReqs; i++ {
		reqChan <- i
	}
	close(reqChan)

	var wg sync.WaitGroup
	var mu sync.Mutex

	durations := make([]time.Duration, 0, *totalReqs)
	modelCounts := make(map[string]int)
	categoryCounts := make(map[string]int)
	var totalCost float64
	var totalBaseline float64
	var totalSavings float64
	successCount := 0
	failCount := 0

	startGlobal := time.Now()

	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for reqIdx := range reqChan {
				prompt := prompts[reqIdx%len(prompts)]
				reqBody, _ := json.Marshal(RouteRequest{Prompt: prompt})

				reqStart := time.Now()
				req, err := http.NewRequest("POST", *url, bytes.NewBuffer(reqBody))
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				dur := time.Since(reqStart)

				if err != nil || resp.StatusCode != http.StatusOK {
					mu.Lock()
					failCount++
					mu.Unlock()
					if resp != nil {
						resp.Body.Close()
					}
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var rResp RouteResponse
				if err := json.Unmarshal(body, &rResp); err == nil {
					mu.Lock()
					durations = append(durations, dur)
					successCount++
					modelCounts[rResp.Decision.SelectedModel]++
					categoryCounts[rResp.Decision.TaskCategory]++
					totalCost += rResp.Decision.EstimatedCost
					totalBaseline += rResp.Decision.BaselineCost
					totalSavings += rResp.Decision.EstimatedSavings
					mu.Unlock()
				} else {
					mu.Lock()
					failCount++
					mu.Unlock()
				}
			}
		}(w)
	}

	wg.Wait()
	totalElapsed := time.Since(startGlobal)

	if len(durations) == 0 {
		fmt.Println("Error: No successful requests recorded.")
		return
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	p50 := durations[len(durations)*50/100]
	p90 := durations[len(durations)*90/100]
	p99 := durations[len(durations)*99/100]

	var sumDuration time.Duration
	for _, d := range durations {
		sumDuration += d
	}
	avgDur := sumDuration / time.Duration(len(durations))
	rps := float64(successCount) / totalElapsed.Seconds()

	savingsPct := 0.0
	if totalBaseline > 0 {
		savingsPct = (totalSavings / totalBaseline) * 100
	}

	fmt.Println("\n══════════════════════ BENCHMARK RESULTS ══════════════════════")
	fmt.Printf("• Total Elapsed Time:   %v\n", totalElapsed)
	fmt.Printf("• Successful Requests:  %d / %d (%.1f%%)\n", successCount, *totalReqs, float64(successCount)/float64(*totalReqs)*100)
	fmt.Printf("• Failed Requests:      %d\n", failCount)
	fmt.Printf("• Throughput:           %.2f req/sec\n", rps)
	fmt.Printf("• Latency Average:      %v\n", avgDur)
	fmt.Printf("• Latency P50:          %v\n", p50)
	fmt.Printf("• Latency P90:          %v\n", p90)
	fmt.Printf("• Latency P99:          %v\n", p99)

	fmt.Println("\n── Category Distribution ──────────────────────────────────────")
	for cat, count := range categoryCounts {
		fmt.Printf("  %-15s: %3d requests (%.1f%%)\n", cat, count, float64(count)/float64(successCount)*100)
	}

	fmt.Println("\n── Model Selection Breakdown ──────────────────────────────────")
	for model, count := range modelCounts {
		fmt.Printf("  %-25s: %3d selections (%.1f%%)\n", model, count, float64(count)/float64(successCount)*100)
	}

	fmt.Println("\n── Cost Optimization Analysis ─────────────────────────────────")
	fmt.Printf("  • Actual Optimized Cost:  $%0.6f\n", totalCost)
	fmt.Printf("  • Unoptimized Baseline:   $%0.6f\n", totalBaseline)
	fmt.Printf("  • Total Cost Savings:     $%0.6f\n", totalSavings)
	fmt.Printf("  • Net Savings Percentage: %.2f%%\n", savingsPct)
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
