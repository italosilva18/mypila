package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"m2m-backend/helpers"
	"m2m-backend/services"
)

// LookupCNPJ busca dados de uma empresa pelo CNPJ
func LookupCNPJ(c *fiber.Ctx) error {
	cnpj := c.Params("cnpj")
	if cnpj == "" {
		return helpers.SendValidationError(c, "cnpj", "CNPJ e obrigatorio")
	}

	// Validate CNPJ format
	if !services.ValidateCNPJ(cnpj) {
		return helpers.SendValidationError(c, "cnpj", "CNPJ invalido: deve conter 14 digitos")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := services.LookupCNPJ(ctx, cnpj)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(data)
}
