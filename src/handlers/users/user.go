package users

import (
	"log"
	"strconv"

	"GO-API-template/src/models"
	"GO-API-template/src/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type PasswordInput struct {
	Password string `json:"password"`
}

// returns hashed password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// returns true if the token token has the same id as the one passed
//! this does not check the token to be valid, only validates an id against the token
func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	uid := int(claims["user_id"].(float64))

	return uid == n
}

// checks if the user with the given id exists and the password validates to the given user's
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
// @Summary      Retrieve user data
// @Description  Check api is active
// @security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        uid  query  string  true  "uid string"
// @Router       /users/{uid} [get]
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      404  {object}  interface{}
// @Failure      500  {object}  interface{}
func GetUser(c *fiber.Ctx) error {
	id := c.Params("uid")
	user, err := getUserById(id)
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

// CreateUser register new user
// @Summary      Register endpoint
// @Description  Register a new user
// @Accept       json
// @Produce      json
// @param	registerData body models.User{} true "initial data for the user"
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      422  {object}  interface{}
// @Failure      500  {object}  interface{}
// @Router       /users/ [post]
func CreateUser(c *fiber.Ctx) error {
	var user models.User
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

	// Lock to create onlyregular user by asignin the "user" role
	user.Role = "user"
	// Set the roleID dynamically so it can be para metrized in the future
	err = user.SetRole()

	// check that the username/email is not already being used
	//? move to the user model as func (models.user)checkUnique() bool ?
	uMail, err := getUserByEmail(user.Email)
	uUsername, err := getUserByUsername(user.Username)
	if uMail != nil || uUsername != nil {
		return c.Status(fiber.StatusLocked).JSON(fiber.Map{"status": "error", "message": "Couldn't create user, user with the same username/email already exists", "data": err})
	}

	_, err = user.Create()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "data": err})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Created user", "data": user.Private()})
}

// UpdateUser update user
// @Summary      update user
// @Description  Update user info
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @param	updateUserData body models.User{} true "data to update, currently only allows to update the fullName field"
// @param	uid path string true "User ID"
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      422  {object}  interface{}
// @Failure      500  {object}  interface{}
// @Router       /users/{uid} [patch]
func UpdateUser(c *fiber.Ctx) error {
	//TODO: add more fields to update (maybe via model.User)
	var uui models.User
	if err := c.BodyParser(&uui); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unpocessable Entity, Review your input",
			"data":    err,
		})
	}
	id := c.Params("uid")
	token := c.Locals("user").(*jwt.Token)

	if id != token.Claims.(jwt.MapClaims)["uid"] {
		user, err := getUserById(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to delete the user",
				"data":    nil,
			})
		}
		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to delete the user, your role cant delete other users.",
				"data":    nil,
			})
		}
	}
	//? dangerous function?
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
		{"$set", bson.D{{"fullName", uui.Name}}},
	}

	_, err = models.UsersCollection.UpdateOne(services.Mongo.Context, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to delete the user",
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
// @Summary      delete user
// @Description  delete user completely
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @param	password body PasswordInput{} false "password of the user to delete, not required if user is admin"
// @param	uid path string true "User ID"
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      422  {object}  interface{}
// @Failure      500  {object}  interface{}
// @Router       /users/{uid} [delete]
func DeleteUser(c *fiber.Ctx) error {
	var pi PasswordInput
	if err := c.BodyParser(&pi); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	id := c.Params("uid")
	token := c.Locals("user").(*jwt.Token)
	var user *models.User

	if id != token.Claims.(jwt.MapClaims)["uid"] {
		var err error
		user, err = getUserById(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to delete the user",
				"data":    nil,
			})
		}

		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to delete the user, your role cant delete other users.",
				"data":    nil,
			})
		}
	}

	//? dangerous code?
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

	var deletedUser models.User
	res.Decode(&deletedUser)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": string(deleted.DeletedCount) + " user successfully deleted",
		"data":    deletedUser,
	})
}
