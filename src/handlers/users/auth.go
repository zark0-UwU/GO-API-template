package users

import (
	"context"
	"template/src/config"
	"template/src/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// CheckPasswordHash compare password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func getUserByField(filter bson.M) (*models.User, error) {
	var user models.User
	res := models.UsersCollection.FindOne(context.Background(), filter)
	err := res.Err()
	if err != nil {
		return nil, nil
	}
	err = res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserByEmail(e string) (*models.User, error) {
	return getUserByField(bson.M{
		"email": e,
	})
}

func getUserByUsername(u string) (*models.User, error) {
	return getUserByField(bson.M{
		"username": u,
	})
}

// Login get user and password
func Login(c *fiber.Ctx) error {
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

	// get userdata from db
	email, err := getUserByEmail(input.Identity)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Error on email", "data": err})
	}

	if email == nil {
		user, err := getUserByUsername(input.Identity)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Error on username", "data": err})

		}

		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User not found", "data": err})
		}

		userData = *user
	} else {
		userData = *email
	}
	// END get userdata from db

	// validate provided credenntials for user
	if !CheckPasswordHash(input.Password, userData.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid password", "data": nil})
	}
	// END validate provided credenntials for user

	// create the claims
	claims := jwt.MapClaims{
		"username": userData.Username,
		"id":       userData.ID.Hex(),
		"role":     userData.Role,
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
