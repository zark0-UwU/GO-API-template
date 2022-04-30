package users

import (
	"GO-API-template/src/config"
	"GO-API-template/src/models"
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type rangeUser struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Amount  int         `json:"amount"`
	Offset  int         `json:"offset"`
	Limmit  int         `json:"limmit"`
	Next    string      `json:"next"`
	Users   interface{} `json:"users"`
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
