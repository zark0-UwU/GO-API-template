package auth

import (
	"GO-API-template/src/config"
	"GO-API-template/src/models"
	"GO-API-template/src/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v4"
)

// Login get user and password
// @Summary      login to get the authentication bearer token
// @Description  Get your user's token to acess users only protected routes
// @Accept       json
// @Produce      json
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      400  {object}  interface{}
// @Router       /auth/renew [get]
func Renew(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)

	type LoginInput struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}

	var input LoginInput
	var userData models.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error on login request",
			"data":    err,
			"expected": fiber.Map{
				"identity": "username or password",
				"password": "user password",
			},
		})
	}

	err := userData.Fill(input.Identity, false, true, true)
	if err != nil {
		return stdAuthError(c, "error", "User not found", err)
	}
	// END get userdata from db

	// validate provided credenntials for user
	if !utils.CheckHash(input.Password, userData.Password) {
		return stdAuthError(c, "error", "Invalid password", nil)
	}
	// END validate provided credenntials for user

	// create the claims
	claims := jwt.MapClaims{
		"username": userData.Username,
		"uid":      userData.ID.Hex(),
		"role":     userData.Role,
		"rid":      userData.RoleID.Hex(),
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response
	t, err := token.SignedString([]byte(config.Config("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Login sucessful", "token": t})
}
