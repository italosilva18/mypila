package router

import (
	"github.com/italosilva18/prod-mpm/controllers"
	"github.com/labstack/echo/v4"
)

func Root(e *echo.Echo) {

	e.POST("/add", controllers.Users.Add)
	e.GET("/show", controllers.Users.Show)
	//e.PUT("/update", controllers.Users.Update)
	//e.DELETE("/delete", controllers.Users.Delete)
	//e.GET("/pont", controllers.Get.Novo)

}
