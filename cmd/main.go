package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/configs"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/database"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/handlers"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/repositories"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/routes"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/services"
)

var router routes.Routes

func init() {
	fmt.Println("Starting LeviathanBolu Orchestrator...")
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db := database.ConnectPostgres(cfg)
	database.MigratePostgres(db)

	// Init Repositories

	userRepo := repositories.InitUserRepository(db)
	deviceRepo := repositories.InitDeviceRepository(db)
	userSettingsRepo := repositories.InitUserSettingsRepository(db)

	// Init Services
	userSettingsService := services.InitUserSettingsService(userSettingsRepo)
	userService := services.InitUserService(userRepo, userSettingsService)
	deviceService := services.InitDeviceService(deviceRepo)

	// Store services in app context for handlers to use
	userHandlers := handlers.InitUserHandler(userService)
	deviceHandlers := handlers.InitDeviceHandler(deviceService)
	userSettingsHandlers := handlers.InitUserSettingsHandler(userSettingsService)

	router = routes.Routes{
		UserHandler:         userHandlers,
		DeviceHandler:       deviceHandlers,
		UserSettingsHandler: userSettingsHandlers,
	}
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "LeviathanBolu Orchestrator",
		ErrorHandler: globalErrorHandler,
	})
	routes.SetupRoutes(app, router)
	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(scalarHTML())
	})
	app.Listen(":8009")
}

func globalErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"status":  fmt.Sprintf("%d", code),
		"message": err.Error(),
	})
}

// scalarHTML returns the single-page Scalar API docs HTML.
func scalarHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
  <title>PMI Blood Stock — API Docs</title>
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
