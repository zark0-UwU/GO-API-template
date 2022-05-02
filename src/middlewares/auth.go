package middlewares

import (
	"GO-API-template/src/config"
	stdMsg "GO-API-template/src/helpers/stdMessages"
	"GO-API-template/src/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Protected protect routes
func Auth() fiber.Handler {

	return jwtware.New(jwtware.Config{
		SigningKey:     []byte(config.Config("JWT_SECRET")),
		ErrorHandler:   jwtError,
		SuccessHandler: TokenCheck,
	})
}

// ckecks a token to be valid against it's user
func TokenCheck(c *fiber.Ctx) error {

	// Get Token of the reader's user
	token := c.Locals("user").(*jwt.Token)
	uid := fmt.Sprintf("%v", token.Claims.(jwt.MapClaims)["uid"])

	var user models.User
	userID, err := primitive.ObjectIDFromHex(uid)
	user.ID = userID
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(stdMsg.ErrorDefault("An error ocurred while Authenticating the request", nil))
	}

	err = user.LoadTokens()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(stdMsg.ErrorDefault("An error ocurred while Authenticating the request", nil))
	}

	allowed := user.BlockedTokens[token.Raw]
	if !allowed {
		return jwtError(c, nil)
	}

	return c.Next()
}

func jwtError(c *fiber.Ctx, err error) error {
	const ErrJWTMissingOrMalformed = "Missing or malformed JWT"
	if err.Error() == ErrJWTMissingOrMalformed {
		return c.Status(fiber.StatusUnauthorized).
			JSON(stdMsg.ErrorDefault(ErrJWTMissingOrMalformed, nil))
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(stdMsg.ErrorDefault("Invalid or expired JWT", nil))
}
