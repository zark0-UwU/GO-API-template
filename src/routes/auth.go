package routes

import (
	"GO-API-template/src/handlers/users"
	"GO-API-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(r *fiber.Router) {
	// Start the route
	route := (*r).Group("/auth")
	// General Middlewares for the route if any

	// Define the subroutes
	route.Get("/login", users.Login)                     // get your jwt
	route.Get("/Renew", middlewares.Auth(), users.Login) // Renew your JWT if not blocked // TODO: renew function

}
