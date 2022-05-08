package src

import (
	"GO-API-template/src/config"
	"GO-API-template/src/loaders"

	"github.com/lightstep/otel-launcher-go/launcher"
)

// @title           GO API template
// @version         1.0
// @description     This is a production ready sample API server with authentication.
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @license.name    GNU GPLv3
// @license.url     https://www.gnu.org/licenses/gpl-3.0.html

// @host      kaomoji.zark0.dev prod
// @host      localhost:5000 devel
// @BasePath  /v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Description for what is this security definition being used
// @tokenUrl                    https://kaomoji.zark0.dev/v1/auth/login
func Start() {
	//Open Telemetry setup
	ls := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("Go-API-Template"),
		launcher.WithAccessToken(config.Config("OTEL-LIGHT-STEP-KEY")),
	)
	defer ls.Shutdown()
	// END Open Telemetry setup

	serverPort := config.Config("PORT")
	// Try connecting to the database (loads and init)
	cancelCtx := loaders.LoadMongo()
	defer (*cancelCtx)()

	app := loaders.LoadFiber()
	app.Listen(":" + serverPort)
}
