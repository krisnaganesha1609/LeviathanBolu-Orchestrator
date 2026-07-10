package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/MCPServer"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/ToolServices/ehydrotel"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/docs"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/assistant"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/configs"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/database"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/handlers"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/routes"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/session"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// bootstrap wires config, infrastructure connections, repositories,
// services and handlers together, returning the assembled route table
// plus the raw connections so main() can manage their lifecycle (health
// probes + graceful shutdown).
//
// This used to live in an init() function. init() runs implicitly before
// main() — including before any `go test` in this package — so a real
// database connection (with log.Fatalf on failure) was firing during
// tooling/test runs too. Doing it explicitly inside main() keeps that
// side effect where it belongs.
func bootstrap() (routes.Routes, *configs.OrchestratorConfig, *gorm.DB, *redis.Client) {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db := database.ConnectPostgres(cfg)
	database.MigratePostgres(db)
	rdb := database.ConnectRedis(cfg)

	// ── Repositories ──────────────────────────────────────────────────────
	userRepo := repositories.InitUserRepository(db)
	deviceRepo := repositories.InitDeviceRepository(db)
	userSettingsRepo := repositories.InitUserSettingsRepository(db)
	deviceSessionRepo := repositories.InitDeviceSessionRepository(db)

	// ── Services ─────────────────────────────────────────────────────────
	userSettingsService := services.InitUserSettingsService(userSettingsRepo)
	userService := services.InitUserService(userRepo, userSettingsService)
	deviceService := services.InitDeviceService(deviceRepo)
	deviceSessionService := services.InitDeviceSessionService(deviceSessionRepo)

	// ── Handlers ─────────────────────────────────────────────────────────
	userHandlers := handlers.InitUserHandler(userService)
	deviceHandlers := handlers.InitDeviceHandler(deviceService)
	userSettingsHandlers := handlers.InitUserSettingsHandler(userSettingsService)

	// ── Tool Registry + Executor ──────────────────────────────────────────
	registry := MCPServer.NewRegistry()
	registry.Register(&ehydrotel.GetStatusTool{})

	executor := MCPServer.NewExecutor(registry)

	// ── LLM Provider ─────────────────────────────────────────────────────
	var llmProvider llm.Provider
	switch cfg.LLMProvider {
	case "openrouter":
		llmProvider = llm.NewOpenRouterProvider(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		log.Printf("[llm] provider=openrouter model=%s", cfg.OpenRouterModel)
	default: // "gemini"
		p, err := llm.NewGeminiProvider(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel)
		if err != nil {
			log.Fatalf("[llm] gemini init failed: %v", err)
		}
		llmProvider = p
		log.Printf("[llm] provider=gemini model=%s", cfg.GeminiModel)
	}

	// ── Assistant ─────────────────────────────────────────────────────────
	assistantSvc := assistant.NewAssistantService(llmProvider, registry, executor)
	systemPrompt := assistant.BuildSystemPrompt(cfg.AssistantName, cfg.AssistantLanguage, "")

	assistantHandler := handlers.InitAssistantHandler(assistantSvc, systemPrompt)

	// ── Session Store (Redis) ─────────────────────────────────────────────
	sessionStore := session.NewStore(rdb, cfg.SessionTTL)

	sttClient := assistant.NewSTTWorkerClient(cfg.STTWorkerURL)
	ttsClient := assistant.NewTTSWorkerClient(cfg.TTSWorkerURL)

	assistantSvc.SetTTS(ttsClient)

	wakeChecker := &assistant.WakeWordChecker{
		STT:      sttClient,
		Provider: assistant.NewDeviceWakeWordProvider(db),
	}

	// ── WebSocket Handler ─────────────────────────────────────────────────
	wsHandler := &handlers.WSHandler{
		Service:              assistantSvc,
		SessionStore:         sessionStore,
		SystemPrompt:         systemPrompt,
		WakeChecker:          wakeChecker,
		STT:                  sttClient,
		DeviceSessionService: deviceSessionService,
	}

	router := routes.Routes{
		UserHandler:         userHandlers,
		DeviceHandler:       deviceHandlers,
		UserSettingsHandler: userSettingsHandlers,
		AssistantHandler:    assistantHandler,
		WSHandler:           wsHandler,
		ReadinessProbe:      readinessProbe(db, rdb),
	}

	return router, cfg, db, rdb
}

// readinessProbe pings Postgres and Redis with a short timeout so
// /readyz reflects whether the app can actually serve requests, not just
// whether the HTTP server happens to be up.
func readinessProbe(db *gorm.DB, rdb *redis.Client) func(c fiber.Ctx) bool {
	return func(c fiber.Ctx) bool {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			return false
		}
		if rdb.Ping(ctx).Err() != nil {
			return false
		}
		return true
	}
}

func main() {
	fmt.Println("Starting LeviathanBolu Orchestrator...")
	router, cfg, db, rdb := bootstrap()

	app := fiber.New(fiber.Config{
		AppName:      "LeviathanBolu Orchestrator",
		ErrorHandler: globalErrorHandler,
	})

	routes.SetupRoutes(app, router)

	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(docs.SwaggerJSON)
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(scalarHTML())
	})

	// Run the server in a goroutine so we can listen for shutdown signals
	// on the main goroutine below.
	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server stopped unexpectedly: %v", err)
		}
	}()

	gracefulShutdown(app, db, rdb)
}

// gracefulShutdown blocks until SIGINT/SIGTERM is received, then drains
// in-flight requests and closes downstream connections before exiting.
// Without this, a container orchestrator's SIGTERM during a rolling
// deploy would kill in-flight requests and leave DB/Redis connections
// dangling instead of closing them cleanly.
func gracefulShutdown(app *fiber.App, db *gorm.DB, rdb *redis.Client) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	_ = rdb.Close()

	log.Println("shutdown complete")
}

// globalErrorHandler centralizes error -> HTTP response mapping so
// handlers can just `return err` from repository/service calls.
func globalErrorHandler(c fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"status":  fmt.Sprintf("%d", fiberErr.Code),
			"message": fiberErr.Message,
		})
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "404",
			"message": "Requested resource not found",
		})
	}

	log.Printf("[error] %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  "500",
		"message": "Internal server error. Please try again later",
	})
}

// scalarHTML returns the single-page Scalar API docs HTML.
func scalarHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
  <title>LEVIATHANBOLU — API Docs</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script
    id="api-reference"
    data-url="/docs/swagger.json"
    data-configuration='{"theme":"deepSpace","layout":"modern"}'></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
}
