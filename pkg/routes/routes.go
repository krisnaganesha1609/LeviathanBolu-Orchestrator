package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/handlers"
)

type Routes struct {
	UserHandler         handlers.UserHandler
	DeviceHandler       handlers.DeviceHandler
	UserSettingsHandler handlers.UserSettingsHandler

	// ReadinessProbe reports whether downstream dependencies (Postgres,
	// Redis, ...) are reachable. Wired by main.go so this package doesn't
	// need to know about *gorm.DB / *redis.Client directly.
	ReadinessProbe func(c fiber.Ctx) bool
}

func SetupRoutes(app *fiber.App, r Routes) {
	// Liveness: is the process up at all? Always true if we can answer.
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))

	// Readiness: are downstream dependencies (DB, Redis) actually reachable?
	readinessProbe := r.ReadinessProbe
	if readinessProbe == nil {
		readinessProbe = func(fiber.Ctx) bool { return true }
	}
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
		Probe:          readinessProbe,
	}))

	api := app.Group("/api")

	// User routes.
	// NOTE: ":id" and ":email" can't coexist as the last segment of the
	// same path shape ("/users/:id" vs "/users/:email") — most routers
	// (fiber included) key matching on path shape, not parameter name, so
	// the second registration would shadow/conflict with the first.
	// Email lookup gets its own sub-path instead.
	api.Post("/users", r.UserHandler.RegisterUser)
	api.Get("/users/:id", r.UserHandler.GetUserByID)
	api.Put("/users/:id", r.UserHandler.UpdateUser)
	api.Get("/users/by-email/:email", r.UserHandler.GetUserByEmail)

	// Device routes
	api.Post("/devices", r.DeviceHandler.RegisterDevice)
	api.Get("/devices/:user_id", r.DeviceHandler.GetUserDevices)

	// User settings routes
	api.Get("/settings/:user_id", r.UserSettingsHandler.GetUserSettings)
	api.Put("/settings/:user_id", r.UserSettingsHandler.UpdateUserSettings)
}
