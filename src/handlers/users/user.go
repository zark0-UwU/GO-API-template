package users

import (
	"context"
	"fmt"
	"strconv"

	"GO-API-template/src/config"
	stdMsg "GO-API-template/src/helpers/stdMessages"
	"GO-API-template/src/models"
	"GO-API-template/src/services"
	"GO-API-template/src/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PasswordInput struct {
	Password string `json:"password"`
}

type rangeUser struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Amount  int         `json:"amount"`
	Offset  int         `json:"offset"`
	Limmit  int         `json:"limmit"`
	Next    string      `json:"next"`
	Users   interface{} `json:"users"`
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
	// Identity of the user to get data from
	identity := c.Params("uid")
	// get the data of the user we want to get data from
	var user models.User
	err := user.Fill(identity, true, true, false)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No user found with ID", "data": nil})
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Found an error while trying to get the user",
		})
	}

	// Get Token of the reader's user
	token := c.Locals("user").(*jwt.Token)
	editorUID := fmt.Sprintf("%v", token.Claims.(jwt.MapClaims)["uid"])

	// Get the reader's data
	var editorUser models.User
	editorUser.Fill(editorUID, true, false, false)
	var editorRole models.Role
	editorRole.Fill(editorUser.RoleID.Hex(), true, false)

	// check what the user is authorised to get and return that
	if user.ID.Hex() == editorUID {
		// User owner
		return c.JSON(fiber.Map{"status": "success", "message": "User found", "user": user.Private()})
	}
	// Parametrized permissons
	if !editorRole.Permissons.UsersAdmin {
		// Any regular user:
		return c.JSON(fiber.Map{"status": "success", "message": "User found", "user": user.Public()})
	}
	// User admins
	return c.JSON(fiber.Map{"status": "success", "message": "User found", "user": user})
}

// GetUsers get the users list
// @Summary      Retrieve users list
// @Description  Retrieve the users id's list
// @security     BearerAuth
// @Accept       json
// @Produce      json
// @Router       /users [get]
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      404  {object}  interface{}
// @Failure      500  {object}  interface{}
func GetUsers(c *fiber.Ctx) error {
	offset, offsetErr := strconv.Atoi(c.Query("o", "0"))
	limmit, limmitErr := strconv.Atoi(c.Query("l", "10"))
	if (offset - limmit) > 100 {
		limmit = offset + 100
	}
	if offsetErr != nil || limmitErr != nil {
		offset = 0
		limmit = 10
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Found an error while parsing your input, review your 'o' and 'l' query params",
		})
	}

	projection := bson.M{"_id": 1, "username": 1, "role": 1}
	cursor, err := models.UsersCollection.Find(
		context.Background(),
		bson.D{},
		options.Find().SetSkip(int64(offset)).SetLimit(int64(limmit)).SetProjection(projection))
	if err != nil {

	}

	var users []models.UserMinimal
	cursor.All(context.Background(), &users)

	r := limmit - offset
	next := ""
	if r <= len(users) {
		next = c.BaseURL() + config.BasePath + fmt.Sprintf("/users?o=%v&l=%v", offset+r, limmit+r)
	}

	return c.Status(fiber.StatusOK).JSON(rangeUser{
		Status:  "success",
		Offset:  offset,
		Limmit:  limmit,
		Amount:  len(users),
		Next:    next,
		Message: "Sucessfuly found users",
		Users:   users,
	})
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

	unique, err := user.CheckUnique()
	if err != nil {
		return c.Status(500).JSON(stdMsg.ErrorDefault(
			"An error ocured while checking if username and email are unique",
			err,
		))
	}
	if !unique {
		return c.Status(fiber.StatusConflict).JSON(stdMsg.ErrorDefault(
			"Specified username or email is already being used",
			err,
		))
	}

	hash, err := utils.Hash(user.Password)
	if err != nil {
		return c.Status(500).JSON(stdMsg.ErrorDefault("Couldn't hash password", err))

	}

	user.Password = hash

	// Lock to create onlyregular user by asignin the "user" role
	user.Role = "user"
	// Set the roleID dynamically so it can be para metrized in the future
	err = user.SetRole()

	// lock tokens to be empty at user creation
	user.Tokens = *new([]string)
	user.BlockedTokens = *new([]string)

	// check that the username/email is not already being used
	isUnique, err := user.CheckUnique()
	if isUnique {
		return c.Status(fiber.StatusLocked).JSON(fiber.Map{"status": "error", "message": "Couldn't create user, user with the same username or email already exists", "data": err})
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
// @param	uid path string true "User ID or username"
// @Success      200  {object}  interface{}
// @Failure      401  {object}  interface{}
// @Failure      422  {object}  interface{}
// @Failure      500  {object}  interface{}
// @Router       /users/{uid} [patch]
func UpdateUser(c *fiber.Ctx) error {
	//TODO: add more fields to update (maybe via model.User)

	// user update input
	var uui models.User
	if err := c.BodyParser(&uui); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status":  "error",
			"message": "Unpocessable Entity, Review your input",
			"data":    err,
		})
	}

	// Identity of the user to modify
	identity := c.Params("uid")
	// get the data of the user we want to modify
	var user models.User
	user.Fill(identity, true, true, false)
	// Token of the editor's user
	token := c.Locals("user").(*jwt.Token)
	editorUID := fmt.Sprintf("%v", token.Claims.(jwt.MapClaims)["uid"])

	// Get the editor's data
	var editorUser models.User
	editorUser.Fill(editorUID, true, false, false)
	var editorRole models.Role
	editorRole.Fill(editorUser.RoleID.Hex(), true, false)

	// check if editor is authorised to do the operation
	if user.ID.Hex() != editorUID {
		// Parametrized permissons
		if !editorRole.Permissons.UsersAdmin {
			/*
				——————————No Perms?——————————————————
				⠀⣞⢽⢪⢣⢣⢣⢫⡺⡵⣝⡮⣗⢷⢽⢽⢽⣮⡷⡽⣜⣜⢮⢺⣜⢷⢽⢝⡽⣝
				⠸⡸⠜⠕⠕⠁⢁⢇⢏⢽⢺⣪⡳⡝⣎⣏⢯⢞⡿⣟⣷⣳⢯⡷⣽⢽⢯⣳⣫⠇
				⠀⠀⢀⢀⢄⢬⢪⡪⡎⣆⡈⠚⠜⠕⠇⠗⠝⢕⢯⢫⣞⣯⣿⣻⡽⣏⢗⣗⠏⠀
				⠀⠪⡪⡪⣪⢪⢺⢸⢢⢓⢆⢤⢀⠀⠀⠀⠀⠈⢊⢞⡾⣿⡯⣏⢮⠷⠁⠀⠀
				⠀⠀⠀⠈⠊⠆⡃⠕⢕⢇⢇⢇⢇⢇⢏⢎⢎⢆⢄⠀⢑⣽⣿⢝⠲⠉⠀⠀⠀⠀
				⠀⠀⠀⠀⠀⡿⠂⠠⠀⡇⢇⠕⢈⣀⠀⠁⠡⠣⡣⡫⣂⣿⠯⢪⠰⠂⠀⠀⠀⠀
				⠀⠀⠀⠀⡦⡙⡂⢀⢤⢣⠣⡈⣾⡃⠠⠄⠀⡄⢱⣌⣶⢏⢊⠂⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⠀⢝⡲⣜⡮⡏⢎⢌⢂⠙⠢⠐⢀⢘⢵⣽⣿⡿⠁⠁⠀⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⠀⠨⣺⡺⡕⡕⡱⡑⡆⡕⡅⡕⡜⡼⢽⡻⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⠀⣼⣳⣫⣾⣵⣗⡵⡱⡡⢣⢑⢕⢜⢕⡝⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⣴⣿⣾⣿⣿⣿⡿⡽⡑⢌⠪⡢⡣⣣⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⡟⡾⣿⢿⢿⢵⣽⣾⣼⣘⢸⢸⣞⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
				⠀⠀⠀⠀⠁⠇⠡⠩⡫⢿⣝⡻⡮⣒⢽⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
				——————————————————————————————————————
			*/
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to update the user, your role cant update other users.",
				"data":    nil,
			})
		}
	}

	// Authenticated & autorized

	// ————————Validate fields to be updated———————————

	// Special field role requires aditional autorization
	if uui.Role != "" {
		/*
		 The editor can only modify users of a higher level (so, less permissons) than itself
		 and can only set the user to a role with a higher level than the editor
		*/

		// get the original role
		err := user.SetRole()
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusBadRequest).JSON(
				stdMsg.ErrorDefault("The specified role does not exist", nil),
			)
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				stdMsg.ErrorDefault("Unknown error while retieving the specified role", nil),
			)
		}
		var role models.Role
		role.Fill(user.RoleID.Hex(), true, false)

		// check if editor can edit this user's role by role level
		if role.Level <= editorRole.Level {
			return c.Status(fiber.StatusForbidden).JSON(
				stdMsg.ErrorDefault("You cant set the the role of this user", nil),
			)
		}
		// check if editor can edit this user's role by role permissons
		if editorRole.Permissons.RolesAdmin && editorRole.Permissons.UsersAdmin {
			user.Role = uui.Role
		} else {
			return c.Status(fiber.StatusForbidden).JSON(
				stdMsg.ErrorDefault("failed to update the user, your role cant modify users roles.", nil),
			)
		}

		// check if the specified role exists and set it
		err = user.SetRole()
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusBadRequest).JSON(
				stdMsg.ErrorDefault("The specified role does not exist", nil),
			)
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				stdMsg.ErrorDefault("Unknown error while retieving the specified role", nil),
			)
		}
		role.Fill(user.RoleID.Hex(), true, false)
		// check if editor can set the specified role
		if role.Level <= editorRole.Level {
			return c.Status(fiber.StatusForbidden).JSON(
				stdMsg.ErrorDefault("You cant set the the role of this user", nil),
			)
		}

	}

	// user and email validation
	if uui.Username != "" {
		user.Username = uui.Username
	}
	if uui.Email != "" || uui.Email != user.Email {
		user.Email = uui.Email
	}
	unique, err := user.CheckUnique()
	if err != nil {
		return c.Status(500).JSON(stdMsg.ErrorDefault(
			"An error ocured while checking if username and email are unique",
			err,
		))
	}
	if !unique {
		return c.Status(fiber.StatusConflict).JSON(stdMsg.ErrorDefault(
			"Specified username or email is already being used",
			err,
		))
	}

	// password validation
	if uui.Password != "" {
		hash, err := utils.Hash(uui.Password)
		if err != nil {
			return c.Status(500).JSON(stdMsg.ErrorDefault("Couldn't hash password", err))

		}
		user.Password = hash
	}

	if uui.Name != "" {
		user.Name = uui.Name
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

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
		"data":    user.Private(),
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
	// password input
	var pIn PasswordInput
	if err := c.BodyParser(&pIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	// Identity of the user to modify
	identity := c.Params("uid")
	// get the data of the user we want to modify
	var user models.User
	user.Fill(identity, true, true, false)
	// Token of the editor's user
	token := c.Locals("user").(*jwt.Token)
	editorUID := fmt.Sprintf("%v", token.Claims.(jwt.MapClaims)["uid"])

	if user.ID.Hex() != editorUID {
		// Get the editor's data
		var editorUser models.User
		editorUser.Fill(editorUID, true, false, false)
		var editorRole models.Role
		editorRole.Fill(editorUser.RoleID.Hex(), true, false)

		// Parametrized permissons
		if !editorRole.Permissons.UsersAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to delete the user, your role cant delete other users.",
				"data":    nil,
			})
		}
	}

	// Authenticated & autorized

	if !utils.CheckHash(pIn.Password, user.Password) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "error", "message": "Not valid user", "data": nil})
	}

	filter := bson.M{"_id": user.ID.Hex()}
	deleted, err := models.UsersCollection.DeleteOne(services.Mongo.Context, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Something went wron while deleting the user", "data": nil})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": string(deleted.DeletedCount) + " user successfully deleted",
		"data":    user,
	})
}
