package routes

import (
	"ToDoList/src/handlers/users"

	"github.com/gofiber/fiber/v2"
)

func UsersRoute(r *fiber.Router) {
	// Start the route
	route := (*r).Group("/users")
	// General Middlewares for the route if any

	// Define the subroutes
	route.Post("/", users.CreateUser)   // Create
	route.Get("/", users.GetUser)       // Read
	route.Patch("/", users.UpdateUser)  // Update
	route.Delete("/", users.DeleteUser) // Delete

}
