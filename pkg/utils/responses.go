package utils

import "github.com/gofiber/fiber/v3"

// ─── Response Structs ─────────────────────────────────────────────────────────

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreatedResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type EditedResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

type UnauthorizedResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Redirect string `json:"redirect"`
}

// ─── Response Helpers ─────────────────────────────────────────────────────────

// ResponseOK — 200: data fetched successfully.
func ResponseOK(c fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Status:  "200",
		Message: message,
		Data:    data,
	})
}

// ResponseCreated — 201: resource created.
func ResponseCreated(c fiber.Ctx, resource string) error {
	return c.Status(fiber.StatusCreated).JSON(CreatedResponse{
		Status:  "201",
		Message: "Successfully created a " + resource,
	})
}

// ResponseEdited — returns HTTP 200 with a status:"204" body.
// HTTP 204 disallows a response body; we use 200 so the frontend can read the message.
// action: "edited" | "deleted"
func ResponseEdited(c fiber.Ctx, resource, action string) error {
	return c.Status(fiber.StatusOK).JSON(EditedResponse{
		Status:  "204",
		Message: "Successfully " + action + " " + resource,
	})
}

// ResponseNotFound — 404.
func ResponseNotFound(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Status:  "404",
		Message: "Requested Resources Not Found",
	})
}

// ResponseBadRequest — 400: validation errors list.
func ResponseBadRequest(c fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Status:  "400",
		Message: "Invalid/Bad Parameter Request",
		Error:   errors,
	})
}

func ResponseInvalidCredentials(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Status:  "400",
		Message: "Invalid email or password",
	})
}

// ResponseUnauthorized — 401.
func ResponseUnauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(UnauthorizedResponse{
		Status:   "401",
		Message:  "Unauthorized. Please Login!",
		Redirect: "/login",
	})
}

// ResponseForbidden — 403.
func ResponseForbidden(c fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Status:  "403",
		Message: "This resource is forbidden for this role",
	})
}

// ResponseInternalError — 500.
func ResponseInternalError(c fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Status:  "500",
		Message: "Internal server error: " + c.Err().Error() + ". Please try again later",
	})
}
