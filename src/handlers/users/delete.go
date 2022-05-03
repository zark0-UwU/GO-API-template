package users

import (
	"GO-API-template/src/models"
	"GO-API-template/src/utils"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
)

type PasswordInput struct {
	Password string `json:"password"`
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
	deleted, err := models.UsersCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Something went wron while deleting the user", "data": nil})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": string(deleted.DeletedCount) + " user successfully deleted",
		"data":    user,
	})
}
