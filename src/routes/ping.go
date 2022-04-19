package routes

import (
	"template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func PingRoute(r *fiber.Router) {
	// Start the route
	route := (*r).Group("/ping")
	// General Middlewares for the route if any

	// Define the subroutes
	route.Get("/", middlewares.Auth(), func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Ping!")
	})
}
