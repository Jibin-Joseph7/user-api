package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Jibin-Joseph7/user-api/internal/handler"
)

func Register(app *fiber.App, h *handler.UserHandler) {
	users := app.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.List)
	users.Get("/:id", h.Get)
	users.Put("/:id", h.Update)
	users.Delete("/:id", h.Delete)
}