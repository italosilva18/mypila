package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var Users *userRouting

type userRouting struct {
}

func (t *userRouting) Show(c echo.Context) error {
	return c.JSON(http.StatusOK, "ponteiro Sw")
}
func (t *userRouting) Add(c echo.Context) error {
	return c.JSON(http.StatusOK, "ponteiro add")
}

func (t *userRouting) Delete(c echo.Context) error {
	return c.JSON(http.StatusOK, "ponteiro del")
}
func (t *userRouting) Update(c echo.Context) error {
	return c.JSON(http.StatusOK, "ponteiro upd")
}

var Get *getRouting

type getRouting struct {
}

func (t *getRouting) Novo(c echo.Context) error {
	return c.String(http.StatusOK, "ola")
}
