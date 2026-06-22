package routes

import (
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/handlers"
)

type Routes struct {
	UserHandler         handlers.UserHandler
	DeviceHandler       handlers.DeviceHandler
	UserSettingsHandler handlers.UserSettingsHandler
	AssistantHandler    handlers.AssistantHandler
	WSHandler           *handlers.WSHandler

	// ReadinessProbe reports whether Postgres + Redis are reachable.
	ReadinessProbe func(c fiber.Ctx) bool
}

func SetupRoutes(app *fiber.App, r Routes) {
	// ── Health ────────────────────────────────────────────────────────────
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))

	readinessProbe := r.ReadinessProbe
	if readinessProbe == nil {
		readinessProbe = func(fiber.Ctx) bool { return true }
	}
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
		Probe:          readinessProbe,
	}))

	// ── WebSocket upgrade middleware ───────────────────────────────────────
	// Must be registered BEFORE the WS routes, scoped to /ws so it doesn't
	// intercept normal HTTP routes.
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// ── WebSocket routes ──────────────────────────────────────────────────
	// ws://host/ws/assistant/:device_id
	// Flutter connects here once and streams multiple turns.
	app.Get("/ws/assistant/:device_id", websocket.New(r.WSHandler.HandleWS))

	// ── REST API ──────────────────────────────────────────────────────────
	api := app.Group("/api")

	// Users
	api.Post("/users", r.UserHandler.RegisterUser)
	api.Get("/users/:id", r.UserHandler.GetUserByID)
	api.Put("/users/:id", r.UserHandler.UpdateUser)
	api.Get("/users/by-email/:email", r.UserHandler.GetUserByEmail)

	// Devices
	api.Post("/devices", r.DeviceHandler.RegisterDevice)
	api.Get("/devices/:user_id", r.DeviceHandler.GetUserDevices)

	// User settings
	api.Get("/settings/:user_id", r.UserSettingsHandler.GetUserSettings)
	api.Put("/settings/:user_id", r.UserSettingsHandler.UpdateUserSettings)

	// Assistant — HTTP fallback (useful for testing without a WS client)
	api.Post("/assistant/chat", r.AssistantHandler.Chat)
}
