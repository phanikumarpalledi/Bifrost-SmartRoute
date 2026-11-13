package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maximhq/bifrost/plugins/smartroute/pkg/api"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/bifrost_plugin"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/router"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

func main() {
	port := flag.String("port", "8080", "Port to serve SmartRoute Gateway API")
	dbPath := flag.String("db", "bifrost_smartroute.db", "SQLite database file path")
	demoMode := flag.Bool("demo", true, "Enable turnkey demo mode with realistic simulated telemetry")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}

	absDB, _ := filepath.Abs(*dbPath)
	fmt.Printf("[SmartRoute] Initializing database at %s...\n", absDB)

	// 1. Initialize SQLite Database
	db, err := storage.NewDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Failed to initialize SQLite database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Seed Initial Models, Providers, and Realistic Logs
	if err := db.SeedInitialData(); err != nil {
		fmt.Printf("[SmartRoute] Note during database seeding: %v\n", err)
	}

	// 3. Initialize Core Subsystems
	failoverMgr := failover.NewManager(3, 30*time.Second)
	costCalc := cost.NewCalculator("gpt-4o")
	engine := router.NewEngine(db, failoverMgr, costCalc)
	plugin := bifrost_plugin.NewSmartRoutePlugin(db, engine, costCalc, failoverMgr, *demoMode)

	// 4. Setup API Routes
	handler := api.NewAPIHandler(db, engine, costCalc, failoverMgr, plugin)
	apiRouter := api.SetupRouter(handler)

	// Check if React dashboard dist exists to serve as unified single binary
	var rootHandler http.Handler = apiRouter
	distCandidates := []string{
		"../../frontend/dist",
		"../frontend/dist",
		"./frontend/dist",
		"C:/Users/PHANI/.gemini/antigravity/scratch/bifrost-smartroute/frontend/dist",
	}

	for _, candidate := range distCandidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			absDist, _ := filepath.Abs(candidate)
			fmt.Printf("[SmartRoute] Serving embedded React dashboard from %s\n", absDist)
			rootHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") || r.URL.Path == "/health" {
					apiRouter.ServeHTTP(w, r)
					return
				}
				// Serve static assets or fallback to index.html for client-side routing
				filePath := filepath.Join(absDist, filepath.Clean(r.URL.Path))
				if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
					http.ServeFile(w, r, filePath)
					return
				}
				http.ServeFile(w, r, filepath.Join(absDist, "index.html"))
			})
			break
		}
	}

	// 5. Welcome Banner
	banner := `
╔═════════════════════════════════════════════════════════════════╗
║   BIFROST SMARTROUTE - INTELLIGENT AI MODEL ROUTING SYSTEM      ║
║   Open-Source Gateway Extension for maximhq/bifrost             ║
╠═════════════════════════════════════════════════════════════════╣
║  • Web Dashboard UI:      http://localhost:%s/               ║
║  • OpenAI Chat Endpoint:  http://localhost:%s/v1/chat/completions║
║  • Playground Route API:  http://localhost:%s/api/smartroute/route║
║  • Dashboard Metrics API: http://localhost:%s/api/smartroute/metrics║
║  • Turnkey Demo Mode:     ACTIVE (Live testing without API keys)║
║  • Database Engine:       SQLite (Turnkey zero-config persistence)║
╚═════════════════════════════════════════════════════════════════╝
`
	fmt.Printf(banner, *port, *port, *port, *port)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      rootHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("[SmartRoute] Gateway service listening on :%s...\n", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
	}
}
