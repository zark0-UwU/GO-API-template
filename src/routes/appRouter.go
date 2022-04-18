package routes

import (
	"github.com/gofiber/fiber/v2"
)

// @title Fiber Example API
// @version 1.0
// @description This is a sample swagger for Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
func SetupRoutes(api *fiber.Router) {
	// Middleware if any
	//? wouldn't this section be useless?, since it would be the same as global middlewares

	// Routes /api/v1
	UsersRoute(api) // >> /users
	AuthRoute(api)  // >> /auth

	PingRoute(api) // >> /ping

}
