package src

import (
	"template/src/config"
	"template/src/loaders"
)

func Start() {
	serverPort := config.Config("PORT")
	// Try connecting to the database (loads and init)
	cancelCtx := loaders.LoadMongo()
	defer (*cancelCtx)()

	app := loaders.LoadFiber()
	app.Listen(":" + serverPort)
}
