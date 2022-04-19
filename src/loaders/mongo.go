package loaders

import (
	"context"

	"GO-API-template/src/models"
	"GO-API-template/src/services"
)

func LoadMongo() *context.CancelFunc {
	cancelCtx := services.Mongo.Init()

	// init all the collections
	models.User{}.CreateSingletonDBAndCollection()
	return cancelCtx
}
