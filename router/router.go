package router

import (
	"push_qq/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api")
	api.Post("/msg", handler.PostMsg) // 发送消息
}
