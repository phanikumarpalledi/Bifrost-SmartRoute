#!/usr/bin/env python3
"""
Bifrost SmartRoute - Interactive Hackathon Demonstration Script (Python)
Zero external dependencies required (uses standard library urllib).
"""

import json
import sys
import time
import urllib.request
import urllib.error

GATEWAY_URL = "http://localhost:8080"
if len(sys.argv) > 1:
    GATEWAY_URL = sys.argv[1]

# ANSI Colors
CYAN = "\033[96m"
GREEN = "\033[92m"
YELLOW = "\033[93m"
MAGENTA = "\033[95m"
GRAY = "\033[90m"
BOLD = "\033[1m"
RESET = "\033[0m"


def header(title: str):
    print(f"\n{MAGENTA}{'═' * 68}{RESET}")
    print(f"  {CYAN}{BOLD}{title}{RESET}")
    print(f"{MAGENTA}{'═' * 68}{RESET}")


def step(num: str, title: str):
    print(f"\n{YELLOW}[{num}] {BOLD}{title}{RESET}")
    print(f"{GRAY}{'-' * 60}{RESET}")


def api_post(endpoint: str, data: dict) -> dict:
    url = f"{GATEWAY_URL}{endpoint}"
    req_data = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(
        url, data=req_data, headers={"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read().decode("utf-8"))


def api_get(endpoint: str) -> dict:
    url = f"{GATEWAY_URL}{endpoint}"
    with urllib.request.urlopen(url, timeout=10) as resp:
        return json.loads(resp.read().decode("utf-8"))


def main():
    print(f"{CYAN}{BOLD}")
    print("╔══════════════════════════════════════════════════════════════════════╗")
    print("║               BIFROST SMARTROUTE - LIVE DEMO TOUR                    ║")
    print("║       Intelligent AI Model Routing & Cost Optimization System        ║")
    print("╚══════════════════════════════════════════════════════════════════════╝")
    print(f"{RESET}")

    # 1. Health & Metrics Check
    step("1/5", "Inspecting Gateway Health & Historical Analytics...")
    try:
        metrics = api_get("/api/smartroute/metrics")
        print(f"  • Total Requests Processed: {GREEN}{metrics.get('total_requests')}{RESET}")
        print(f"  • Cumulative Expenditure:   {GREEN}${metrics.get('total_expenditure', 0):.4f}{RESET}")
        print(f"  • Unoptimized Baseline:     {YELLOW}${metrics.get('baseline_cost', 0):.4f}{RESET}")
        print(
            f"  • Net Dollar Savings:       {MAGENTA}${metrics.get('estimated_savings', 0):.4f} "
            f"({metrics.get('savings_percent', 0):.1f}% SAVINGS){RESET}"
        )
        print(f"  • Gateway Average Latency:  {GREEN}{metrics.get('avg_response_time_ms', 0):.1f} ms{RESET}")
    except Exception as e:
        print(f"\033[91mError connecting to Bifrost SmartRoute at {GATEWAY_URL}: {e}\033[0m")
        print("Please start the server first: ./bifrost-smartroute.exe -port 8080")
        sys.exit(1)

    # 2. Simple Query Routing
    step("2/5", "Routing Simple Informational Query...")
    prompt1 = "What is the capital of Australia?"
    print(f"  Prompt: {GRAY}\"{prompt1}\"{RESET}")
    res1 = api_post("/api/smartroute/route", {"prompt": prompt1})
    dec1 = res1.get("decision", {})
    print(f"  → Category Detected:  {GREEN}{dec1.get('task_category', '').upper()} (Complexity: {dec1.get('complexity_score')}/10){RESET}")
    print(f"  → Selected Model:     {CYAN}{dec1.get('selected_model')} on {dec1.get('selected_provider')}{RESET}")
    print(f"  → Cost vs Baseline:   {MAGENTA}${dec1.get('estimated_cost'):.6f} vs ${dec1.get('baseline_cost'):.6f} ({dec1.get('savings_percent', 0):.1f}% Savings!){RESET}")
    print(f"  → Justification:      {GRAY}{dec1.get('explanation')}{RESET}")

    # 3. Complex Coding Routing
    step("3/5", "Routing Complex Engineering / Coding Prompt...")
    prompt2 = "Implement an asynchronous connection pool in Go with mutex striping and backoff retry logic"
    print(f"  Prompt: {GRAY}\"{prompt2}\"{RESET}")
    res2 = api_post("/api/smartroute/route", {"prompt": prompt2})
    dec2 = res2.get("decision", {})
    fallbacks = " -> ".join([f.get("model_name", "") for f in dec2.get("fallback_chain", [])])
    print(f"  → Category Detected:  {GREEN}{dec2.get('task_category', '').upper()} (Complexity: {dec2.get('complexity_score')}/10){RESET}")
    print(f"  → Selected Model:     {CYAN}{dec2.get('selected_model')} on {dec2.get('selected_provider')}{RESET}")
    print(f"  → Fallback Chain:     {YELLOW}{fallbacks}{RESET}")
    print(f"  → Justification:      {GRAY}{dec2.get('explanation')}{RESET}")

    # 4. Deep Reasoning / Math Routing
    step("4/5", "Routing Complex Multi-Step Reasoning Puzzle...")
    prompt3 = "Analyze the paradox of thrift under liquidity trap conditions and derive the IS-LM mathematical equilibrium"
    print(f"  Prompt: {GRAY}\"{prompt3}\"{RESET}")
    res3 = api_post("/api/smartroute/route", {"prompt": prompt3})
    dec3 = res3.get("decision", {})
    print(f"  → Category Detected:  {GREEN}{dec3.get('task_category', '').upper()} (Complexity: {dec3.get('complexity_score')}/10){RESET}")
    print(f"  → Selected Model:     {CYAN}{dec3.get('selected_model')} on {dec3.get('selected_provider')}{RESET}")
    print(f"  → Justification:      {GRAY}{dec3.get('explanation')}{RESET}")

    # 5. Circuit Breaker Outage & Self-Healing Failover
    step("5/5", "Simulating Provider Outage & Self-Healing Failover...")
    print(f"  Simulating HTTP 503 on 'anthropic'...")
    try:
        api_post("/api/smartroute/failover/simulate", {"provider": "anthropic", "error_code": 503})
        providers = api_get("/api/smartroute/providers")
        c_status = providers.get("circuit_statuses", {}).get("anthropic", {}).get("state", "unknown")
        print(f"  • Anthropic Circuit Breaker State: {YELLOW}{c_status.upper()}{RESET}")
    except Exception as e:
        print(f"  Note: Simulation endpoint response: {e}")

    print(f"  Submitting request with primary provider under stress...")
    failover_res = api_post("/api/smartroute/route", {"prompt": prompt2})
    print(f"  → Failover Status: {GREEN}{failover_res.get('fallback_status')}{RESET}")
    print(f"  → Active Routing:  {CYAN}{failover_res.get('decision', {}).get('selected_model')}{RESET}")

    header("DEMONSTRATION COMPLETE!")
    print(f"To explore the full interactive dashboard, open:")
    print(f"{CYAN}{BOLD}👉 http://localhost:8080{RESET}\n")


if __name__ == "__main__":
    main()
