package routes

import (
	"GO-API-template/src/handlers/users"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(r *fiber.Router) {
	// Start the route
	route := (*r).Group("/auth")
	// General Middlewares for the route if any

	// Define the subroutes
	route.Get("/login", users.Login) // get your jwt

}
