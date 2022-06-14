package main

import (
	"web_app/configs"

	"web_app/router"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	configs.ConnectDB()

	router.Root(e)

	e.Logger.Fatal(e.Start(":6000"))
}
