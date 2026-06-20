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
}

func SetupRoutes(app *fiber.App, r Routes) {
	api := app.Group("/api")
	// /api/readyz
	api.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{ResponseFormat: healthcheck.FormatJSON}))

	// User routes
	api.Post("/users", r.UserHandler.RegisterUser)
	api.Get("/users/:id", r.UserHandler.GetUserByID)
	api.Get("/users/:email", r.UserHandler.GetUserByEmail)

	// Device routes
	api.Post("/devices", r.DeviceHandler.RegisterDevice)
	api.Get("/devices/:user_id", r.DeviceHandler.GetUserDevices)

	// User settings routes
	api.Get("/settings/:user_id", r.UserSettingsHandler.GetUserSettings)
	api.Put("/settings/:user_id", r.UserSettingsHandler.UpdateUserSettings)
}
