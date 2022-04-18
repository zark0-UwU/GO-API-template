package loaders

import (
	"ToDoList/src/models"
	"ToDoList/src/services"
	"context"
)

func LoadMongo() *context.CancelFunc {
	cancelCtx := services.Mongo.Init()

	// init all the collections
	models.User{}.CreateSingletonDBAndCollection()
	return cancelCtx
}
