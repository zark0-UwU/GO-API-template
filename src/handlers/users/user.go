package users

import (
	"log"
	"strconv"
	"template/src/models"
	"template/src/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	uid := int(claims["user_id"].(float64))

	if uid != n {
		return false
	}

	return true
}

func validUser(id string, p string) bool {
	var user models.User

	filter := bson.D{
		{"_id", id},
	}
	res := models.UsersCollection.FindOne(services.Mongo.Context, filter)
	err := res.Decode(&user)
	if err != nil {
		return false
	}

	if user.Username == "" {
		return false
	}

	return CheckPasswordHash(p, user.Password)
}

// GetUser get a user
func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	filter := bson.D{
		{"_id", id},
	}
	res := models.UsersCollection.FindOne(services.Mongo.Context, filter)
	err := res.Decode(&user)
	if err != nil {
		log.Println(err.Error())
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Found an error while trying to get the user",
		})
	}

	if user.Username == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No user found with ID", "data": nil})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Product found", "data": user})
}

// CreateUser new user
func CreateUser(c *fiber.Ctx) error {
	type NewUser struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unpocessable Entity, Review your input",
			"data":    err,
			"required": fiber.Map{
				"username": "SuperCoolUsername",
				"email":    "you@example.com",
				"password": "YourSuperSecurePassword",
				"fullName": "yourName",
			},
		})
	}

	hash, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't hash password", "data": err})

	}

	user.Password = hash
	_, err = user.Create()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "data": err})
	}

	newUser := NewUser{
		Email:    user.Email,
		Username: user.Username,
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Created user", "data": newUser})
}

// UpdateUser update user
func UpdateUser(c *fiber.Ctx) error {
	//TODO: add more fields to update (maybe via model.User)
	type UpdateUserInput struct {
		FullName string `json:"fullName"`
	}
	var uui UpdateUserInput
	if err := c.BodyParser(&uui); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unpocessable Entity, Review your input",
			"data":    err,
		})
	}
	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)

	if !validToken(token, id) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unprocessable Entity, Invalid token id",
			"data":    nil,
		})
	}

	userOID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unprocessable Entity, Invalid id",
			"data":    nil,
		})
	}

	filter := bson.M{"_id": userOID}
	update := bson.D{
		{"$set", bson.D{{"fullName", uui.FullName}}},
	}

	_, err = models.UsersCollection.UpdateOne(services.Mongo.Context, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to update the user",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User successfully updated",
		"data":    nil,
	})
}

// DeleteUser delete user
func DeleteUser(c *fiber.Ctx) error {
	type PasswordInput struct {
		Password string `json:"password"`
	}
	var pi PasswordInput
	if err := c.BodyParser(&pi); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)

	if !validToken(token, id) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "error", "message": "Invalid token id", "data": nil})

	}

	if !validUser(id, pi.Password) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "error", "message": "Not valid user", "data": nil})

	}

	userOID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "error", "message": "Invalid id", "data": nil})
	}

	filter := bson.M{"_id": userOID}
	res := models.UsersCollection.FindOne(services.Mongo.Context, filter)
	if res.Err() != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "error", "message": "Invalid id", "data": nil})
	}
	deleted, err := models.UsersCollection.DeleteOne(services.Mongo.Context, filter)

	var user models.User
	res.Decode(&user)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": string(deleted.DeletedCount) + " user successfully deleted",
		"data":    user,
	})
}
