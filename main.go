package main

import (
	"github.com/italosilva18/prod-mpm/configs"
	"github.com/italosilva18/prod-mpm/router"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	configs.ConnectDB()

	router.Root(e)

	e.Logger.Fatal(e.Start(":6000"))
}
