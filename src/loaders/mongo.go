package loaders

import (
	"context"
	"template/src/models"
	"template/src/services"
)

func LoadMongo() *context.CancelFunc {
	cancelCtx := services.Mongo.Init()

	// init all the collections
	models.User{}.CreateSingletonDBAndCollection()
	return cancelCtx
}
