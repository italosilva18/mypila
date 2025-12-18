package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m2m-backend/helpers"
	"m2m-backend/models"
	"m2m-backend/services"
)

var companyService = services.NewCompanyService()

// GetCompanies returns all companies owned by the authenticated user
func GetCompanies(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Call service layer
	companies, err := companyService.GetCompaniesByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch companies"})
	}

	return c.JSON(companies)
}

// CreateCompany creates a new company for the authenticated user
func CreateCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Call service layer
	company, validationErrors, err := companyService.CreateCompany(userID, req)
	if err != nil {
		if err == services.ErrInvalidInput && validationErrors != nil {
			return helpers.SendValidationErrors(c, validationErrors)
		}
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar empresa"})
	}

	return c.Status(201).JSON(company)
}

// UpdateCompany updates an existing company (ownership validated)
func UpdateCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Call service layer
	company, validationErrors, err := companyService.UpdateCompany(companyID, userID, req)
	if err != nil {
		if err == services.ErrInvalidInput && validationErrors != nil {
			return helpers.SendValidationErrors(c, validationErrors)
		}
		if err == services.ErrCompanyNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "Company not found"})
		}
		if err == services.ErrUnauthorized {
			return c.Status(403).JSON(fiber.Map{
				"error":   "Forbidden: you do not have permission to access this company",
				"details": "This company belongs to another user",
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar empresa"})
	}

	return c.JSON(company)
}

// DeleteCompany deletes a company and all related data in cascade (ownership validated)
func DeleteCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Call service layer
	err = companyService.DeleteCompany(companyID, userID)
	if err != nil {
		if err == services.ErrCompanyNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "Company not found"})
		}
		if err == services.ErrUnauthorized {
			return c.Status(403).JSON(fiber.Map{
				"error":   "Forbidden: you do not have permission to delete this company",
				"details": "This company belongs to another user",
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete company"})
	}

	return c.JSON(fiber.Map{"message": "Company deleted successfully"})
}
